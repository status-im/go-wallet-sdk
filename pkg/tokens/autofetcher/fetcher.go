package autofetcher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

var (
	ErrStoredListOfTokenListsIsEmpty = errors.New("stored list of token lists is empty")
	ErrFetcherNotProvided            = errors.New("fetcher not provided")
	ErrContentStoreNotProvided       = errors.New("content store not provided")
)

// autofetcher handles the background fetch of token lists (thread-safe for concurrent access).
type autofetcher struct {
	mu        sync.Mutex
	cancel    context.CancelFunc
	refreshCh chan error

	contentStore ContentStore
	fetcher      fetcher.Fetcher
	wg           sync.WaitGroup

	remoteListOfTokenListsFetchDetails types.ListDetails
	remoteListOfTokenListsParser       parsers.ListOfTokenListsParser

	tokenLists []types.ListDetails

	lastUpdate               time.Time
	autoRefreshInterval      time.Duration
	autoRefreshCheckInterval time.Duration // must be <= AutoRefreshInterval
}

// NewAutofetcherFromRemoteListOfTokenLists creates a new autofetcher from the remote list of token lists.
// It fetches the remote list of token lists defined by the config and fetches all the token lists from the remote list
// of token lists and stores them in the content store.
func NewAutofetcherFromRemoteListOfTokenLists(config ConfigRemoteListOfTokenLists, fetcher fetcher.Fetcher,
	contentStore ContentStore) (*autofetcher, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if fetcher == nil {
		return nil, ErrFetcherNotProvided
	}

	if contentStore == nil {
		return nil, ErrContentStoreNotProvided
	}

	return &autofetcher{
		fetcher:      fetcher,
		contentStore: contentStore,

		remoteListOfTokenListsFetchDetails: config.RemoteListOfTokenListsFetchDetails,
		remoteListOfTokenListsParser:       config.RemoteListOfTokenListsParser,

		lastUpdate:               config.LastUpdate,
		autoRefreshInterval:      config.AutoRefreshInterval,
		autoRefreshCheckInterval: config.AutoRefreshCheckInterval,
	}, nil
}

// NewAutofetcherFromTokenLists creates a new autofetcher from the token lists.
// It fetches all the token lists defined by the config and stores them in the content store.
func NewAutofetcherFromTokenLists(config ConfigTokenLists, fetcher fetcher.Fetcher,
	contentStore ContentStore) (*autofetcher, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if fetcher == nil {
		return nil, ErrFetcherNotProvided
	}

	if contentStore == nil {
		return nil, ErrContentStoreNotProvided
	}

	return &autofetcher{
		fetcher:      fetcher,
		contentStore: contentStore,

		tokenLists: config.TokenLists,

		lastUpdate:               config.LastUpdate,
		autoRefreshInterval:      config.AutoRefreshInterval,
		autoRefreshCheckInterval: config.AutoRefreshCheckInterval,
	}, nil
}

// Start starts the background autofetcher process.
func (a *autofetcher) Start(ctx context.Context) (refreshCh chan error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		return a.refreshCh
	}

	a.refreshCh = make(chan error)

	childCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	a.wg.Add(1)
	go a.run(childCtx, a.refreshCh)

	return a.refreshCh
}

// Stop stops the background autofetcher process.
func (a *autofetcher) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel == nil {
		return
	}

	a.cancel()
	a.wg.Wait()
	a.cancel = nil
	a.refreshCh = nil
}

func (a *autofetcher) run(ctx context.Context, refreshCh chan error) {
	defer a.wg.Done()
	defer close(refreshCh)

	ticker := time.NewTicker(a.autoRefreshCheckInterval)
	defer ticker.Stop()

	// Check immediately on start
	select {
	case <-ctx.Done():
		return
	default:
		a.checkAndRefresh(ctx, refreshCh)
	}

	// Check every autoRefreshCheckInterval
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.checkAndRefresh(ctx, refreshCh)
		}
	}
}

func (a *autofetcher) fetchAndStoreListOfTokenLists(ctx context.Context) (fetcher.FetchedData, error) {
	// use the last etag from the content store, error is not important here
	etag, _ := a.contentStore.GetEtag(a.remoteListOfTokenListsFetchDetails.ID)

	fetchedData, err := a.fetcher.Fetch(ctx, fetcher.FetchDetails{
		ListDetails: a.remoteListOfTokenListsFetchDetails,
		Etag:        etag,
	})
	if err != nil || len(fetchedData.JsonData) == 0 {
		storedListOfTokenLists, err := a.contentStore.Get(a.remoteListOfTokenListsFetchDetails.ID)
		if err != nil {
			return fetchedData, err
		}

		if len(storedListOfTokenLists.Data) == 0 {
			return fetchedData, ErrStoredListOfTokenListsIsEmpty
		}

		fetchedData.JsonData = storedListOfTokenLists.Data
		fetchedData.Etag = storedListOfTokenLists.Etag
		fetchedData.Fetched = storedListOfTokenLists.Fetched
	}

	return fetchedData, nil
}

func (a *autofetcher) fetchAndStoreTokenLists(ctx context.Context, details []types.ListDetails) error {
	fetchDetails := make([]fetcher.FetchDetails, len(details))
	for i, detail := range details {
		// use the last etag from the content store, error is not important here
		etag, _ := a.contentStore.GetEtag(detail.ID)

		fetchDetails[i] = fetcher.FetchDetails{
			ListDetails: types.ListDetails{
				ID:        detail.ID,
				SourceURL: detail.SourceURL,
				Schema:    detail.Schema,
			},
			Etag: etag,
		}
	}

	fetchedTokenLists, err := a.fetcher.FetchConcurrent(ctx, fetchDetails)
	if err != nil {
		return err
	}

	for _, fetched := range fetchedTokenLists {
		if fetched.Error != nil || len(fetched.JsonData) == 0 {
			// ignore storing the token list if it failed to fetch or if the data is empty (304 Not Modified - the same etag)
			continue
		}

		err = a.contentStore.Set(fetched.ID, Content{
			SourceURL: fetched.SourceURL,
			Data:      fetched.JsonData,
			Etag:      fetched.Etag,
			Fetched:   fetched.Fetched,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *autofetcher) checkAndRefresh(ctx context.Context, refreshCh chan error) {
	if time.Since(a.lastUpdate) < a.autoRefreshInterval {
		return
	}

	var (
		err        error
		tokenLists = a.tokenLists
	)

	if len(tokenLists) == 0 {
		fetchedData, err := a.fetchAndStoreListOfTokenLists(ctx)
		if err != nil {
			select {
			case refreshCh <- err:
			case <-ctx.Done():
			}
			return
		}

		listOfTokenLists, err := a.remoteListOfTokenListsParser.Parse(fetchedData.JsonData)
		if err != nil {
			select {
			case refreshCh <- err:
			case <-ctx.Done():
			}
			return
		}

		err = a.contentStore.Set(a.remoteListOfTokenListsFetchDetails.ID, Content{
			SourceURL: a.remoteListOfTokenListsFetchDetails.SourceURL,
			Data:      fetchedData.JsonData,
			Etag:      fetchedData.Etag,
			Fetched:   fetchedData.Fetched,
		})
		if err != nil {
			select {
			case refreshCh <- err:
			case <-ctx.Done():
			}
			return
		}

		tokenLists = listOfTokenLists.TokenLists
	}

	err = a.fetchAndStoreTokenLists(ctx, tokenLists)
	if err == nil {
		a.lastUpdate = time.Now()
	}

	// Send result, but check if context is done first
	select {
	case refreshCh <- err:
	case <-ctx.Done():
	}
}
