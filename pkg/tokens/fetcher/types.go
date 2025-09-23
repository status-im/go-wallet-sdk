package fetcher

//go:generate mockgen -destination=mock/fetcher.go . Fetcher

import (
	"context"
	_ "embed"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

//go:embed list_of_token_lists_schema.json
var ListOfTokenListsSchema string

type Fetcher interface {
	// FetchConcurrent fetches multiple resources concurrently from the URLs specified in the details.
	FetchConcurrent(ctx context.Context, details []FetchDetails) ([]FetchedData, error)

	// Fetch fetches a single resource from the URL specified in the details.
	Fetch(ctx context.Context, details FetchDetails) (FetchedData, error)
}

// FetchDetails represents a token list in the remote list of token lists.
type FetchDetails struct {
	types.ListDetails
	Etag string
}

// FetchedTokenList represents a fetched token list.
type FetchedData struct {
	FetchDetails
	Fetched  time.Time
	JsonData []byte
	Error    error
}
