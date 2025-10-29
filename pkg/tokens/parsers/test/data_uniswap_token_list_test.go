package parsers_test

import (
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

// #nosec G101
const uniswapInvalidTokensJsonResponse = `[
	{
		"chainId": 1,
		"address": "invalid-address",
		"name": "1inch",
		"symbol": "1INCH",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/13469/thumb/1inch-token.png?1608803028"
	},
  {
		"chainId": 1,
		"address": "0x744d70FDBE2Ba4CF95131626614a1763DF805B9E",
		"name": "Status",
		"symbol": "SNT",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778"
	}
]`

// #nosec G101
const uniswapTokensJsonResponse = `[
	{
		"chainId": 1,
		"address": "0x111111111117dC0aa78b770fA6A738034120C302",
		"name": "1inch",
		"symbol": "1INCH",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/13469/thumb/1inch-token.png?1608803028",
		"extensions": {
			"bridgeInfo": {
				"10": {
					"tokenAddress": "0xAd42D013ac31486B73b6b059e748172994736426"
				},
				"56": {
					"tokenAddress": "0x111111111117dC0aa78b770fA6A738034120C302"
				},
				"130": {
					"tokenAddress": "0xbe41cde1C5e75a7b6c2c70466629878aa9ACd06E"
				},
				"137": {
					"tokenAddress": "0x9c2C5fd7b07E95EE044DDeba0E97a665F142394f"
				},
				"8453": {
					"tokenAddress": "0xc5fecC3a29Fb57B5024eEc8a2239d4621e111CBE"
				},
				"42161": {
					"tokenAddress": "0x6314C31A7a1652cE482cffe247E9CB7c3f4BB9aF"
				},
				"43114": {
					"tokenAddress": "0xd501281565bf7789224523144Fe5D98e8B28f267"
				}
			}
		}
	},
	{
		"chainId": 1,
		"address": "0x744d70FDBE2Ba4CF95131626614a1763DF805B9E",
		"name": "Status",
		"symbol": "SNT",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		"extensions": {
			"bridgeInfo": {
				"10": {
					"tokenAddress": "0x650AF3C15AF43dcB218406d30784416D64Cfb6B2"
				},
				"130": {
					"tokenAddress": "0x914f7CE2B080B2186159C2213B1e193E265aBF5F"
				},
				"8453": {
					"tokenAddress": "0x662015EC830DF08C0FC45896FaB726542e8AC09E"
				},
				"42161": {
					"tokenAddress": "0x707F635951193dDaFBB40971a0fCAAb8A6415160"
				}
			}
		}
	},
	{
		"chainId": 10,
		"address": "0x650AF3C15AF43dcB218406d30784416D64Cfb6B2",
		"name": "Status",
		"symbol": "SNT",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		"extensions": {
			"bridgeInfo": {
				"1": {
					"tokenAddress": "0x744d70FDBE2Ba4CF95131626614a1763DF805B9E"
				}
			}
		}
	},
	{
		"chainId": 8453,
		"address": "0x662015EC830DF08C0FC45896FaB726542e8AC09E",
		"name": "Status",
		"symbol": "SNT",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		"extensions": {
			"bridgeInfo": {
				"1": {
					"tokenAddress": "0x744d70FDBE2Ba4CF95131626614a1763DF805B9E"
				}
			}
		}
	},
	{
		"chainId": 1,
		"address": "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9",
		"name": "Aave",
		"symbol": "AAVE",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/12645/thumb/AAVE.png?1601374110",
		"extensions": {
			"bridgeInfo": {
				"10": {
					"tokenAddress": "0x76FB31fb4af56892A25e32cFC43De717950c9278"
				},
				"56": {
					"tokenAddress": "0xfb6115445Bff7b52FeB98650C87f44907E58f802"
				},
				"130": {
					"tokenAddress": "0x02a24C380dA560E4032Dc6671d8164cfbEEAAE1e"
				},
				"137": {
					"tokenAddress": "0xD6DF932A45C0f255f85145f286eA0b292B21C90B"
				},
				"8453": {
					"tokenAddress": "0x63706e401c06ac8513145b7687A14804d17f814b"
				},
				"42161": {
					"tokenAddress": "0xba5DdD1f9d7F570dc94a51479a000E3BCE967196"
				},
				"43114": {
					"tokenAddress": "0x63a72806098Bd3D9520cC43356dD78afe5D386D9"
				}
			}
		}
	}
]`

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
