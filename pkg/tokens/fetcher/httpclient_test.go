package fetcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient(t *testing.T) {
	config := fetcher.DefaultConfig()
	client := fetcher.NewHTTPClient(config)
	assert.NotNil(t, client)
}

func TestNewHTTPClient_CustomConfig(t *testing.T) {
	config := fetcher.Config{
		Timeout:            10 * time.Second,
		IdleConnTimeout:    60 * time.Second,
		MaxIdleConns:       20,
		DisableCompression: true,
	}
	client := fetcher.NewHTTPClient(config)
	assert.NotNil(t, client)
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

	client := fetcher.NewHTTPClient(fetcher.DefaultConfig())
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
	client := fetcher.NewHTTPClient(fetcher.DefaultConfig())
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

	client := fetcher.NewHTTPClient(fetcher.DefaultConfig())
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

	config := fetcher.DefaultConfig()
	config.Timeout = 10 * time.Millisecond

	client := fetcher.NewHTTPClient(config)

	ctx := context.Background()

	data, etag, err := client.DoGetRequestWithEtag(ctx, server.URL, "")
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Empty(t, etag)
	assert.True(t, err != nil)
}
