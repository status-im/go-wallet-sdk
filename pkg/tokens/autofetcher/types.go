package autofetcher

//go:generate mockgen -destination=mock/autofetcher.go . AutoFetcher,ContentStore

import (
	"context"
	"time"
)

// AutoFetcher is the interface for fetching provided token lists or remote list of token lists and all the token lists
// from it, depending on the config and storing them in the content store.
// Implementations are thread-safe and can be used concurrently from multiple goroutines.
type AutoFetcher interface {
	// Start starts the background autofetcher process.
	// Returns a channel that receives errors from refresh operations, if no error is returned, the refresh was successful.
	// Can be called multiple times safely - subsequent calls return the same channel.
	Start(ctx context.Context) (refreshCh chan error)

	// Stop stops the background autofetcher process.
	// Blocks until the background goroutine has finished.
	// Can be called multiple times safely.
	Stop()
}

type Content struct {
	SourceURL string
	Etag      string
	Data      []byte
	Fetched   time.Time
}

// ContentStore interface for storing and retrieving fetched content.
// Implementations MUST be thread-safe for concurrent access.
type ContentStore interface {
	// GetEtag retrieves the Etag for a given ID.
	GetEtag(id string) (string, error)
	// Get retrieves the content for a given ID.
	Get(id string) (Content, error)
	// Set stores the content for a given ID.
	Set(id string, content Content) error
	// GetAll retrieves all content.
	GetAll() (map[string]Content, error)
}
