package manager_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/common"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
	mock_autofetcher "github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher/mock"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	mock_fetcher "github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher/mock"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/manager"
	mock_manager "github.com/status-im/go-wallet-sdk/pkg/tokens/manager/mock"
	mock_parsers "github.com/status-im/go-wallet-sdk/pkg/tokens/parsers/mock"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

var (
	testChains = []uint64{common.EthereumMainnet, common.BSCMainnet}
)

func createTestConfig() *manager.Config {
	return &manager.Config{
		MainListID: "main-list",
		InitialLists: map[string][]byte{
			"main-list": []byte(`{"name": "Main List", "tokens": []}`),
			"list2":     []byte(`{"name": "List 2", "tokens": []}`),
		},
		Chains: testChains,
	}
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	t.Run("valid config", func(t *testing.T) {
		config := createTestConfig()
		m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
		require.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("invalid config", func(t *testing.T) {
		config := &manager.Config{} // empty config
		m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
		assert.Error(t, err)
		assert.Nil(t, m)
	})

	t.Run("nil content store", func(t *testing.T) {
		config := createTestConfig()
		m, err := manager.New(config, mockFetcher, nil, customTokenStore)
		assert.ErrorIs(t, err, manager.ErrContentStoreNotProvided)
		assert.Nil(t, m)
	})

	t.Run("with auto fetcher config", func(t *testing.T) {
		mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)
		config := createTestConfig()
		config.AutoFetcherConfig = &autofetcher.ConfigRemoteListOfTokenLists{
			Config: autofetcher.Config{
				AutoRefreshInterval:      time.Hour,
				AutoRefreshCheckInterval: time.Minute,
			},
			RemoteListOfTokenListsFetchDetails: types.ListDetails{
				ID:        "remote-list",
				SourceURL: "https://example.com/remote.json",
				Schema:    "standard",
			},
			RemoteListOfTokenListsParser: mockParser,
		}
		m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
		require.NoError(t, err)
		assert.NotNil(t, m)
	})
}

func TestManager_StartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("start without auto refresh", func(t *testing.T) {
		err := m.Start(ctx, false, nil)
		require.NoError(t, err)

		// Should have tokens after start
		tokens := m.UniqueTokens()
		assert.NotEmpty(t, tokens)

		err = m.Stop()
		require.NoError(t, err)
	})

	t.Run("start with auto refresh but no notify channel", func(t *testing.T) {
		err := m.Start(ctx, true, nil)
		assert.ErrorIs(t, err, manager.ErrAutoRefreshEnabledButNotifyChannelNotProvided)
	})

	t.Run("start with notify channel but no auto fetcher", func(t *testing.T) {
		notifyCh := make(chan struct{}, 1)
		err := m.Start(ctx, false, notifyCh)
		assert.ErrorIs(t, err, manager.ErrAutoFetcherNotProvided)
	})

	t.Run("double start", func(t *testing.T) {
		err := m.Start(ctx, false, nil)
		require.NoError(t, err)

		err = m.Start(ctx, false, nil)
		require.NoError(t, err)

		err = m.Stop()
		require.NoError(t, err)
	})

	t.Run("stop without start", func(t *testing.T) {
		err := m.Stop()
		require.NoError(t, err)
	})
}

