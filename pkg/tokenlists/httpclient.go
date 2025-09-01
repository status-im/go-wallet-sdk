package tokenlists

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	defaultRequestTimeout  = 5 * time.Second
	defaultIdleConnTimeout = 90 * time.Second
	defaultMaxIdleConns    = 10
)

// HTTPClient represents an HTTP client with configurable options
type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: defaultRequestTimeout,
			Transport: &http.Transport{
				MaxIdleConns:       defaultMaxIdleConns,
				IdleConnTimeout:    defaultIdleConnTimeout,
				DisableCompression: false,
			},
		},
	}
}

// DoGetRequestWithEtag performs a GET request with the given URL and parameters
// If etag is not empty, it will add an If-None-Match header to the request
// If the server responds with a 304 status code (`http.StatusNotModified`), it will return an empty body and the same etag
func (c *HTTPClient) DoGetRequestWithEtag(ctx context.Context, url string, etag string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err.Error())
		}
	}()

	if resp.StatusCode == http.StatusNotModified {
		return nil, etag, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, url)
	}

	newETag := resp.Header.Get("ETag")

	body, err := c.readResponse(resp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return body, newETag, nil
}

func (c *HTTPClient) readResponse(resp *http.Response) ([]byte, error) {
	var reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		var gzipErr error
		reader, gzipErr = gzip.NewReader(resp.Body)
		if gzipErr != nil {
			return nil, gzipErr
		}
		defer func() {
			if err := reader.Close(); err != nil {
				log.Println(err.Error())
			}
		}()
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return body, nil
}
