package autofetcher_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
	mock_autofetcher "github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher/mock"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	mock_fetcher "github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher/mock"
	mock_parsers "github.com/status-im/go-wallet-sdk/pkg/tokens/parsers/mock"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func createValidRemoteListConfig(ctrl *gomock.Controller) autofetcher.ConfigRemoteListOfTokenLists {
	mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)
	return autofetcher.ConfigRemoteListOfTokenLists{
		Config: autofetcher.Config{
			LastUpdate:               time.Now().Add(-2 * time.Hour),
			AutoRefreshInterval:      time.Hour,
			AutoRefreshCheckInterval: time.Minute,
		},
		RemoteListOfTokenListsFetchDetails: types.ListDetails{
			ID:        "test-remote-list",
			SourceURL: "https://example.com/token-lists.json",
			Schema:    "https://example.com/schema.json",
		},
		RemoteListOfTokenListsParser: mockParser,
	}
}

func createValidTokenListsConfig() autofetcher.ConfigTokenLists {
	return autofetcher.ConfigTokenLists{
		Config: autofetcher.Config{
			LastUpdate:               time.Now().Add(-2 * time.Hour),
			AutoRefreshInterval:      time.Hour,
			AutoRefreshCheckInterval: time.Minute,
		},
		TokenLists: []types.ListDetails{
			{
				ID:        "uniswap",
				SourceURL: "https://tokens.uniswap.org",
				Schema:    "https://uniswap.org/tokenlist.schema.json",
			},
			{
				ID:        "compound",
				SourceURL: "https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json",
			},
		},
	}
}

func TestNewAutofetcherFromRemoteListOfTokenLists_ValidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidRemoteListConfig(ctrl)
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)
	require.NotNil(t, af)
}

func TestNewAutofetcherFromRemoteListOfTokenLists_InvalidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidRemoteListConfig(ctrl)
	config.RemoteListOfTokenListsParser = nil
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, mockFetcher, mockContentStore)

	assert.Error(t, err)
	assert.Nil(t, af)
	assert.Equal(t, autofetcher.ErrRemoteListOfTokenListsParserNotProvided, err)
}

func TestNewAutofetcherFromRemoteListOfTokenLists_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidRemoteListConfig(ctrl)
	config.RemoteListOfTokenListsFetchDetails.SourceURL = "not-a-url"
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, mockFetcher, mockContentStore)

	assert.Error(t, err)
	assert.Nil(t, af)
}

func TestNewAutofetcherFromRemoteListOfTokenLists_InvalidInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidRemoteListConfig(ctrl)
	config.AutoRefreshCheckInterval = 2 * time.Hour
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, mockFetcher, mockContentStore)

	assert.Error(t, err)
	assert.Nil(t, af)
	assert.Equal(t, autofetcher.ErrAutoRefreshCheckIntervalGreaterThanInterval, err)
}

func TestNewAutofetcherFromRemoteListOfTokenLists_NilContentStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidRemoteListConfig(ctrl)
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)

	af, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, mockFetcher, nil)

	assert.Error(t, err)
	assert.Nil(t, af)
	assert.Equal(t, autofetcher.ErrContentStoreNotProvided, err)
}

func TestNewAutofetcherFromRemoteListOfTokenLists_NilFetcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidRemoteListConfig(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, nil, mockContentStore)

	assert.Error(t, err)
	assert.Nil(t, af)
	assert.Equal(t, autofetcher.ErrFetcherNotProvided, err)
}

func TestNewAutofetcherFromTokenLists_ValidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)
	require.NotNil(t, af)
}

func TestNewAutofetcherFromTokenLists_EmptyTokenLists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	config.TokenLists = []types.ListDetails{}
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)

	assert.Error(t, err)
	assert.Nil(t, af)
	assert.Equal(t, autofetcher.ErrTokenListsNotProvided, err)
}