func TestManager_TokenOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()
	err = m.Start(ctx, false, nil)
	require.NoError(t, err)
	defer func() {
		err := m.Stop()
		require.NoError(t, err)
	}()

	t.Run("unique tokens", func(t *testing.T) {
		tokens := m.UniqueTokens()
		assert.NotEmpty(t, tokens)

		// Should include native tokens for both chains
		foundEth := false
		foundBsc := false
		for _, token := range tokens {
			if token.ChainID == common.EthereumMainnet && token.Symbol == "ETH" {
				foundEth = true
			}
			if token.ChainID == common.BSCMainnet && token.Symbol == "BNB" {
				foundBsc = true
			}
		}
		assert.True(t, foundEth, "Should include ETH native token")
		assert.True(t, foundBsc, "Should include BNB native token")
	})

	t.Run("get token by chain address", func(t *testing.T) {
		tokens := m.UniqueTokens()
		require.NotEmpty(t, tokens)

		firstToken := tokens[0]
		token, exists := m.GetTokenByChainAddress(firstToken.ChainID, firstToken.Address)
		assert.True(t, exists)
		assert.Equal(t, firstToken, token)

		token, exists = m.GetTokenByChainAddress(999, gethcommon.HexToAddress("0x0000000000000000000000000000000000000000"))
		assert.False(t, exists)
		assert.Nil(t, token)
	})

	t.Run("get tokens by chain", func(t *testing.T) {
		ethTokens := m.GetTokensByChain(common.EthereumMainnet)
		assert.NotEmpty(t, ethTokens)
		for _, token := range ethTokens {
			assert.Equal(t, common.EthereumMainnet, token.ChainID)
		}

		bscTokens := m.GetTokensByChain(common.BSCMainnet)
		assert.NotEmpty(t, bscTokens)
		for _, token := range bscTokens {
			assert.Equal(t, common.BSCMainnet, token.ChainID)
		}

		unknownTokens := m.GetTokensByChain(999)
		assert.Empty(t, unknownTokens)
	})

	t.Run("get tokens by keys", func(t *testing.T) {
		allTokens := m.UniqueTokens()
		require.NotEmpty(t, allTokens)

		keys := make([]string, 0)
		expectedTokens := make(map[string]*types.Token)
		for i := 0; i < len(allTokens); i++ {
			token := allTokens[i]
			key := types.TokenKey(token.ChainID, token.Address)
			keys = append(keys, key)
			expectedTokens[key] = token
		}

		tokens, err := m.GetTokensByKeys(keys)
		assert.NoError(t, err)
		assert.Len(t, tokens, len(keys))

		// Verify all returned tokens match expected
		for _, token := range tokens {
			key := types.TokenKey(token.ChainID, token.Address)
			expectedToken, exists := expectedTokens[key]
			assert.True(t, exists)
			assert.Equal(t, expectedToken, token)
		}
	})

	t.Run("get tokens by keys - empty keys", func(t *testing.T) {
		tokens, err := m.GetTokensByKeys([]string{})
		assert.NoError(t, err)
		assert.Empty(t, tokens)
	})

	t.Run("get tokens by keys - non-existent keys", func(t *testing.T) {
		keys := []string{"1-0x0000000000000000000000000000000000000001", "999-0x0000000000000000000000000000000000000002"}
		tokens, err := m.GetTokensByKeys(keys)
		assert.NoError(t, err)
		assert.Empty(t, tokens)
	})

	t.Run("get tokens by keys - mixed valid and invalid keys", func(t *testing.T) {
		allTokens := m.UniqueTokens()
		require.NotEmpty(t, allTokens)

		// Mix of valid and invalid keys
		validToken := allTokens[0]
		validKey := types.TokenKey(validToken.ChainID, validToken.Address)
		invalidKeys := []string{"999-0x0000000000000000000000000000000000000001", "888-0x0000000000000000000000000000000000000002"}

		keys := append([]string{validKey}, invalidKeys...)

		tokens, err := m.GetTokensByKeys(keys)
		assert.NoError(t, err)
		assert.Len(t, tokens, 1)
		assert.Equal(t, validToken, tokens[0])
	})

	t.Run("get tokens by keys - duplicate keys", func(t *testing.T) {
		allTokens := m.UniqueTokens()
		require.NotEmpty(t, allTokens)

		token := allTokens[0]
		key := types.TokenKey(token.ChainID, token.Address)

		// Use the same key multiple times
		keys := []string{key, key, key}

		tokens, err := m.GetTokensByKeys(keys)
		assert.NoError(t, err)
		// Should return the token multiple times since we requested it multiple times
		assert.Len(t, tokens, 3)
		for _, returnedToken := range tokens {
			assert.Equal(t, token, returnedToken)
		}
	})

	t.Run("get tokens by keys - case insensitive", func(t *testing.T) {
		allTokens := m.UniqueTokens()
		require.NotEmpty(t, allTokens)

		token := allTokens[0]
		// Generate key with different cases
		lowerKey := types.TokenKey(token.ChainID, token.Address)
		upperKey := strings.ToUpper(lowerKey)
		mixedKey := strings.ToUpper(lowerKey[:5]) + lowerKey[5:]

		// All variations should return the same token
		keys := []string{lowerKey, upperKey, mixedKey}

		tokens, err := m.GetTokensByKeys(keys)
		assert.NoError(t, err)
		assert.Len(t, tokens, 3)
		for _, returnedToken := range tokens {
			assert.Equal(t, token, returnedToken)
		}
	})

	t.Run("token lists", func(t *testing.T) {
		lists := m.TokenLists()
		assert.NotEmpty(t, lists)

		foundNative := false
		for _, list := range lists {
			if list.Name == "Native tokens" {
				foundNative = true
				break
			}
		}
		assert.True(t, foundNative, "Should include native token list")
	})

	t.Run("get token list by id", func(t *testing.T) {
		list, exists := m.TokenList("native")
		assert.True(t, exists)
		assert.Equal(t, "Native tokens", list.Name)

		list, exists = m.TokenList("non-existent")
		assert.False(t, exists)
		assert.Nil(t, list)
	})
}

