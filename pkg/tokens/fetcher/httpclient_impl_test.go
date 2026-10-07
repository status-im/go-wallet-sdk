package fetcher

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClient_ReadResponse_Plain(t *testing.T) {
	response := &http.Response{
		Body: io.NopCloser(strings.NewReader("plain text response")),
	}

	client := NewHTTPClient(DefaultConfig())
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

	client := NewHTTPClient(DefaultConfig())
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

	client := NewHTTPClient(DefaultConfig())
	data, err := client.readResponse(response)
	assert.Error(t, err)
	assert.Empty(t, data)
}

func TestHTTPClient_ReadResponse_ReadError(t *testing.T) {
	response := &http.Response{
		Body: &failingReader{},
	}

	client := NewHTTPClient(DefaultConfig())
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

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestHTTPClient_UsesConfiguredTransport(t *testing.T) {
	var requested []string
	config := DefaultConfig()
	config.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requested = append(requested, req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Etag": []string{`"v1"`}},
			Body:       io.NopCloser(strings.NewReader(`{"tokens":[]}`)),
		}, nil
	})

	data, etag, err := NewHTTPClient(config).DoGetRequestWithEtag(t.Context(), "https://example.com/list.json", "")
	assert.NoError(t, err)
	assert.Equal(t, `{"tokens":[]}`, string(data))
	assert.Equal(t, `"v1"`, etag)
	assert.Equal(t, []string{"https://example.com/list.json"}, requested)
}
