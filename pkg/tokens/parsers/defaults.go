package parsers

import (
	"github.com/status-im/go-wallet-sdk/pkg/common"
)

// DefaultCoinGeckoChainsMapper provides the default mapping from CoinGecko platform names to chain IDs.
var DefaultCoinGeckoChainsMapper = map[string]common.ChainID{
	"ethereum":            common.EthereumMainnet,
	"optimistic-ethereum": common.OptimismMainnet,
	"arbitrum-one":        common.ArbitrumMainnet,
	"binance-smart-chain": common.BSCMainnet,
	"base":                common.BaseMainnet,
}