func TestManager_CustomTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	config := createTestConfig()

	// Add custom token
	customToken := &types.Token{
		CrossChainID: "custom-token",
		ChainID:      common.EthereumMainnet,
		Address:      gethcommon.HexToAddress("0x1111111111111111111111111111111111111111"),
		Decimals:     18,
		Name:         "Custom Token",
		Symbol:       "CUSTOM",
	}

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{customToken}, nil).AnyTimes()

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()
	err = m.Start(ctx, false, nil)
	require.NoError(t, err)
	defer func() {
		err := m.Stop()
		require.NoError(t, err)
	}()

	t.Run("custom tokens included", func(t *testing.T) {
		tokens := m.UniqueTokens()

		foundCustom := false
		for _, token := range tokens {
			if token.Symbol == "CUSTOM" {
				foundCustom = true
				assert.Equal(t, customToken.Name, token.Name)
				assert.Equal(t, customToken.Address, token.Address)
				break
			}
		}
		assert.True(t, foundCustom, "Should include custom token")
	})

	t.Run("custom token list exists", func(t *testing.T) {
		list, exists := m.TokenList("custom")
		assert.True(t, exists)
		assert.Equal(t, "Custom tokens", list.Name)
		assert.Len(t, list.Tokens, 1)
		assert.Equal(t, customToken.Symbol, list.Tokens[0].Symbol)
	})
}

func TestManager_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	config := createTestConfig()

	t.Run("content store error", func(t *testing.T) {
		contentStore := mock_autofetcher.NewMockContentStore(ctrl)
		customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)

		contentStore.EXPECT().GetAll().Return(nil, errors.New("content store error")).AnyTimes()
		contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
		customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

		m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
		require.NoError(t, err)

		ctx := context.Background()
		err = m.Start(ctx, false, nil)
		require.Error(t, err) // GetAll error is propagated
		assert.Contains(t, err.Error(), "content store error")
	})

	t.Run("custom token store error", func(t *testing.T) {
		contentStore := mock_autofetcher.NewMockContentStore(ctrl)
		customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)

		contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
		contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
		customTokenStore.EXPECT().GetAll().Return(nil, errors.New("custom token store error")).AnyTimes()

		m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
		require.NoError(t, err)

		ctx := context.Background()
		err = m.Start(ctx, false, nil)
		require.Error(t, err) // fail because custom token store returns error
		assert.Contains(t, err.Error(), "custom token store error")
	})
}

func TestManager_Concurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()
	err = m.Start(ctx, false, nil)
	require.NoError(t, err)
	defer func() {
		err := m.Stop()
		require.NoError(t, err)
	}()

	t.Run("concurrent reads", func(t *testing.T) {
		const numGoroutines = 10
		const numIterations = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			// concurrent access to manager
			go func() {
				defer wg.Done()
				for j := 0; j < numIterations; j++ {
					_ = m.UniqueTokens()
					_ = m.TokenLists()
					_, _ = m.GetTokenByChainAddress(common.EthereumMainnet, gethcommon.HexToAddress("0x0000000000000000000000000000000000000000"))
					_ = m.GetTokensByChain(common.EthereumMainnet)
					_, _ = m.TokenList("native")
				}
			}()
		}

		wg.Wait()
	})
}

