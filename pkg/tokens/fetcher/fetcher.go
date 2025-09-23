package fetcher

import (
	"context"
	"sync"
	"time"
)

type fetcher struct {
	httpClient *HTTPClient
}

// New creates a new fetcher with the provided configuration
func New(config Config) *fetcher {
	return &fetcher{
		httpClient: NewHTTPClient(config),
	}
}

// Fetch fetches a single resource from the URL specified in the details.
func (t *fetcher) Fetch(ctx context.Context, details FetchDetails) (FetchedData, error) {
	var fetchedData = FetchedData{
		FetchDetails: details,
	}

	err := details.Validate()
	if err != nil {
		fetchedData.Error = err
		return fetchedData, err
	}

	body, newEtag, err := t.httpClient.DoGetRequestWithEtag(ctx, details.SourceURL, details.Etag)
	if err != nil {
		fetchedData.Error = err
		return fetchedData, err
	}

	if details.Etag != "" && newEtag == details.Etag {
		return fetchedData, nil
	}

	if details.Schema != "" {
		err = validateJsonAgainstSchema(string(body), details.Schema)
		if err != nil {
			fetchedData.Error = err
			return fetchedData, err
		}
	}

	fetchedData.Etag = newEtag
	fetchedData.Fetched = time.Now()
	fetchedData.JsonData = body

	return fetchedData, nil
}

// FetchConcurrent fetches multiple resources concurrently from the URLs specified in the details.
func (t *fetcher) FetchConcurrent(ctx context.Context, details []FetchDetails) ([]FetchedData, error) {
	if len(details) == 0 {
		return []FetchedData{}, nil
	}

	var wg sync.WaitGroup
	ch := make(chan FetchedData, len(details))

	for _, d := range details {
		wg.Add(1)
		go func(ctx context.Context, details FetchDetails) {
			defer wg.Done()

			fetchedData, _ := t.Fetch(ctx, details)
			select {
			case ch <- fetchedData:
			case <-ctx.Done():
			}
		}(ctx, d)
	}

	wg.Wait()
	close(ch)

	var fetchedData []FetchedData
	for fetchedList := range ch {
		fetchedData = append(fetchedData, fetchedList)
	}

	ch = nil

	return fetchedData, nil
}
