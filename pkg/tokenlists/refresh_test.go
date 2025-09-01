package tokenlists

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRefreshWorker(t *testing.T) {
	config := &Config{
		logger: zap.NewNop(),
	}

	worker := newRefreshWorker(config)
	assert.NotNil(t, worker)
	assert.NotNil(t, worker.fetcher)
}

func TestRefreshWorker_StartStop(t *testing.T) {
	config := &Config{
		AutoRefreshCheckInterval:      100 * time.Millisecond,
		AutoRefreshInterval:           200 * time.Millisecond,
		logger:                        zap.NewNop(),
		PrivacyGuard:                  &defaultPrivacyGuard{},
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  &defaultContentStore{},
	}

	worker := newRefreshWorker(config)
	ctx := context.Background()

	// start
	refreshCh := worker.start(ctx)
	assert.NotNil(t, refreshCh)
	assert.True(t, worker.running.Load())

	// start again
	refreshCh2 := worker.start(ctx)
	assert.Nil(t, refreshCh2)

	// stop
	worker.stop()
	assert.False(t, worker.running.Load())

	// stop again
	worker.stop()
}

func TestRefreshWorker_Run(t *testing.T) {
	config := &Config{
		AutoRefreshCheckInterval:      50 * time.Millisecond,
		AutoRefreshInterval:           100 * time.Millisecond,
		logger:                        zap.NewNop(),
		PrivacyGuard:                  &defaultPrivacyGuard{},
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  &defaultContentStore{},
	}

	worker := newRefreshWorker(config)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	refreshCh := worker.start(ctx)
	require.NotNil(t, refreshCh)

	// Wait for refresh signal or context cancellation
	select {
	case <-refreshCh:
	case <-ctx.Done():
		// Expected behavior, any of those two
	}

	worker.stop()
}

func TestRefreshWorker_PrivacyModeOn(t *testing.T) {
	config := &Config{
		AutoRefreshCheckInterval:      50 * time.Millisecond,
		AutoRefreshInterval:           100 * time.Millisecond,
		logger:                        zap.NewNop(),
		PrivacyGuard:                  NewDefaultPrivacyGuard(true),
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  &defaultContentStore{},
	}

	worker := newRefreshWorker(config)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	refreshCh := worker.start(ctx)
	require.NotNil(t, refreshCh)

	// With privacy mode on, no refresh should be triggered
	select {
	case <-refreshCh:
		t.Fatal("Refresh should not be triggered when privacy mode is on")
	case <-ctx.Done():
		// Expected behavior
	}

	worker.stop()
}

func TestRefreshWorker_CheckAndRefresh_PrivacyModeOff(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		RemoteListOfTokenListsURL:     server.URL + listOfTokenListsURL,
		AutoRefreshCheckInterval:      100 * time.Millisecond,
		AutoRefreshInterval:           200 * time.Millisecond,
		logger:                        zap.NewNop(),
		PrivacyGuard:                  NewDefaultPrivacyGuard(false),
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  NewDefaultContentStore(),
	}

	allContent, err := config.ContentStore.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allContent, 0)

	worker := newRefreshWorker(config)

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	refreshCh := make(chan struct{}, 1)

	worker.checkAndRefresh(ctx, refreshCh)

	select {
	case <-refreshCh:
		// check if the content store has the data
		allContent, err := config.ContentStore.GetAll()
		assert.NoError(t, err)
		assert.Len(t, allContent, 3) // list of token lists, status token list, uniswap token list

		assert.Contains(t, allContent, "status")
		assert.Contains(t, allContent, "uniswap")
		assert.Equal(t, statusTokenListJsonResponse, string(allContent["status"].Data))
		assert.Equal(t, uniswapTokenListJsonResponse, string(allContent["uniswap"].Data))
		return

	case <-ctx.Done():
		t.Fatal("context done")
	}
}

func TestRefreshWorker_CheckAndRefresh_PrivacyModeOn(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		RemoteListOfTokenListsURL:     server.URL + listOfTokenListsURL,
		AutoRefreshCheckInterval:      100 * time.Millisecond,
		AutoRefreshInterval:           200 * time.Millisecond,
		logger:                        zap.NewNop(),
		PrivacyGuard:                  NewDefaultPrivacyGuard(true),
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  NewDefaultContentStore(),
	}

	allContent, err := config.ContentStore.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allContent, 0)

	worker := newRefreshWorker(config)

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	refreshCh := make(chan struct{}, 1)

	worker.checkAndRefresh(ctx, refreshCh)

	select {
	case <-refreshCh:
		t.Fatal("refreshCh received")
	case <-ctx.Done():
		// Expected behavior
	}

	allContent, err = config.ContentStore.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allContent, 0)
}