func TestNewAutofetcherFromTokenLists_InvalidTokenList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	config.TokenLists[0].SourceURL = "not-a-url"
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)

	assert.Error(t, err)
	assert.Nil(t, af)
}

func TestNewAutofetcherFromTokenLists_NilFetcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, nil, mockContentStore)

	assert.Error(t, err)
	assert.Nil(t, af)
	assert.Equal(t, autofetcher.ErrFetcherNotProvided, err)
}

func TestNewAutofetcherFromTokenLists_NilContentStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, nil)

	assert.Error(t, err)
	assert.Nil(t, af)
	assert.Equal(t, autofetcher.ErrContentStoreNotProvided, err)
}

func TestAutofetcher_StartStop_Basic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	config.AutoRefreshInterval = 100 * time.Millisecond
	config.AutoRefreshCheckInterval = 10 * time.Millisecond

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	mockContentStore.EXPECT().GetEtag(gomock.Any()).Return("", nil).AnyTimes()
	mockContentStore.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFetcher.EXPECT().FetchConcurrent(gomock.Any(), gomock.Any()).Return([]fetcher.FetchedData{}, nil).AnyTimes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	refreshCh := af.Start(ctx)
	require.NotNil(t, refreshCh)

	// Wait for at least one refresh cycle
	select {
	case err := <-refreshCh:
		assert.NoError(t, err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected refresh within timeout")
	}

	af.Stop()

	// Verify channel is closed
	select {
	case _, ok := <-refreshCh:
		assert.False(t, ok, "Channel should be closed after Stop")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel should be closed quickly after Stop")
	}
}

func TestAutofetcher_Start_MultipleCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	ctx := context.Background()

	refreshCh1 := af.Start(ctx)
	require.NotNil(t, refreshCh1)

	refreshCh2 := af.Start(ctx)
	assert.Equal(t, refreshCh1, refreshCh2)

	af.Stop()
}

func TestAutofetcher_Stop_MultipleCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	ctx := context.Background()

	af.Start(ctx)
	af.Stop()

	af.Stop()
	af.Stop()
}

func TestAutofetcher_Start_AfterStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	ctx := context.Background()

	refreshCh1 := af.Start(ctx)
	af.Stop()

	refreshCh2 := af.Start(ctx)
	assert.NotEqual(t, refreshCh1, refreshCh2, "Should get new channel after restart")

	af.Stop()
}

func TestAutofetcher_RefreshLogic_WithTokenLists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	config.LastUpdate = time.Now().Add(-2 * time.Hour)

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	// Set up mock expectations for the refresh cycle
	mockContentStore.EXPECT().GetEtag(gomock.Any()).Return("", nil).AnyTimes()
	mockContentStore.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockFetcher.EXPECT().FetchConcurrent(gomock.Any(), gomock.Any()).Return([]fetcher.FetchedData{
		{
			FetchDetails: fetcher.FetchDetails{
				ListDetails: config.TokenLists[0],
			},
			JsonData: []byte(`{"tokens": []}`),
			Fetched:  time.Now(),
		},
		{
			FetchDetails: fetcher.FetchDetails{
				ListDetails: config.TokenLists[1],
			},
			JsonData: []byte(`{"tokens": []}`),
			Fetched:  time.Now(),
		},
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	refreshCh := af.Start(ctx)

	select {
	case err := <-refreshCh:
		assert.NoError(t, err)
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("Expected refresh within timeout")
	}

	af.Stop()
}

func TestAutofetcher_RefreshLogic_FetchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	config.LastUpdate = time.Now().Add(-2 * time.Hour)

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	// Set up mock expectations for the refresh cycle
	mockContentStore.EXPECT().GetEtag(gomock.Any()).Return("", nil).AnyTimes()

	mockFetcher.EXPECT().FetchConcurrent(gomock.Any(), gomock.Any()).Return([]fetcher.FetchedData{}, errors.New("fetch failed"))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	refreshCh := af.Start(ctx)

	select {
	case err := <-refreshCh:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fetch failed")
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("Expected refresh error within timeout")
	}

	af.Stop()
}

