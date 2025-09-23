package parsers_test

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

// #nosec G101
const coingeckoTokensJsonResponse = `[
  {
		"id": "usd-coin",
		"symbol": "usdc",
		"name": "USDC",
		"platforms": {
			"ethereum": "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
			"arbitrum-one": "0xaf88d065e77c8cc2239327c5edb3a432268e5831",
			"optimistic-ethereum": "0x0b2c639c533813f4aa9d7837caf62653d097ff85",
			"base": "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
			"avalanche": "0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e",
			"algorand": "31566704",
			"stellar": "USDC-GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
			"celo": "0xceba9300f2b948710d2653dd7b07f33a8b32118c",
			"sui": "0xdba34672e30cb065b1f93e3ab55318768fd6fef66c15942c9f7cb846e2f900e7::usdc::USDC",
			"polygon-pos": "0x3c499c542cef5e3811e1192ce70d8cc03d5c3359",
			"solana": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
		}
	},
	{
		"id": "wrapped-bitcoin",
		"symbol": "wbtc",
		"name": "Wrapped Bitcoin",
		"platforms": {
			"ethereum": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
			"osmosis": "factory/osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743/wbtc",
			"solana": "5XZw2LKTyrfvfiskJ78AMpackRjPcyCif1WhUsPDuVqQ"
		}
	}
]`

// #nosec G101
const coingeckoTokensJsonResponseInvalidTokens = `[
  {
		"id": "usd-coin",
		"symbol": "usdc",
		"name": "USDC",
		"platforms": {
			"ethereum": "invalid-address",
			"arbitrum-one": "invalid-address",
			"optimistic-ethereum": "0x0b2c639c533813f4aa9d7837caf62653d097ff85",
			"base": "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
			"avalanche": "0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e",
			"algorand": "invalid-address",
			"stellar": "USDC-GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
			"celo": "0xceba9300f2b948710d2653dd7b07f33a8b32118c",
			"sui": "0xdba34672e30cb065b1f93e3ab55318768fd6fef66c15942c9f7cb846e2f900e7::usdc::USDC",
			"polygon-pos": "0x3c499c542cef5e3811e1192ce70d8cc03d5c3359",
			"solana": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
		}
	},
	{
		"id": "wrapped-bitcoin",
		"symbol": "wbtc",
		"name": "Wrapped Bitcoin",
		"platforms": {
			"ethereum": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
			"osmosis": "factory/osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743/wbtc",
			"solana": "5XZw2LKTyrfvfiskJ78AMpackRjPcyCif1WhUsPDuVqQ"
		}
	}
]`

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
