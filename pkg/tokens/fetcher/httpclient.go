package fetcher

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultRequestTimeout  = 5 * time.Second
	defaultIdleConnTimeout = 90 * time.Second
	defaultMaxIdleConns    = 10
)

// Config represents the configuration for the HTTP client
type Config struct {
	Timeout            time.Duration
	IdleConnTimeout    time.Duration
	MaxIdleConns       int
	DisableCompression bool
}

// DefaultConfig returns the default configuration for the HTTP client
func DefaultConfig() Config {
	return Config{
		Timeout:            defaultRequestTimeout,
		IdleConnTimeout:    defaultIdleConnTimeout,
		MaxIdleConns:       defaultMaxIdleConns,
		DisableCompression: false,
	}
}

// HTTPClient represents an HTTP client with configurable options
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client with the provided configuration
func NewHTTPClient(config Config) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:       config.MaxIdleConns,
				IdleConnTimeout:    config.IdleConnTimeout,
				DisableCompression: config.DisableCompression,
			},
		},
	}
}

// DoGetRequestWithEtag performs a GET request with the given URL and parameters
// If etag is not empty, it will add an If-None-Match header to the request
// If the server responds with a 304 status code (`http.StatusNotModified`), it will return an empty body and the same etag
func (c *HTTPClient) DoGetRequestWithEtag(ctx context.Context, url string, etag string) (data []byte, newETag string, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}

	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	var resp *http.Response
	resp, err = c.client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to fetch %s: %w", url, err)
		return
	}
	defer func() {
		err1 := resp.Body.Close()
		if err == nil && err1 != nil {
			err = fmt.Errorf("failed to close response body: %w", err1)
		}
	}()

	if resp.StatusCode == http.StatusNotModified {
		newETag = etag
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, url)
		return
	}

	newETag = resp.Header.Get("ETag")

	data, err = c.readResponse(resp)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %w", err)
	}

	return
}

func (c *HTTPClient) readResponse(resp *http.Response) (data []byte, err error) {
	var reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		var gzipErr error
		reader, gzipErr = gzip.NewReader(resp.Body)
		if gzipErr != nil {
			return nil, gzipErr
		}
		defer func() {
			err1 := reader.Close()
			if err == nil && err1 != nil {
				err = fmt.Errorf("failed to close response body: %w", err1)
			}
		}()
	}
	data, err = io.ReadAll(reader)
	return
}