func TestAutofetcher_RefreshLogic_NoRefreshNeeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	config.LastUpdate = time.Now() // Recent update, no refresh needed
	config.AutoRefreshInterval = time.Hour
	config.AutoRefreshCheckInterval = 10 * time.Millisecond

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	refreshCh := af.Start(ctx)

	// Should not receive any refresh events
	select {
	case <-refreshCh:
		t.Fatal("Should not refresh when lastUpdate is recent")
	case <-ctx.Done():
		// Expected - no refresh should occur
	}

	af.Stop()
}

func TestAutofetcher_RefreshLogic_WithRemoteList_FetchError_EmptyStoredData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidRemoteListConfig(ctrl)
	config.LastUpdate = time.Now().Add(-2 * time.Hour)

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	// Setup: fetch fails, stored data is empty
	mockContentStore.EXPECT().GetEtag(gomock.Any()).Return("", nil).AnyTimes()
	mockFetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(fetcher.FetchedData{}, errors.New("fetch failed"))
	mockContentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	refreshCh := af.Start(ctx)

	// Wait for refresh error
	select {
	case err := <-refreshCh:
		assert.Error(t, err)
		assert.Equal(t, autofetcher.ErrStoredListOfTokenListsIsEmpty, err)
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("Expected refresh error within timeout")
	}

	af.Stop()
}

func TestAutofetcher_ConcurrentStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidTokenListsConfig()
	config.LastUpdate = time.Now() // Set recent update to avoid refresh
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	ctx := context.Background()

	// Concurrent start/stop operations
	var wg sync.WaitGroup
	channels := make([]chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			channels[i] = af.Start(ctx)
		}(i)
	}

	wg.Wait()

	// All channels should be the same
	for i := 1; i < len(channels); i++ {
		assert.Equal(t, channels[0], channels[i])
	}

	// Concurrent stops
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			af.Stop()
		}()
	}

	wg.Wait()
}

func TestAutofetcher_RefreshLogic_WithRemoteList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := createValidRemoteListConfig(ctrl)
	config.LastUpdate = time.Now().Add(-2 * time.Hour)

	mockParser := config.RemoteListOfTokenListsParser.(*mock_parsers.MockListOfTokenListsParser)
	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	mockContentStore := mock_autofetcher.NewMockContentStore(ctrl)

	af, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, mockFetcher, mockContentStore)
	require.NoError(t, err)

	listOfTokenListsData := []byte(`{"tokenLists": [{"id": "uniswap", "sourceUrl": "https://tokens.uniswap.org"}]}`)

	// Set up mock expectations for the refresh cycle
	mockContentStore.EXPECT().GetEtag(gomock.Any()).Return("", nil).AnyTimes()
	mockContentStore.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockParser.EXPECT().Parse(listOfTokenListsData).Return(&types.ListOfTokenLists{
		TokenLists: []types.ListDetails{
			{
				ID:        "uniswap",
				SourceURL: "https://tokens.uniswap.org",
			},
		},
	}, nil)

	mockFetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(fetcher.FetchedData{
		JsonData: listOfTokenListsData,
		Fetched:  time.Now(),
	}, nil)

	mockFetcher.EXPECT().FetchConcurrent(gomock.Any(), gomock.Any()).Return([]fetcher.FetchedData{
		{
			FetchDetails: fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "uniswap",
					SourceURL: "https://tokens.uniswap.org",
				},
			},
			JsonData: []byte(`{"tokens": []}`),
			Fetched:  time.Now(),
		},
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	refreshCh := af.Start(ctx)

	// Wait for refresh
	select {
	case err := <-refreshCh:
		assert.NoError(t, err)
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("Expected refresh within timeout")
	}

	af.Stop()
}
