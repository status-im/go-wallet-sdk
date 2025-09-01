package tokenlists

import (
	"testing"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDefaultParsers(t *testing.T) {
	assert.NotNil(t, DefaultParsers)

	assert.Contains(t, DefaultParsers, StatusListID)
	assert.NotNil(t, DefaultParsers[StatusListID])

	assert.Contains(t, DefaultParsers, UniswapListID)
	assert.NotNil(t, DefaultParsers[UniswapListID])

	assert.Contains(t, DefaultParsers, CoingeckoAllTokensListID)
	assert.NotNil(t, DefaultParsers[CoingeckoAllTokensListID])

	assert.Contains(t, DefaultParsers, CoingeckoEthereumListID)
	assert.NotNil(t, DefaultParsers[CoingeckoEthereumListID])

	assert.Contains(t, DefaultParsers, CoingeckoOptimismListID)
	assert.NotNil(t, DefaultParsers[CoingeckoOptimismListID])

	assert.Contains(t, DefaultParsers, CoingeckoArbitrumListID)
	assert.NotNil(t, DefaultParsers[CoingeckoArbitrumListID])

	assert.Contains(t, DefaultParsers, CoingeckoBSCListID)
	assert.NotNil(t, DefaultParsers[CoingeckoBSCListID])

	assert.Contains(t, DefaultParsers, CoingeckoBaseListID)
	assert.NotNil(t, DefaultParsers[CoingeckoBaseListID])
}

func TestDefaultCoinGeckoChainsMapper(t *testing.T) {
	assert.NotEmpty(t, DefaultCoinGeckoChainsMapper)

	assert.Contains(t, DefaultCoinGeckoChainsMapper, "ethereum")
	assert.Equal(t, common.EthereumMainnet, DefaultCoinGeckoChainsMapper["ethereum"])

	assert.Contains(t, DefaultCoinGeckoChainsMapper, "optimistic-ethereum")
	assert.Equal(t, common.OptimismMainnet, DefaultCoinGeckoChainsMapper["optimistic-ethereum"])

	assert.Contains(t, DefaultCoinGeckoChainsMapper, "arbitrum-one")
	assert.Equal(t, common.ArbitrumMainnet, DefaultCoinGeckoChainsMapper["arbitrum-one"])

	assert.Contains(t, DefaultCoinGeckoChainsMapper, "binance-smart-chain")
	assert.Equal(t, common.BSCMainnet, DefaultCoinGeckoChainsMapper["binance-smart-chain"])

	assert.Contains(t, DefaultCoinGeckoChainsMapper, "base")
	assert.Equal(t, common.BaseMainnet, DefaultCoinGeckoChainsMapper["base"])
}

func TestNewDefaultPrivacyGuard(t *testing.T) {
	guard := NewDefaultPrivacyGuard(false)
	assert.NotNil(t, guard)
	privacyOn, err := guard.IsPrivacyOn()
	assert.NoError(t, err)
	assert.False(t, privacyOn)

	guard = NewDefaultPrivacyGuard(true)
	assert.NotNil(t, guard)
	privacyOn, err = guard.IsPrivacyOn()
	assert.NoError(t, err)
	assert.True(t, privacyOn)
}

func TestNewDefaultLastTokenListsUpdateTimeStore(t *testing.T) {
	store := NewDefaultLastTokenListsUpdateTimeStore()
	assert.NotNil(t, store)
}

func TestDefaultLastTokenListsUpdateTimeStore_GetSet(t *testing.T) {
	store := NewDefaultLastTokenListsUpdateTimeStore()

	initialTime, err := store.Get()
	assert.NoError(t, err)
	assert.Equal(t, time.Time{}, initialTime)

	testTime := time.Date(2025, 9, 30, 12, 0, 0, 0, time.UTC)
	err = store.Set(testTime)
	assert.NoError(t, err)

	retrievedTime, err := store.Get()
	assert.NoError(t, err)
	assert.Equal(t, testTime, retrievedTime)
}

func TestNewDefaultContentStore(t *testing.T) {
	store := NewDefaultContentStore()
	assert.NotNil(t, store)
}

func TestDefaultContentStore_Operations(t *testing.T) {
	store := NewDefaultContentStore()

	etag, err := store.GetEtag("test-id")
	assert.Error(t, err)
	assert.Empty(t, etag)

	content, err := store.Get("test-id")
	assert.Error(t, err)
	assert.Equal(t, Content{}, content)

	allContent, err := store.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allContent, 0)

	testContent := Content{
		SourceURL: "https://example.com/test.json",
		Etag:      "test-etag",
		Data:      []byte("test data"),
		Fetched:   time.Now(),
	}

	err = store.Set("test-id", testContent)
	assert.NoError(t, err)

	etag, err = store.GetEtag("test-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-etag", etag)

	content, err = store.Get("test-id")
	assert.NoError(t, err)
	assert.Equal(t, testContent, content)

	allContent, err = store.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allContent, 1)
	assert.Contains(t, allContent, "test-id")
}

func TestNewDefaultCustomTokenStore(t *testing.T) {
	store := NewDefaultCustomTokenStore()
	assert.NotNil(t, store)
}