func TestManager_AutoRefreshOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()
	err = m.Start(ctx, false, nil)
	require.NoError(t, err)
	defer func() {
		err := m.Stop()
		require.NoError(t, err)
	}()

	t.Run("auto refresh operations without auto fetcher", func(t *testing.T) {
		err := m.EnableAutoRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)

		err = m.DisableAutoRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)

	})
}

func TestManager_AutoRefreshWithAutoFetcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)
	config := createTestConfig()

	// Set up mock expectations - these may be called by autofetcher background processes
	mockFetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(fetcher.FetchedData{
		JsonData: []byte("{}"),
	}, nil).AnyTimes()
	mockFetcher.EXPECT().FetchConcurrent(gomock.Any(), gomock.Any()).Return([]fetcher.FetchedData{}, nil).AnyTimes()
	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().GetEtag(gomock.Any()).Return("", nil).AnyTimes()
	contentStore.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()
	mockParser.EXPECT().Parse(gomock.Any()).Return(&types.ListOfTokenLists{}, nil).AnyTimes()

	config.AutoFetcherConfig = &autofetcher.ConfigRemoteListOfTokenLists{
		Config: autofetcher.Config{
			LastUpdate:               time.Now(), // Recent update to prevent immediate refresh
			AutoRefreshInterval:      time.Hour,
			AutoRefreshCheckInterval: time.Hour,
		},
		RemoteListOfTokenListsFetchDetails: types.ListDetails{
			ID:        "remote-list",
			SourceURL: "https://example.com/remote.json",
			Schema:    "standard",
		},
		RemoteListOfTokenListsParser: mockParser,
	}

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()
	notifyCh := make(chan struct{}, 1)

	// Start without auto refresh initially to avoid triggering autofetcher background processes
	err = m.Start(ctx, false, notifyCh)
	require.NoError(t, err)
	defer func() {
		err := m.Stop()
		require.NoError(t, err)
	}()

	t.Run("enable auto refresh when disabled", func(t *testing.T) {
		err := m.EnableAutoRefresh(ctx)
		assert.NoError(t, err)
	})

	t.Run("enable auto refresh when already enabled", func(t *testing.T) {
		err := m.EnableAutoRefresh(ctx)
		assert.NoError(t, err)
	})

	t.Run("disable auto refresh", func(t *testing.T) {
		err := m.DisableAutoRefresh(ctx)
		assert.NoError(t, err)
	})

	t.Run("disable auto refresh when already disabled", func(t *testing.T) {
		err := m.DisableAutoRefresh(ctx)
		assert.NoError(t, err)
	})

	t.Run("enable auto refresh after disabling", func(t *testing.T) {
		err := m.EnableAutoRefresh(ctx)
		assert.NoError(t, err)
	})
}

func TestManager_AutoRefreshWithoutNotifyChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	config.AutoFetcherConfig = &autofetcher.ConfigRemoteListOfTokenLists{
		Config: autofetcher.Config{
			AutoRefreshInterval:      time.Hour,
			AutoRefreshCheckInterval: time.Hour,
		},
		RemoteListOfTokenListsFetchDetails: types.ListDetails{
			ID:        "remote-list",
			SourceURL: "https://example.com/remote.json",
			Schema:    "standard",
		},
		RemoteListOfTokenListsParser: mockParser,
	}

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()

	// start without notify channel
	err = m.Start(ctx, false, nil)
	require.NoError(t, err)
	defer func() {
		err := m.Stop()
		require.NoError(t, err)
	}()

	t.Run("auto refresh operations without notify channel", func(t *testing.T) {
		err := m.EnableAutoRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)

		err = m.DisableAutoRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)
	})
}

