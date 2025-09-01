package tokenlists

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewTokenListsFetcher(t *testing.T) {
	config := &Config{
		logger: zap.NewNop(),
	}

	fetcher := newTokenListsFetcher(config)
	assert.NotNil(t, fetcher)
	assert.NotNil(t, fetcher.httpClient)
}

func TestTokenListsFetcher_FetchAndStore(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:                    zap.NewNop(),
		ContentStore:              NewDefaultContentStore(),
		RemoteListOfTokenListsURL: server.URL + listOfTokenListsURL,
	}

	fetcher := newTokenListsFetcher(config)

	count, err := fetcher.fetchAndStore(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, count, "Should fetch and store two token lists")

	content, err := config.ContentStore.Get("status")
	assert.NoError(t, err)
	assert.Equal(t, statusTokenListJsonResponse, string(content.Data))

	content, err = config.ContentStore.Get("uniswap")
	assert.NoError(t, err)
	assert.Equal(t, uniswapTokenListJsonResponse, string(content.Data))
}

func TestTokenListsFetcher_FetchAndStore_NoRemoteURL(t *testing.T) {
	config := &Config{
		logger:       zap.NewNop(),
		ContentStore: NewDefaultContentStore(),
	}

	fetcher := newTokenListsFetcher(config)

	count, err := fetcher.fetchAndStore(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTokenListsFetcher_FetchAndStore_EmptyTokenLists(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:                    zap.NewNop(),
		ContentStore:              NewDefaultContentStore(),
		RemoteListOfTokenListsURL: server.URL + emptyTokenListsURL,
	}

	fetcher := newTokenListsFetcher(config)

	count, err := fetcher.fetchAndStore(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTokenListsFetcher_FetchAndStore_ContextCancellation(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:                    zap.NewNop(),
		ContentStore:              NewDefaultContentStore(),
		RemoteListOfTokenListsURL: server.URL + delayedResponseURL,
	}

	fetcher := newTokenListsFetcher(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count, err := fetcher.fetchAndStore(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTokenListsFetcher_FetchAndStore_PartialFailures(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:                    zap.NewNop(),
		ContentStore:              NewDefaultContentStore(),
		RemoteListOfTokenListsURL: server.URL + listOfTokenListsSomeWrongUrlsURL,
	}

	fetcher := newTokenListsFetcher(config)

	count, err := fetcher.fetchAndStore(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	content, err := config.ContentStore.Get("status")
	assert.NoError(t, err)
	assert.NotNil(t, content.Data)
	assert.Equal(t, statusTokenListJsonResponse, string(content.Data))

	content, err = config.ContentStore.Get("invalid-list")
	assert.Error(t, err)
	assert.Nil(t, content.Data)
}