func TestDefaultCustomTokenStore_GetAll(t *testing.T) {
	store := &defaultCustomTokenStore{
		customTokens: []*Token{
			{
				Symbol:  "CUSTOM1",
				Name:    "Custom Token 1",
				ChainID: 1,
			},
			{
				Symbol:  "CUSTOM2",
				Name:    "Custom Token 2",
				ChainID: 137,
			},
		},
	}

	tokens, err := store.GetAll()
	assert.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Equal(t, "CUSTOM1", tokens[0].Symbol)
	assert.Equal(t, "CUSTOM2", tokens[1].Symbol)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotNil(t, config)

	// Test default values
	assert.Equal(t, common.AllChains, config.Chains)
	assert.Equal(t, DefaultCoinGeckoChainsMapper, config.CoinGeckoChainsMapper)
	assert.Equal(t, StatusListID, config.MainListID)
	assert.Equal(t, defaultAutoRefreshInterval, config.AutoRefreshInterval)
	assert.Equal(t, defaultAutoRefreshCheckInterval, config.AutoRefreshCheckInterval)
	assert.NotNil(t, config.logger)
	assert.NotNil(t, config.PrivacyGuard)
	assert.NotNil(t, config.LastTokenListsUpdateTimeStore)
	assert.NotNil(t, config.ContentStore)
	assert.NotNil(t, config.CustomTokenStore)
	assert.NotNil(t, config.Parsers)
}

func TestConfig_WithMainList(t *testing.T) {
	config := &Config{}
	mainListData := []byte("test data")
	mainListID := "test-main-list"

	result := config.WithMainList(mainListID, mainListData)
	assert.Equal(t, config, result)
	assert.Equal(t, mainListID, config.MainListID)
	assert.Equal(t, mainListData, config.MainList)
}

func TestConfig_WithInitialLists(t *testing.T) {
	config := &Config{}
	initialLists := map[string][]byte{
		"list1": []byte("data1"),
		"list2": []byte("data2"),
	}

	result := config.WithInitialLists(initialLists)
	assert.Equal(t, config, result)
	assert.Equal(t, initialLists, config.InitialLists)
}

func TestConfig_WithParsers(t *testing.T) {
	config := &Config{}
	parsers := map[string]Parser{
		"parser1": &StatusTokenListParser{},
		"parser2": &StandardTokenListParser{},
	}

	result := config.WithParsers(parsers)
	assert.Equal(t, config, result)
	assert.Equal(t, parsers, config.Parsers)
}

func TestConfig_WithChains(t *testing.T) {
	config := &Config{}
	chains := []uint64{1, 137, 56}

	result := config.WithChains(chains)
	assert.Equal(t, config, result)
	assert.Equal(t, chains, config.Chains)
}

func TestConfig_WithCoinGeckoChainsMapper(t *testing.T) {
	config := &Config{}
	mapper := map[string]uint64{
		"ethereum": 1,
		"polygon":  137,
	}

	result := config.WithCoinGeckoChainsMapper(mapper)
	assert.Equal(t, config, result)
	assert.Equal(t, mapper, config.CoinGeckoChainsMapper)
}

func TestConfig_WithRemoteListOfTokenListsURL(t *testing.T) {
	config := &Config{}
	url := "https://example.com/tokenlists.json"

	result := config.WithRemoteListOfTokenListsURL(url)
	assert.Equal(t, config, result)
	assert.Equal(t, url, config.RemoteListOfTokenListsURL)
}

func TestConfig_WithAutoRefreshInterval(t *testing.T) {
	config := &Config{}
	interval := 1 * time.Hour
	checkInterval := 10 * time.Minute

	result := config.WithAutoRefreshInterval(interval, checkInterval)
	assert.Equal(t, config, result)
	assert.Equal(t, interval, config.AutoRefreshInterval)
	assert.Equal(t, checkInterval, config.AutoRefreshCheckInterval)
}

func TestConfig_WithLogger(t *testing.T) {
	config := &Config{}
	logger := zap.NewNop()

	result := config.WithLogger(logger)
	assert.Equal(t, config, result)
	assert.Equal(t, logger, config.logger)
}

func TestConfig_WithPrivacyGuard(t *testing.T) {
	config := &Config{}
	guard := &defaultPrivacyGuard{}

	result := config.WithPrivacyGuard(guard)
	assert.Equal(t, config, result)
	assert.Equal(t, guard, config.PrivacyGuard)
}

func TestConfig_WithLastTokenListsUpdateTimeStore(t *testing.T) {
	config := &Config{}
	store := NewDefaultLastTokenListsUpdateTimeStore()

	result := config.WithLastTokenListsUpdateTimeStore(store)
	assert.Equal(t, config, result)
	assert.Equal(t, store, config.LastTokenListsUpdateTimeStore)
}

func TestConfig_WithContentStore(t *testing.T) {
	config := &Config{}
	store := &defaultContentStore{}

	result := config.WithContentStore(store)
	assert.Equal(t, config, result)
	assert.Equal(t, store, config.ContentStore)
}

func TestConfig_WithCustomTokenStore(t *testing.T) {
	config := &Config{}
	store := &defaultCustomTokenStore{}

	result := config.WithCustomTokenStore(store)
	assert.Equal(t, config, result)
	assert.Equal(t, store, config.CustomTokenStore)
}