func TestManager_AutoRefreshToggling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().GetEtag(gomock.Any()).Return("", nil).AnyTimes()
	contentStore.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()
	mockFetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(fetcher.FetchedData{JsonData: []byte("{}")}, nil).AnyTimes()
	mockFetcher.EXPECT().FetchConcurrent(gomock.Any(), gomock.Any()).Return([]fetcher.FetchedData{}, nil).AnyTimes()
	mockParser.EXPECT().Parse(gomock.Any()).Return(&types.ListOfTokenLists{}, nil).AnyTimes()

	config.AutoFetcherConfig = &autofetcher.ConfigRemoteListOfTokenLists{
		Config: autofetcher.Config{
			AutoRefreshInterval:      time.Hour,
			AutoRefreshCheckInterval: time.Hour,
		},
		RemoteListOfTokenListsFetchDetails: types.ListDetails{
			ID:        "remote-list",
			SourceURL: "https://example.com/remote.json",
			Schema:    "standard",
		},
		RemoteListOfTokenListsParser: mockParser,
	}

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()
	notifyCh := make(chan struct{}, 10)

	err = m.Start(ctx, false, notifyCh)
	require.NoError(t, err)
	defer func() {
		err := m.Stop()
		require.NoError(t, err)
	}()

	t.Run("toggle auto refresh multiple times", func(t *testing.T) {
		err := m.EnableAutoRefresh(ctx)
		assert.NoError(t, err)

		err = m.EnableAutoRefresh(ctx)
		assert.NoError(t, err)

		err = m.DisableAutoRefresh(ctx)
		assert.NoError(t, err)

		err = m.DisableAutoRefresh(ctx)
		assert.NoError(t, err)

		err = m.EnableAutoRefresh(ctx)
		assert.NoError(t, err)

		err = m.DisableAutoRefresh(ctx)
		assert.NoError(t, err)
	})
}

func TestManager_AutoRefreshBeforeStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	config.AutoFetcherConfig = &autofetcher.ConfigRemoteListOfTokenLists{
		Config: autofetcher.Config{
			AutoRefreshInterval:      time.Hour,
			AutoRefreshCheckInterval: time.Hour,
		},
		RemoteListOfTokenListsFetchDetails: types.ListDetails{
			ID:        "remote-list",
			SourceURL: "https://example.com/remote.json",
			Schema:    "standard",
		},
		RemoteListOfTokenListsParser: mockParser,
	}

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("auto refresh operations before start", func(t *testing.T) {
		err := m.EnableAutoRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)

		err = m.DisableAutoRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)

		err = m.TriggerRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)
	})
}

func TestManager_AutoRefreshNotificationChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)
	config := createTestConfig()

	mockFetcher.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(fetcher.FetchedData{
		JsonData: []byte("{}"),
	}, nil).AnyTimes()
	mockFetcher.EXPECT().FetchConcurrent(gomock.Any(), gomock.Any()).Return([]fetcher.FetchedData{}, nil).AnyTimes()
	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().GetEtag(gomock.Any()).Return("", nil).AnyTimes()
	contentStore.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()
	mockParser.EXPECT().Parse(gomock.Any()).Return(&types.ListOfTokenLists{}, nil).AnyTimes()

	config.AutoFetcherConfig = &autofetcher.ConfigRemoteListOfTokenLists{
		Config: autofetcher.Config{
			AutoRefreshInterval:      500 * time.Millisecond,
			AutoRefreshCheckInterval: 200 * time.Millisecond,
		},
		RemoteListOfTokenListsFetchDetails: types.ListDetails{
			ID:        "remote-list",
			SourceURL: "https://prod.market.status.im/static/lists.json",
		},
		RemoteListOfTokenListsParser: mockParser,
	}

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	ctx := context.Background()
	notifyCh := make(chan struct{}, 10)

	// Start with auto refresh disabled
	err = m.Start(ctx, false, notifyCh)
	require.NoError(t, err)
	defer func() {
		err := m.Stop()
		require.NoError(t, err)
	}()

	t.Run("no notifications when auto refresh disabled", func(t *testing.T) {
		for len(notifyCh) > 0 {
			<-notifyCh
		}

		time.Sleep(1000 * time.Millisecond)

		select {
		case <-notifyCh:
			t.Error("Should not receive notification when auto refresh is disabled")
		default:
			// Expected - no notification received
		}
	})

	t.Run("notifications received when auto refresh enabled", func(t *testing.T) {
		for len(notifyCh) > 0 {
			<-notifyCh
		}

		err := m.EnableAutoRefresh(ctx)
		require.NoError(t, err)

		timeout := time.After(1 * time.Second)
		select {
		case <-notifyCh:
			// Expected - notification received after enabling auto refresh
			t.Log("✓ Received notification after enabling auto refresh")
		case <-timeout:
			t.Error("Should have received notification after enabling auto refresh within 1 second")
		}
	})

	t.Run("no more notifications after disabling auto refresh", func(t *testing.T) {
		err := m.DisableAutoRefresh(ctx)
		require.NoError(t, err)

		for len(notifyCh) > 0 {
			<-notifyCh
		}

		time.Sleep(1000 * time.Millisecond)

		select {
		case <-notifyCh:
			t.Error("Should not receive notification after disabling auto refresh")
		default:
			// Expected - no notification received
			t.Log("✓ No notifications received after disabling auto refresh")
		}
	})
}

