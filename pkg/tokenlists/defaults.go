package tokenlists

import (
	"fmt"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"go.uber.org/zap"
)

const (
	defaultAutoRefreshInterval      = 30 * time.Minute // interval after which the token lists should be fetched from the remote source (or use the default one if remote source is not set)
	defaultAutoRefreshCheckInterval = 3 * time.Minute  // interval after which the auto-refresh should be checked if it should trigger the refresh
)

var DefaultParsers = map[string]Parser{
	StatusListID:             &StatusTokenListParser{},
	UniswapListID:            &StandardTokenListParser{}, // Uniswap token list follows the StandardTokenList format
	CoingeckoAllTokensListID: NewCoinGeckoAllTokensParser(DefaultCoinGeckoChainsMapper),
	// Coingecko platform specific token lists follow the StandardTokenList format
	CoingeckoEthereumListID: &StandardTokenListParser{},
	CoingeckoOptimismListID: &StandardTokenListParser{},
	CoingeckoArbitrumListID: &StandardTokenListParser{},
	CoingeckoBSCListID:      &StandardTokenListParser{},
	CoingeckoBaseListID:     &StandardTokenListParser{},
}

// DefaultCoinGeckoChainsMapper provides the default mapping from CoinGecko platform names to chain IDs.
var DefaultCoinGeckoChainsMapper = map[string]common.ChainID{
	"ethereum":            common.EthereumMainnet,
	"optimistic-ethereum": common.OptimismMainnet,
	"arbitrum-one":        common.ArbitrumMainnet,
	"binance-smart-chain": common.BSCMainnet,
	"base":                common.BaseMainnet,
}

// defaultPrivacyGuard provides a default privacy guard implementation.
type defaultPrivacyGuard struct {
	privacyOn bool
}

func (p *defaultPrivacyGuard) IsPrivacyOn() (bool, error) {
	return p.privacyOn, nil
}

// SetPrivacyMode is not the interface method, but it's here to be able to set the privacy on for testing
func (p *defaultPrivacyGuard) SetPrivacyMode(privacyOn bool) {
	p.privacyOn = privacyOn
}

func NewDefaultPrivacyGuard(initialPrivacy bool) PrivacyGuard {
	return &defaultPrivacyGuard{privacyOn: initialPrivacy}
}

// defaultLastTokenListsUpdateTimeStore provides a default last token lists update time store implementation.
type defaultLastTokenListsUpdateTimeStore struct {
	lastUpdateTime time.Time
}

func (s *defaultLastTokenListsUpdateTimeStore) Get() (time.Time, error) {
	return s.lastUpdateTime, nil
}

func (s *defaultLastTokenListsUpdateTimeStore) Set(time time.Time) error {
	s.lastUpdateTime = time
	return nil
}

func NewDefaultLastTokenListsUpdateTimeStore() LastTokenListsUpdateTimeStore {
	return &defaultLastTokenListsUpdateTimeStore{}
}

// defaultContentStore provides a default content store implementation.
type defaultContentStore struct {
	content map[string]Content
}

func (s *defaultContentStore) GetEtag(id string) (string, error) {
	content, ok := s.content[id]
	if !ok {
		return "", fmt.Errorf("etag not found")
	}
	return content.Etag, nil
}

func (s *defaultContentStore) Get(id string) (Content, error) {
	content, ok := s.content[id]
	if !ok {
		return Content{}, fmt.Errorf("content not found")
	}
	return content, nil
}

func (s *defaultContentStore) Set(id string, content Content) error {
	s.content[id] = content
	return nil
}

func (s *defaultContentStore) GetAll() (map[string]Content, error) {
	return s.content, nil
}

func NewDefaultContentStore() ContentStore {
	return &defaultContentStore{
		content: make(map[string]Content),
	}
}

// defaultCustomTokenStore provides a default custom token store implementation.
type defaultCustomTokenStore struct {
	customTokens []*Token
}

func (s *defaultCustomTokenStore) GetAll() ([]*Token, error) {
	return s.customTokens, nil
}

func NewDefaultCustomTokenStore() CustomTokenStore {
	return &defaultCustomTokenStore{}
}

// DefaultConfig provides sensible defaults for TokensList configuration.
func DefaultConfig() *Config {
	return &Config{
		Chains:                        common.AllChains,
		CoinGeckoChainsMapper:         DefaultCoinGeckoChainsMapper,
		MainListID:                    StatusListID,
		AutoRefreshInterval:           defaultAutoRefreshInterval,
		AutoRefreshCheckInterval:      defaultAutoRefreshCheckInterval,
		logger:                        zap.NewNop(),
		PrivacyGuard:                  NewDefaultPrivacyGuard(false),
		LastTokenListsUpdateTimeStore: NewDefaultLastTokenListsUpdateTimeStore(),
		ContentStore:                  NewDefaultContentStore(),
		CustomTokenStore:              NewDefaultCustomTokenStore(),
		Parsers:                       make(map[string]Parser),
	}
}
