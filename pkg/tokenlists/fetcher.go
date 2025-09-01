package tokenlists

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type tokenListsFetcher struct {
	config     *Config
	httpClient *HTTPClient
}

func newTokenListsFetcher(config *Config) *tokenListsFetcher {
	return &tokenListsFetcher{
		config:     config,
		httpClient: NewHTTPClient(),
	}
}

func (t *tokenListsFetcher) fetchAndStore(ctx context.Context) (int, error) {
	remoteListOfTokenLists, err := t.resolveListOfTokenLists(ctx)
	if err != nil || len(remoteListOfTokenLists.TokenLists) == 0 {
		t.config.logger.Error("cannot resolve list of token lists", zap.Error(err))
		return 0, nil
	}

	var wg sync.WaitGroup
	tokenChannel := make(chan fetchedTokenList, len(remoteListOfTokenLists.TokenLists))

	for _, tokenList := range remoteListOfTokenLists.TokenLists {
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()

			dbEtag, err := t.config.ContentStore.GetEtag(tokenList.ID)
			if err != nil {
				// don't return, but fetch using an empty etag
				t.config.logger.Error("cannot get cached etag for token list", zap.String("list-id", tokenList.ID))
			}
			err = t.fetchTokenList(ctx, tokenList, dbEtag, tokenChannel)
			if err != nil {
				t.config.logger.Error("failed to fetch token list", zap.Error(err), zap.String("list-id", tokenList.ID))
			}
		}(ctx)
	}

	wg.Wait()
	close(tokenChannel)

	var successfullyFetchedListsCount int
	for fetchedList := range tokenChannel {
		if err := t.config.ContentStore.Set(fetchedList.ID, Content{
			SourceURL: fetchedList.SourceURL,
			Etag:      fetchedList.Etag,
			Data:      fetchedList.JsonData,
			Fetched:   fetchedList.Fetched,
		}); err != nil {
			t.config.logger.Error("failed to store token list", zap.Error(err))
		} else {
			successfullyFetchedListsCount++
		}
	}

	tokenChannel = nil

	return successfullyFetchedListsCount, nil
}
