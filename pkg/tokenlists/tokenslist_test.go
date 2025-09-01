package tokenlists

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewTokensList(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		uniqueTokens int
		wantErr      bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "missing main list",
			config: &Config{
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				PrivacyGuard:                  NewDefaultPrivacyGuard(false),
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  NewDefaultContentStore(),
				Parsers:                       make(map[string]Parser),
				CoinGeckoChainsMapper:         DefaultCoinGeckoChainsMapper,
				AutoRefreshInterval:           30 * time.Minute,
				AutoRefreshCheckInterval:      3 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "missing main list ID",
			config: &Config{
				MainList:                      []byte("{}"),
				Chains:                        []uint64{1},
				PrivacyGuard:                  NewDefaultPrivacyGuard(false),
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  NewDefaultContentStore(),
				Parsers:                       make(map[string]Parser),
				CoinGeckoChainsMapper:         DefaultCoinGeckoChainsMapper,
				AutoRefreshInterval:           30 * time.Minute,
				AutoRefreshCheckInterval:      3 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "missing chains",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				PrivacyGuard:                  NewDefaultPrivacyGuard(false),
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  NewDefaultContentStore(),
				Parsers:                       make(map[string]Parser),
				CoinGeckoChainsMapper:         DefaultCoinGeckoChainsMapper,
				AutoRefreshInterval:           30 * time.Minute,
				AutoRefreshCheckInterval:      3 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "missing privacy guard",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  NewDefaultContentStore(),
				Parsers:                       make(map[string]Parser),
				CoinGeckoChainsMapper:         DefaultCoinGeckoChainsMapper,
				AutoRefreshInterval:           30 * time.Minute,
				AutoRefreshCheckInterval:      3 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "missing last update time store",
			config: &Config{
				MainList:                 []byte("{}"),
				MainListID:               StatusListID,
				Chains:                   []uint64{1},
				PrivacyGuard:             NewDefaultPrivacyGuard(false),
				ContentStore:             NewDefaultContentStore(),
				Parsers:                  make(map[string]Parser),
				CoinGeckoChainsMapper:    DefaultCoinGeckoChainsMapper,
				AutoRefreshInterval:      30 * time.Minute,
				AutoRefreshCheckInterval: 3 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "missing content store",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				PrivacyGuard:                  NewDefaultPrivacyGuard(false),
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				Parsers:                       make(map[string]Parser),
				CoinGeckoChainsMapper:         DefaultCoinGeckoChainsMapper,
				AutoRefreshInterval:           30 * time.Minute,
				AutoRefreshCheckInterval:      3 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid refresh intervals",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				AutoRefreshInterval:           1 * time.Minute,
				AutoRefreshCheckInterval:      2 * time.Minute,
				PrivacyGuard:                  NewDefaultPrivacyGuard(false),
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  NewDefaultContentStore(),
			},
			wantErr: true,
		},
		{
			name: "not provided logger",
			config: &Config{
				MainList:                      []byte("{}"),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				PrivacyGuard:                  NewDefaultPrivacyGuard(false),
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  NewDefaultContentStore(),
				Parsers:                       make(map[string]Parser),
				AutoRefreshInterval:           30 * time.Minute,
				AutoRefreshCheckInterval:      3 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "valid config with initial lists",
			config: &Config{
				MainList:                      []byte(StatusTokenListJSON),
				MainListID:                    StatusListID,
				Chains:                        []uint64{1},
				PrivacyGuard:                  NewDefaultPrivacyGuard(false),
				LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
				ContentStore:                  NewDefaultContentStore(),
				Parsers:                       DefaultParsers,
				AutoRefreshInterval:           30 * time.Minute,
				AutoRefreshCheckInterval:      3 * time.Minute,
				logger:                        zap.NewNop(),
			},
			uniqueTokens: 2, // eth and status for chain 1
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config != nil {
				t.Logf("Config: AutoRefreshInterval=%v, AutoRefreshCheckInterval=%v",
					tt.config.AutoRefreshInterval, tt.config.AutoRefreshCheckInterval)
			}
			tl, err := NewTokensList(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				notifyCh := make(chan struct{}, 1)
				err = tl.Start(context.Background(), notifyCh)
				assert.NoError(t, err)

				tokens := tl.UniqueTokens()
				assert.Equal(t, tt.uniqueTokens, len(tokens))

				err = tl.Stop()
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokensList_Start(t *testing.T) {
	config := &Config{
		MainList:                      []byte("{}"),
		MainListID:                    StatusListID,
		Chains:                        []uint64{1},
		PrivacyGuard:                  NewDefaultPrivacyGuard(false),
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  NewDefaultContentStore(),
		Parsers:                       make(map[string]Parser),
		CoinGeckoChainsMapper:         DefaultCoinGeckoChainsMapper,
		AutoRefreshInterval:           30 * time.Minute,
		AutoRefreshCheckInterval:      3 * time.Minute,
		logger:                        zap.NewNop(),
	}

	tl, err := NewTokensList(config)
	require.NoError(t, err)

	ctx := context.Background()
	notifyCh := make(chan struct{}, 1)

	err = tl.Start(ctx, notifyCh)
	assert.NoError(t, err)

	// Give some time for the start operation to complete
	time.Sleep(100 * time.Millisecond)

	tokens := tl.UniqueTokens()
	assert.NotNil(t, tokens)
}

func TestTokensList_PrivacyModeUpdated(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		RemoteListOfTokenListsURL:     server.URL + listOfTokenListsURL,
		MainList:                      []byte("{}"),
		MainListID:                    StatusListID,
		Chains:                        []uint64{1},
		PrivacyGuard:                  NewDefaultPrivacyGuard(true),
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  NewDefaultContentStore(),
		Parsers:                       make(map[string]Parser),
		AutoRefreshInterval:           200 * time.Millisecond,
		AutoRefreshCheckInterval:      100 * time.Millisecond,
		logger:                        zap.NewNop(),
	}

	tl, err := NewTokensList(config)
	require.NoError(t, err)

	ctx := context.Background()

	ctxTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	notifyCh := make(chan struct{}, 1)

	err = tl.Start(ctx, notifyCh)
	require.NoError(t, err)

	// check that nothing will be fetched in privacy mode
	select {
	case <-notifyCh:
		t.Fatal("notifyCh received")
	case <-ctxTimeout.Done():
		// Expected behavior
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	config.PrivacyGuard.(*defaultPrivacyGuard).SetPrivacyMode(false)

	err = tl.PrivacyModeUpdated(ctx)
	assert.NoError(t, err)

	select {
	case <-notifyCh:
		// check if the content store has the data
		allContent, err := config.ContentStore.GetAll()
		assert.NoError(t, err)
		assert.Len(t, allContent, 3)

		assert.Contains(t, allContent, "status")
		assert.Contains(t, allContent, "uniswap")
		assert.Equal(t, statusTokenListJsonResponse, string(allContent["status"].Data))
		assert.Equal(t, uniswapTokenListJsonResponse, string(allContent["uniswap"].Data))
	case <-ctxTimeout.Done():
		t.Fatal("context done")
	}

	// reset the content store to check if the new data will be stored when switching back to privacy mode
	config.ContentStore = NewDefaultContentStore()

	ctxTimeout, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	config.PrivacyGuard.(*defaultPrivacyGuard).SetPrivacyMode(true)

	err = tl.PrivacyModeUpdated(ctx)
	assert.NoError(t, err)

	select {
	case <-notifyCh:
		t.Fatal("notifyCh received")
	case <-ctxTimeout.Done():
		// Expected behavior
	}

	allContent, err := config.ContentStore.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allContent, 0)

	err = tl.Stop()
	assert.NoError(t, err)
}

func TestTokensList_RefreshNow(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		RemoteListOfTokenListsURL:     server.URL + listOfTokenListsURL,
		MainList:                      []byte("{}"),
		MainListID:                    StatusListID,
		Chains:                        []uint64{1},
		PrivacyGuard:                  NewDefaultPrivacyGuard(true),
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  NewDefaultContentStore(),
		Parsers:                       make(map[string]Parser),
		AutoRefreshInterval:           200 * time.Millisecond,
		AutoRefreshCheckInterval:      100 * time.Millisecond,
		logger:                        zap.NewNop(),
	}

	tl, err := NewTokensList(config)
	require.NoError(t, err)

	lastRefreshTime, err := tl.LastRefreshTime()
	require.NoError(t, err)
	assert.True(t, lastRefreshTime.IsZero())

	ctx := context.Background()

	initialContent := Content{
		SourceURL: "https://example.com/status-token-list.json",
		Etag:      "123",
		Data:      []byte("some data"),
		Fetched:   time.Now(),
	}
	err = config.ContentStore.Set("initial-list", initialContent)
	require.NoError(t, err)

	ctxTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	notifyCh := make(chan struct{}, 1)

	err = tl.Start(ctx, notifyCh)
	require.NoError(t, err)

	// check that nothing will be fetched in privacy mode
	select {
	case <-notifyCh:
		t.Fatal("notifyCh received")
	case <-ctxTimeout.Done():
		// Expected behavior
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	err = tl.RefreshNow(ctx)
	assert.NoError(t, err)

	// check that only initial content is stored in the content store in privacy mode after RefreshNow
	select {
	case <-notifyCh:
		// check if the content store has the data
		allContent, err := config.ContentStore.GetAll()
		assert.NoError(t, err)
		assert.Len(t, allContent, 1)
		assert.Contains(t, allContent, "initial-list")
		assert.Equal(t, initialContent.Data, allContent["initial-list"].Data)
	case <-ctxTimeout.Done():
		t.Fatal("context done")
	}

	lastRefreshTime, err = tl.LastRefreshTime()
	require.NoError(t, err)
	assert.True(t, lastRefreshTime.IsZero())

	ctxTimeout, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	config.PrivacyGuard.(*defaultPrivacyGuard).SetPrivacyMode(false)

	err = tl.PrivacyModeUpdated(ctx)
	assert.NoError(t, err)

	// check that the content store has the initial content plus fetched content for non privacy mode after RefreshNow
	select {
	case <-notifyCh:
		// check if the content store has the data
		allContent, err := config.ContentStore.GetAll()
		assert.NoError(t, err)
		assert.Len(t, allContent, 4)

		assert.Contains(t, allContent, "initial-list")
		assert.Equal(t, initialContent.Data, allContent["initial-list"].Data)
		assert.Contains(t, allContent, "status")
		assert.Contains(t, allContent, "uniswap")
		assert.Equal(t, statusTokenListJsonResponse, string(allContent["status"].Data))
		assert.Equal(t, uniswapTokenListJsonResponse, string(allContent["uniswap"].Data))
	case <-ctxTimeout.Done():
		t.Fatal("context done")
	}

	lastRefreshTime, err = tl.LastRefreshTime()
	require.NoError(t, err)
	assert.False(t, lastRefreshTime.IsZero())

	err = tl.Stop()
	assert.NoError(t, err)
}

func TestTokensList_TokenMethods(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		RemoteListOfTokenListsURL:     server.URL + listOfTokenListsURL,
		MainList:                      []byte("{}"),
		MainListID:                    StatusListID,
		Parsers:                       DefaultParsers,
		Chains:                        common.AllChains,
		PrivacyGuard:                  NewDefaultPrivacyGuard(false),
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  NewDefaultContentStore(),
		AutoRefreshInterval:           200 * time.Millisecond,
		AutoRefreshCheckInterval:      100 * time.Millisecond,
		logger:                        zap.NewNop(),
	}

	tl, err := NewTokensList(config)
	require.NoError(t, err)

	tokens := tl.UniqueTokens()
	assert.Empty(t, tokens)

	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	notifyCh := make(chan struct{}, 1)

	err = tl.Start(ctx, notifyCh)
	require.NoError(t, err)

	select {
	case <-notifyCh:
		// Expected behavior
	case <-ctxTimeout.Done():
		t.Fatal("context done")
	}

	// Test after fetching the token list
	tokens = tl.UniqueTokens()
	assert.NotNil(t, tokens)

	// check if no duplicate tokens
	tokensByKey := make(map[string]struct{})
	tokensByChainID := make(map[uint64][]*Token)
	for _, token := range tokens {
		if _, ok := tokensByKey[token.Key()]; ok {
			t.Fatal("duplicate token", token.Key())
		}
		tokensByKey[token.Key()] = struct{}{}
		tokensByChainID[token.ChainID] = append(tokensByChainID[token.ChainID], token)
	}

	randomIndex := rand.Intn(len(tokens)) //nolint:gosec
	randomToken := tokens[randomIndex]

	token, ok := tl.GetTokenByChainAddress(randomToken.ChainID, randomToken.Address)
	assert.True(t, ok)
	assert.Equal(t, randomToken, token)

	tokensByChain := tl.GetTokensByChain(randomToken.ChainID)
	assert.Equal(t, len(tokensByChainID[randomToken.ChainID]), len(tokensByChain))

	for _, tokne := range tokensByChain {
		assert.Equal(t, tokne.ChainID, randomToken.ChainID)
	}

	tokenLists := tl.TokenLists()
	assert.Equal(t, len(tokenLists), 3)
	for _, tokenList := range tokenLists {
		var expectedTokensLength int
		switch tokenList.Name {
		case statusTokenList.Name:
			expectedTokensLength = len(statusTokenList.Tokens)
		case uniswapTokenList.Name:
			expectedTokensLength = len(uniswapTokenList.Tokens)
		default:
			expectedTokensLength = len(config.Chains)
		}

		assert.Equal(t, expectedTokensLength, len(tokenList.Tokens))
	}

	tokenList, ok := tl.TokenList(StatusListID)
	assert.True(t, ok)
	assert.Equal(t, statusTokenList.Name, tokenList.Name)
	assert.Equal(t, statusTokenList.Timestamp, tokenList.Timestamp)
	assert.Equal(t, statusTokenList.Version, tokenList.Version)
	assert.Equal(t, statusTokenList.Tags, tokenList.Tags)
	assert.Equal(t, statusTokenList.LogoURI, tokenList.LogoURI)
	assert.Equal(t, statusTokenList.Keywords, tokenList.Keywords)
	assert.Equal(t, len(statusTokenList.Tokens), len(tokenList.Tokens))

	// check if all native tokens have correct cross chain ID
	realNumOfNativeTokens := 0
	for _, token := range tokens {
		if !token.IsNative() {
			continue
		}

		realNumOfNativeTokens++
		if token.ChainID == common.BSCMainnet || token.ChainID == common.BSCTestnet {
			assert.Equal(t, BinanceSmartChainNativeCrossChainID, token.CrossChainID)
		} else {
			assert.Equal(t, EthereumNativeCrossChainID, token.CrossChainID)
		}
	}
	assert.Equal(t, len(config.Chains), realNumOfNativeTokens)

	err = tl.Stop()
	assert.NoError(t, err)
}
