package common

type ChainID = uint64

const (
	UnknownChainID       ChainID = 0
	EthereumMainnet      ChainID = 1
	EthereumSepolia      ChainID = 11155111
	OptimismMainnet      ChainID = 10
	OptimismSepolia      ChainID = 11155420
	ArbitrumMainnet      ChainID = 42161
	ArbitrumSepolia      ChainID = 421614
	BSCMainnet           ChainID = 56
	BSCTestnet           ChainID = 97
	BaseMainnet          ChainID = 8453
	BaseSepolia          ChainID = 84532
	StatusNetworkSepolia ChainID = 1660990954
)

var AllChains = []ChainID{
	EthereumMainnet,
	EthereumSepolia,
	OptimismMainnet,
	OptimismSepolia,
	ArbitrumMainnet,
	ArbitrumSepolia,
	BSCMainnet,
	BSCTestnet,
	BaseMainnet,
	BaseSepolia,
	StatusNetworkSepolia,
}
