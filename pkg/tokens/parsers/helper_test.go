package parsers_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

//go:embed test/coingecko_tokens_response.json
var coingeckoTokensJsonResponse string

//go:embed test/coingecko_tokens_response_invalid_tokens.json
var coingeckoTokensJsonResponseInvalidTokens string

//go:embed test/status_token_list_response_template.json
var statusTokenListJsonResponseTemplate string

//go:embed test/status_tokens_response.json
var statusTokensJsonResponse string

//go:embed test/status_invalid_tokens_response.json
var statusInvalidTokensJsonResponse string

//go:embed test/uniswap_tokens_response.json
var uniswapTokensJsonResponse string

//go:embed test/uniswap_invalid_tokens_response.json
var uniswapInvalidTokensJsonResponse string

var fetchedCoingeckoTokenList = fetcher.FetchedData{
	FetchDetails: fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			SourceURL: "https://example.com/coingecko-token-list.json",
		},
	},
	JsonData: []byte(coingeckoTokensJsonResponse),
}

var fetchedCoingeckoTokenListInvalidTokens = fetcher.FetchedData{
	FetchDetails: fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			SourceURL: "https://example.com/coingecko-token-list.json",
		},
	},
	JsonData: []byte(coingeckoTokensJsonResponseInvalidTokens),
}

var coingeckoTokenList = types.TokenList{
	Tokens: []*types.Token{
		{
			ChainID: 1,
			Address: common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
			Name:    "USDC",
			Symbol:  "usdc",
		},
		{
			ChainID: 42161,
			Address: common.HexToAddress("0xaf88d065e77c8cc2239327c5edb3a432268e5831"),
			Name:    "USDC",
			Symbol:  "usdc",
		},
		{
			ChainID: 10,
			Address: common.HexToAddress("0x0b2c639c533813f4aa9d7837caf62653d097ff85"),
			Name:    "USDC",
			Symbol:  "usdc",
		},
		{
			ChainID: 8453,
			Address: common.HexToAddress("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913"),
			Name:    "USDC",
			Symbol:  "usdc",
		},
		{
			ChainID: 1,
			Address: common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"),
			Name:    "Wrapped Bitcoin",
			Symbol:  "wbtc",
		},
	},
}

var coingeckoTokenListInvalidTokens = types.TokenList{
	Tokens: []*types.Token{
		{
			ChainID: 10,
			Address: common.HexToAddress("0x0b2c639c533813f4aa9d7837caf62653d097ff85"),
			Name:    "USDC",
			Symbol:  "usdc",
		},
		{
			ChainID: 8453,
			Address: common.HexToAddress("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913"),
			Name:    "USDC",
			Symbol:  "usdc",
		},
		{
			ChainID: 1,
			Address: common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"),
			Name:    "Wrapped Bitcoin",
			Symbol:  "wbtc",
		},
	},
}

func createStatusTokenListJsonResponse(name string, timestamp string, major int, minor int, tokens string) string {
	list := strings.ReplaceAll(statusTokenListJsonResponseTemplate, "NAME", name)
	list = strings.ReplaceAll(list, "TIMESTAMP", timestamp)
	list = strings.ReplaceAll(list, "MAJOR", fmt.Sprintf("%d", major))
	list = strings.ReplaceAll(list, "MINOR", fmt.Sprintf("%d", minor))
	return strings.ReplaceAll(list, "TOKENS", tokens)
}

var statusTokenListJsonResponse = createStatusTokenListJsonResponse("Status Token List", "2025-09-01T13:00:00.000Z", 0, 1, statusTokensJsonResponse)

var statusTokenListInvalidTokensJsonResponse = createStatusTokenListJsonResponse("Status Token List", "2025-09-01T13:00:00.000Z", 0, 1, statusInvalidTokensJsonResponse)

var statusEmptyTokensJsonResponse = createStatusTokenListJsonResponse("Status Token List", "2025-09-01T13:00:00.000Z", 0, 1, "[]")

var fetchedStatusTokenList = createFetchedTokenListFromResponse(statusTokenListJsonResponse)

var fetchedStatusTokenListInvalidTokens = createFetchedTokenListFromResponse(statusTokenListInvalidTokensJsonResponse)

var fetchedStatusTokenListEmpty = createFetchedTokenListFromResponse(statusEmptyTokensJsonResponse)