func TestManager_TriggerRefreshErrorConditions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	t.Run("trigger refresh without auto fetcher", func(t *testing.T) {
		m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
		require.NoError(t, err)

		ctx := context.Background()
		err = m.Start(ctx, false, nil)
		require.NoError(t, err)
		defer func() {
			err := m.Stop()
			require.NoError(t, err)
		}()

		err = m.TriggerRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)
	})

	t.Run("trigger refresh without notify channel", func(t *testing.T) {
		mockParser := mock_parsers.NewMockListOfTokenListsParser(ctrl)
		config.AutoFetcherConfig = &autofetcher.ConfigRemoteListOfTokenLists{
			Config: autofetcher.Config{
				AutoRefreshInterval:      time.Hour,
				AutoRefreshCheckInterval: time.Hour,
			},
			RemoteListOfTokenListsFetchDetails: types.ListDetails{
				ID:        "remote-list",
				SourceURL: "https://example.com/remote.json",
				Schema:    "standard",
			},
			RemoteListOfTokenListsParser: mockParser,
		}

		m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
		require.NoError(t, err)

		ctx := context.Background()
		err = m.Start(ctx, false, nil)
		require.NoError(t, err)
		defer func() {
			err := m.Stop()
			require.NoError(t, err)
		}()

		err = m.TriggerRefresh(ctx)
		assert.ErrorIs(t, err, manager.ErrManagerNotConfiguredForAutoRefresh)
	})
}

func TestManager_EmptyState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := mock_fetcher.NewMockFetcher(ctrl)
	contentStore := mock_autofetcher.NewMockContentStore(ctrl)
	customTokenStore := mock_manager.NewMockCustomTokenStore(ctrl)
	config := createTestConfig()

	contentStore.EXPECT().GetAll().Return(map[string]autofetcher.Content{}, nil).AnyTimes()
	contentStore.EXPECT().Get(gomock.Any()).Return(autofetcher.Content{}, nil).AnyTimes()
	customTokenStore.EXPECT().GetAll().Return([]*types.Token{}, nil).AnyTimes()

	m, err := manager.New(config, mockFetcher, contentStore, customTokenStore)
	require.NoError(t, err)

	t.Run("operations before start", func(t *testing.T) {
		tokens := m.UniqueTokens()
		assert.Nil(t, tokens)

		token, exists := m.GetTokenByChainAddress(common.EthereumMainnet, gethcommon.HexToAddress("0x0000000000000000000000000000000000000000"))
		assert.False(t, exists)
		assert.Nil(t, token)

		chainTokens := m.GetTokensByChain(common.EthereumMainnet)
		assert.Nil(t, chainTokens)

		keyTokens, err := m.GetTokensByKeys([]string{"1-0x0000000000000000000000000000000000000000"})
		assert.NoError(t, err)
		assert.Nil(t, keyTokens)

		lists := m.TokenLists()
		assert.Nil(t, lists)

		list, exists := m.TokenList("native")
		assert.False(t, exists)
		assert.Nil(t, list)
	})
}
