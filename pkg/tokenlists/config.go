package tokenlists

import (
	"time"

	"go.uber.org/zap"
)

func (c *Config) WithMainList(id string, data []byte) *Config {
	c.MainListID = id
	c.MainList = data
	return c
}

func (c *Config) WithInitialLists(lists map[string][]byte) *Config {
	c.InitialLists = lists
	return c
}

func (c *Config) WithParsers(parsers map[string]Parser) *Config {
	c.Parsers = parsers
	return c
}

func (c *Config) WithChains(chains []uint64) *Config {
	c.Chains = chains
	return c
}

func (c *Config) WithCoinGeckoChainsMapper(mapper map[string]uint64) *Config {
	c.CoinGeckoChainsMapper = mapper
	return c
}

func (c *Config) WithRemoteListOfTokenListsURL(url string) *Config {
	c.RemoteListOfTokenListsURL = url
	return c
}

func (c *Config) WithAutoRefreshInterval(interval, checkInterval time.Duration) *Config {
	c.AutoRefreshInterval = interval
	c.AutoRefreshCheckInterval = checkInterval
	return c
}

func (c *Config) WithLogger(logger *zap.Logger) *Config {
	c.logger = logger
	return c
}

func (c *Config) WithPrivacyGuard(guard PrivacyGuard) *Config {
	c.PrivacyGuard = guard
	return c
}

func (c *Config) WithLastTokenListsUpdateTimeStore(store LastTokenListsUpdateTimeStore) *Config {
	c.LastTokenListsUpdateTimeStore = store
	return c
}

func (c *Config) WithContentStore(store ContentStore) *Config {
	c.ContentStore = store
	return c
}

func (c *Config) WithCustomTokenStore(store CustomTokenStore) *Config {
	c.CustomTokenStore = store
	return c
}