var statusTokenListEmpty = types.TokenList{
	Name:      "Status Token List",
	Timestamp: "2025-09-01T13:00:00.000Z",
	Version: types.Version{
		Major: 0,
		Minor: 1,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens:   []*types.Token{},
}

var statusTokenListInvalidTokens = types.TokenList{
	Name:      "Status Token List",
	Timestamp: "2025-09-01T13:00:00.000Z",
	Version: types.Version{
		Major: 0,
		Minor: 1,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens: []*types.Token{
		{
			CrossChainID: "status",
			ChainID:      1,
			Address:      common.HexToAddress("0x744d70fdbe2ba4cf95131626614a1763df805b9e"),
			Name:         "Status",
			Symbol:       "SNT",
			Decimals:     18,
			LogoURI:      "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
	},
}

var statusTokenList = types.TokenList{
	Name:      "Status Token List",
	Timestamp: "2025-09-01T13:00:00.000Z",
	Version: types.Version{
		Major: 0,
		Minor: 1,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens: []*types.Token{
		{
			CrossChainID: "status",
			ChainID:      1,
			Address:      common.HexToAddress("0x744d70fdbe2ba4cf95131626614a1763df805b9e"),
			Name:         "Status",
			Symbol:       "SNT",
			Decimals:     18,
			LogoURI:      "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			CrossChainID: "status",
			ChainID:      10,
			Address:      common.HexToAddress("0x650af3c15af43dcb218406d30784416d64cfb6b2"),
			Name:         "Status",
			Symbol:       "SNT",
			Decimals:     18,
			LogoURI:      "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			CrossChainID: "status",
			ChainID:      8453,
			Address:      common.HexToAddress("0x662015ec830df08c0fc45896fab726542e8ac09e"),
			Name:         "Status",
			Symbol:       "SNT",
			Decimals:     18,
			LogoURI:      "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			CrossChainID: "status",
			ChainID:      42161,
			Address:      common.HexToAddress("0x707f635951193ddafbb40971a0fcaab8a6415160"),
			Name:         "Status",
			Symbol:       "SNT",
			Decimals:     18,
			LogoURI:      "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			CrossChainID: "status-test-token",
			ChainID:      84532,
			Address:      common.HexToAddress("0xfdb3b57944943a7724fcc0520ee2b10659969a06"),
			Name:         "Status Test Token",
			Symbol:       "STT",
			Decimals:     18,
			LogoURI:      "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			CrossChainID: "status-test-token",
			ChainID:      11155111,
			Address:      common.HexToAddress("0xe452027cdef746c7cd3db31cb700428b16cd8e51"),
			Name:         "Status Test Token",
			Symbol:       "STT",
			Decimals:     18,
			LogoURI:      "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			CrossChainID: "status-test-token",
			ChainID:      1660990954,
			Address:      common.HexToAddress("0x1c3ac2a186c6149ae7cb4d716ebbd0766e4f898a"),
			Name:         "Status Test Token",
			Symbol:       "STT",
			Decimals:     18,
			LogoURI:      "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      1,
			Address:      common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      10,
			Address:      common.HexToAddress("0x0b2c639c533813f4aa9d7837caf62653d097ff85"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      8453,
			Address:      common.HexToAddress("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      42161,
			Address:      common.HexToAddress("0xaf88d065e77c8cc2239327c5edb3a432268e5831"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      84532,
			Address:      common.HexToAddress("0x036cbd53842c5426634e7929541ec2318f3dcf7e"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      421614,
			Address:      common.HexToAddress("0x75faf114eafb1bdbe2f0316df893fd58ce46aa4d"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      11155111,
			Address:      common.HexToAddress("0x1c7d4b196cb0c7b01d743fbc6116a902379c7238"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      11155420,
			Address:      common.HexToAddress("0x5fd84259d66cd46123540766be93dfe6d43130d7"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin",
			ChainID:      1660990954,
			Address:      common.HexToAddress("0xc445a18ca49190578dad62fba3048c07efc07ffe"),
			Name:         "USDC (EVM)",
			Symbol:       "USDC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "usd-coin-bsc",
			ChainID:      56,
			Address:      common.HexToAddress("0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d"),
			Name:         "USDC (BSC)",
			Symbol:       "USDC",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			CrossChainID: "tether",
			ChainID:      1,
			Address:      common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
			Name:         "USDT (EVM)",
			Symbol:       "USDT",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xdAC17F958D2ee523a2206206994597C13D831ec7/logo.png",
		},
		{
			CrossChainID: "tether",
			ChainID:      10,
			Address:      common.HexToAddress("0x94b008aa00579c1307b0ef2c499ad98a8ce58e58"),
			Name:         "USDT (EVM)",
			Symbol:       "USDT",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xdAC17F958D2ee523a2206206994597C13D831ec7/logo.png",
		},
		{
			CrossChainID: "tether",
			ChainID:      8453,
			Address:      common.HexToAddress("0xfde4c96c8593536e31f229ea8f37b2ada2699bb2"),
			Name:         "USDT (EVM)",
			Symbol:       "USDT",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xdAC17F958D2ee523a2206206994597C13D831ec7/logo.png",
		},
		{
			CrossChainID: "tether",
			ChainID:      42161,
			Address:      common.HexToAddress("0xfd086bc7cd5c481dcc9c85ebe478a1c0b69fcbb9"),
			Name:         "USDT (EVM)",
			Symbol:       "USDT",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xdAC17F958D2ee523a2206206994597C13D831ec7/logo.png",
		},
		{
			CrossChainID: "tether-bsc",
			ChainID:      56,
			Address:      common.HexToAddress("0x55d398326f99059ff775485246999027b3197955"),
			Name:         "USDT (BSC)",
			Symbol:       "USDT",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xdAC17F958D2ee523a2206206994597C13D831ec7/logo.png",
		},
		{
			CrossChainID: "dai",
			ChainID:      1,
			Address:      common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
			Name:         "DAI",
			Symbol:       "DAI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x6B175474E89094C44Da98b954EedeAC495271d0F/logo.png",
		},
		{
			CrossChainID: "dai",
			ChainID:      56,
			Address:      common.HexToAddress("0x1af3f329e8be154074d8769d1ffa4ee058b1dbc3"),
			Name:         "DAI",
			Symbol:       "DAI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x6B175474E89094C44Da98b954EedeAC495271d0F/logo.png",
		},
		{
			CrossChainID: "dai",
			ChainID:      8453,
			Address:      common.HexToAddress("0x50c5725949a6f0c72e6c4a641f24049a917db0cb"),
			Name:         "DAI",
			Symbol:       "DAI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x6B175474E89094C44Da98b954EedeAC495271d0F/logo.png",
		},
		{
			CrossChainID: "dai",
			ChainID:      42161,
			Address:      common.HexToAddress("0xda10009cbd5d07dd0cecc66161fc93d7c9000da1"),
			Name:         "DAI",
			Symbol:       "DAI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x6B175474E89094C44Da98b954EedeAC495271d0F/logo.png",
		},
		{
			CrossChainID: "dai",
			ChainID:      11155111,
			Address:      common.HexToAddress("0x3e622317f8c93f7328350cf0b56d9ed4c620c5d6"),
			Name:         "DAI",
			Symbol:       "DAI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x6B175474E89094C44Da98b954EedeAC495271d0F/logo.png",
		},
		{
			CrossChainID: "binancecoin",
			ChainID:      56,
			Address:      common.HexToAddress("0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c"),
			Name:         "Wrapped BNB",
			Symbol:       "WBNB",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/Uniswap/assets/master/blockchains/smartchain/assets/0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c/logo.png",
		},
		{
			CrossChainID: "chainlink",
			ChainID:      1,
			Address:      common.HexToAddress("0x514910771af9ca656af840dff83e8264ecf986ca"),
			Name:         "Chainlink",
			Symbol:       "LINK",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x514910771AF9Ca656af840dff83E8264EcF986CA/logo.png",
		},
		{
			CrossChainID: "chainlink",
			ChainID:      10,
			Address:      common.HexToAddress("0x350a791bfc2c21f9ed5d10980dad2e2638ffa7f6"),
			Name:         "Chainlink",
			Symbol:       "LINK",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x514910771AF9Ca656af840dff83E8264EcF986CA/logo.png",
		},
		{
			CrossChainID: "chainlink",
			ChainID:      56,
			Address:      common.HexToAddress("0xf8a0bf9cf54bb92f17374d9e9a321e6a111a51bd"),
			Name:         "Chainlink",
			Symbol:       "LINK",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x514910771AF9Ca656af840dff83E8264EcF986CA/logo.png",
		},
		{
			CrossChainID: "chainlink",
			ChainID:      8453,
			Address:      common.HexToAddress("0x88fb150bdc53a65fe94dea0c9ba0a6daf8c6e196"),
			Name:         "Chainlink",
			Symbol:       "LINK",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x514910771AF9Ca656af840dff83E8264EcF986CA/logo.png",
		},
		{
			CrossChainID: "chainlink",
			ChainID:      42161,
			Address:      common.HexToAddress("0xf97f4df75117a78c1a5a0dbb814af92458539fb4"),
			Name:         "Chainlink",
			Symbol:       "LINK",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x514910771AF9Ca656af840dff83E8264EcF986CA/logo.png",
		},
		{
			CrossChainID: "wrapped-bitcoin",
			ChainID:      1,
			Address:      common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599"),
			Name:         "Wrapped BTC",
			Symbol:       "WBTC",
			Decimals:     8,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599/logo.png",
		},
		{
			CrossChainID: "wrapped-bitcoin",
			ChainID:      10,
			Address:      common.HexToAddress("0x68f180fcce6836688e9084f035309e29bf0a2095"),
			Name:         "Wrapped BTC",
			Symbol:       "WBTC",
			Decimals:     8,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599/logo.png",
		},
		{
			CrossChainID: "wrapped-bitcoin",
			ChainID:      42161,
			Address:      common.HexToAddress("0x2f2a2543b76a4166549f7aab2e75bef0aefc5b0f"),
			Name:         "Wrapped BTC",
			Symbol:       "WBTC",
			Decimals:     8,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599/logo.png",
		},
		{
			CrossChainID: "ethena-usde",
			ChainID:      1,
			Address:      common.HexToAddress("0x4c9edd5852cd905f086c759e8383e09bff1e68b3"),
			Name:         "Ethena USDe",
			Symbol:       "USDE",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x4c9edd5852cd905f086c759e8383e09bff1e68b3/logo.png",
		},
		{
			CrossChainID: "ethena-usde",
			ChainID:      42161,
			Address:      common.HexToAddress("0x5d3a1ff2b6bab83b63cd9ad0787074081a52ef34"),
			Name:         "Ethena USDe",
			Symbol:       "USDE",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x4c9edd5852cd905f086c759e8383e09bff1e68b3/logo.png",
		},
		{
			CrossChainID: "uniswap",
			ChainID:      1,
			Address:      common.HexToAddress("0x1f9840a85d5af5bf1d1762f925bdaddc4201f984"),
			Name:         "Uniswap",
			Symbol:       "UNI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1f9840a85d5af5bf1d1762f925bdaddc4201f984/logo.png",
		},
		{
			CrossChainID: "uniswap",
			ChainID:      10,
			Address:      common.HexToAddress("0x6fd9d7ad17242c41f7131d257212c54a0e816691"),
			Name:         "Uniswap",
			Symbol:       "UNI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1f9840a85d5af5bf1d1762f925bdaddc4201f984/logo.png",
		},
		{
			CrossChainID: "uniswap",
			ChainID:      56,
			Address:      common.HexToAddress("0xbf5140a22578168fd562dccf235e5d43a02ce9b1"),
			Name:         "Uniswap",
			Symbol:       "UNI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1f9840a85d5af5bf1d1762f925bdaddc4201f984/logo.png",
		},
		{
			CrossChainID: "uniswap",
			ChainID:      42161,
			Address:      common.HexToAddress("0xfa7f8980b0f1e64a2062791cc3b0871572f1f7f0"),
			Name:         "Uniswap",
			Symbol:       "UNI",
			Decimals:     18,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1f9840a85d5af5bf1d1762f925bdaddc4201f984/logo.png",
		},
		{
			CrossChainID: "eurc",
			ChainID:      1,
			Address:      common.HexToAddress("0x1abaea1f7c830bd89acc67ec4af516284b1bc33c"),
			Name:         "EURC",
			Symbol:       "EURC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1abaea1f7c830bd89acc67ec4af516284b1bc33c/logo.png",
		},
		{
			CrossChainID: "eurc",
			ChainID:      8453,
			Address:      common.HexToAddress("0x60a3e35cc302bfa44cb288bc5a4f316fdb1adb42"),
			Name:         "EURC",
			Symbol:       "EURC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1abaea1f7c830bd89acc67ec4af516284b1bc33c/logo.png",
		},
		{
			CrossChainID: "eurc",
			ChainID:      42161,
			Address:      common.HexToAddress("0x0863708032b5c328e11abcb0df9d79c71fc52a48"),
			Name:         "EURC",
			Symbol:       "EURC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1abaea1f7c830bd89acc67ec4af516284b1bc33c/logo.png",
		},
		{
			CrossChainID: "eurc",
			ChainID:      84532,
			Address:      common.HexToAddress("0x808456652fdb597867f38412077a9182bf77359f"),
			Name:         "EURC",
			Symbol:       "EURC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1abaea1f7c830bd89acc67ec4af516284b1bc33c/logo.png",
		},
		{
			CrossChainID: "eurc",
			ChainID:      11155111,
			Address:      common.HexToAddress("0x08210f9170f89ab7658f0b5e3ff39b0e03c594d4"),
			Name:         "EURC",
			Symbol:       "EURC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1abaea1f7c830bd89acc67ec4af516284b1bc33c/logo.png",
		},
		{
			CrossChainID: "eurc",
			ChainID:      1660990954,
			Address:      common.HexToAddress("0xfe8be27656b1508194d9302d12a940b4d7c35b99"),
			Name:         "EURC",
			Symbol:       "EURC",
			Decimals:     6,
			LogoURI:      "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0x1abaea1f7c830bd89acc67ec4af516284b1bc33c/logo.png",
		},
	},
}

var uniswapTokenListTokensJsonResponse = createStatusTokenListJsonResponse("Uniswap Labs Default", "2025-08-26T21:30:26.717Z", 13, 45, uniswapTokensJsonResponse)

var uniswapTokenListInvalidTokensJsonResponse = createStatusTokenListJsonResponse("Uniswap Labs Default", "2025-08-26T21:30:26.717Z", 13, 45, uniswapInvalidTokensJsonResponse)

var uniswapTokenListEmptyTokensJsonResponse = createStatusTokenListJsonResponse("Uniswap Labs Default", "2025-08-26T21:30:26.717Z", 13, 45, "[]")

func createFetchedTokenListFromResponse(response string) fetcher.FetchedData {
	var list fetcher.FetchedData
	err := json.Unmarshal([]byte(response), &list)
	if err != nil {
		panic(err)
	}
	list.Fetched = time.Now()
	return list
}

var fetchedUniswapTokenList = createFetchedTokenListFromResponse(uniswapTokenListTokensJsonResponse)

var fetchedUniswapTokenListInvalidTokens = createFetchedTokenListFromResponse(uniswapTokenListInvalidTokensJsonResponse)

var fetchedUniswapTokenListEmpty = createFetchedTokenListFromResponse(uniswapTokenListEmptyTokensJsonResponse)

var uniswapTokenListEmpty = types.TokenList{
	Name:      "Uniswap Labs Default",
	Timestamp: "2025-08-26T21:30:26.717Z",
	Version: types.Version{
		Major: 13,
		Minor: 45,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens:   []*types.Token{},
}

var uniswapTokenListInvalidTokens = types.TokenList{
	Name:      "Uniswap Labs Default",
	Timestamp: "2025-08-26T21:30:26.717Z",
	Version: types.Version{
		Major: 13,
		Minor: 45,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens: []*types.Token{
		{
			ChainID:  1,
			Address:  common.HexToAddress("0x744d70FDBE2Ba4CF95131626614a1763DF805B9E"),
			Name:     "Status",
			Symbol:   "SNT",
			Decimals: 18,
			LogoURI:  "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
	},
}

var uniswapTokenList = types.TokenList{
	Name:      "Uniswap Labs Default",
	Timestamp: "2025-08-26T21:30:26.717Z",
	Version: types.Version{
		Major: 13,
		Minor: 45,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens: []*types.Token{
		{
			ChainID:  1,
			Address:  common.HexToAddress("0x111111111117dC0aa78b770fA6A738034120C302"),
			Name:     "1inch",
			Symbol:   "1INCH",
			Decimals: 18,
			LogoURI:  "https://assets.coingecko.com/coins/images/13469/thumb/1inch-token.png?1608803028",
		},
		{
			ChainID:  1,
			Address:  common.HexToAddress("0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9"),
			Name:     "Aave",
			Symbol:   "AAVE",
			Decimals: 18,
			LogoURI:  "https://assets.coingecko.com/coins/images/12645/thumb/AAVE.png?1601374110",
		},
		{
			ChainID:  1,
			Address:  common.HexToAddress("0x744d70FDBE2Ba4CF95131626614a1763DF805B9E"),
			Name:     "Status",
			Symbol:   "SNT",
			Decimals: 18,
			LogoURI:  "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			ChainID:  10,
			Address:  common.HexToAddress("0x650AF3C15AF43dcB218406d30784416D64Cfb6B2"),
			Name:     "Status",
			Symbol:   "SNT",
			Decimals: 18,
			LogoURI:  "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
		{
			ChainID:  8453,
			Address:  common.HexToAddress("0x662015EC830DF08C0FC45896FaB726542e8AC09E"),
			Name:     "Status",
			Symbol:   "SNT",
			Decimals: 18,
			LogoURI:  "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		},
	},
}
