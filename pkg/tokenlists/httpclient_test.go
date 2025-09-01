package tokenlists

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.Equal(t, defaultRequestTimeout, client.client.Timeout)
}

func TestHTTPClient_DoGetRequestWithEtag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if etag := r.Header.Get("If-None-Match"); etag != "" {
			if etag == "test-etag" {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.Header().Set("ETag", "new-etag")
		_, err := w.Write([]byte("test response"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()

	data, etag, err := client.DoGetRequestWithEtag(ctx, server.URL, "")
	assert.NoError(t, err)
	assert.Equal(t, "test response", string(data))
	assert.Equal(t, "new-etag", etag)

	data, etag, err = client.DoGetRequestWithEtag(ctx, server.URL, "test-etag")
	assert.NoError(t, err)
	assert.Empty(t, data)
	assert.Equal(t, "test-etag", etag)

	data, etag, err = client.DoGetRequestWithEtag(ctx, server.URL, "non-matching-etag")
	assert.NoError(t, err)
	assert.Equal(t, "test response", string(data))
	assert.Equal(t, "new-etag", etag)
}

func TestHTTPClient_DoGetRequestWithEtag_Error(t *testing.T) {
	client := NewHTTPClient()
	ctx := context.Background()

	data, etag, err := client.DoGetRequestWithEtag(ctx, "invalid-url", "")
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, etag)
}

func TestHTTPClient_DoGetRequestWithEtag_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("not found"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()

	data, etag, err := client.DoGetRequestWithEtag(ctx, server.URL, "")
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, etag)
	assert.Contains(t, err.Error(), "unexpected status code 404")
}

func TestHTTPClient_DoGetRequestWithEtag_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, err := w.Write([]byte("delayed response"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Millisecond,
		},
	}

	ctx := context.Background()

	data, etag, err := client.DoGetRequestWithEtag(ctx, server.URL, "")
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, etag)
	assert.True(t, err != nil)
}

func TestHTTPClient_ReadResponse_Plain(t *testing.T) {
	response := &http.Response{
		Body: io.NopCloser(strings.NewReader("plain text response")),
	}

	client := NewHTTPClient()
	data, err := client.readResponse(response)
	assert.NoError(t, err)
	assert.Equal(t, "plain text response", string(data))
}

func TestHTTPClient_ReadResponse_Gzip(t *testing.T) {
	var buf strings.Builder
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte("gzipped content"))
	assert.NoError(t, err)
	err = gw.Close()
	assert.NoError(t, err)

	response := &http.Response{
		Body: io.NopCloser(strings.NewReader(buf.String())),
		Header: http.Header{
			"Content-Encoding": []string{"gzip"},
		},
	}

	client := NewHTTPClient()
	data, err := client.readResponse(response)
	assert.NoError(t, err)
	assert.Equal(t, "gzipped content", string(data))
}

func TestHTTPClient_ReadResponse_GzipError(t *testing.T) {
	response := &http.Response{
		Body: io.NopCloser(strings.NewReader("invalid gzip content")),
		Header: http.Header{
			"Content-Encoding": []string{"gzip"},
		},
	}

	client := NewHTTPClient()
	data, err := client.readResponse(response)
	assert.Error(t, err)
	assert.Empty(t, data)
}

func TestHTTPClient_ReadResponse_ReadError(t *testing.T) {
	response := &http.Response{
		Body: &failingReader{},
	}

	client := NewHTTPClient()
	data, err := client.readResponse(response)
	assert.Error(t, err)
	assert.Empty(t, data)
}

type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}

func (f *failingReader) Close() error {
	return nil
}
