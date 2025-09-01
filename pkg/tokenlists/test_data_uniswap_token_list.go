package tokenlists

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	uniswapSchemaURL = "/uniswap.schema.json"

	uniswapEtag        = "uniswapEtag"
	uniswapNewEtag     = "uniswapNewEtag"
	uniswapURL         = "/uniswap.json"
	uniswapWithEtagURL = "/uniswap-with-etag.json"      // #nosec G101
	uniswapSameEtagURL = "/uniswap-with-same-etag.json" // #nosec G101
	uniswapNewEtagURL  = "/uniswap-with-new-etag.json"  // #nosec G101
)

// #nosec G101
const uniswapTokenListJsonResponseTemplate = `{
	"name": "NAME",
  "timestamp": "TIMESTAMP",
  "version": {
    "major": MAJOR,
    "minor": MINOR,
    "patch": 0
  },
	"tags": {},
  "logoURI": "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
  "keywords": [
    "uniswap",
    "default"
  ],
  "tokens": TOKENS
}`

// #nosec G101
const uniswapTokensJsonResponse = `[
	{
		"chainId": 1,
		"address": "0x744d70FDBE2Ba4CF95131626614a1763DF805B9E",
		"name": "Status",
		"symbol": "SNT",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
		"extensions": {
			"bridgeInfo": {
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
	}
]`

// #nosec G101
const uniswapTokensJsonResponse1 = `[
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
	}
]`

// #nosec G101
const uniswapTokensJsonResponse2 = `[
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
		"address": "0x3E5A19c91266aD8cE2477B91585d1856B84062dF",
		"name": "Ancient8",
		"symbol": "A8",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/39170/standard/A8_Token-04_200x200.png?1720798300",
		"extensions": {
			"bridgeInfo": {
				"130": {
					"tokenAddress": "0x44D618C366D7bC85945Bfc922ACad5B1feF7759A"
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
		"name": "USDCoin",
		"address": "0x0b2C639c533813f4Aa9D7837CAf62653d097Ff85",
		"symbol": "USDC",
		"decimals": 6,
		"chainId": 10,
		"logoURI": "https://ethereum-optimism.github.io/data/USDC/logo.png",
		"extensions": {
			"bridgeInfo": {
				"1": {
					"tokenAddress": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
				}
			}
		}
	},
	{
		"name": "USDCoin",
		"address": "0xaf88d065e77c8cC2239327C5EDb3A432268e5831",
		"symbol": "USDC",
		"decimals": 6,
		"chainId": 42161,
		"logoURI": "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		"extensions": {
			"bridgeInfo": {
				"1": {
					"tokenAddress": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
				}
			}
		}
	},
	{
		"name": "Wrapped Ether",
		"address": "0xA6FA4fB5f76172d178d61B04b0ecd319C5d1C0aa",
		"symbol": "WETH",
		"decimals": 18,
		"chainId": 80001,
		"logoURI": "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2/logo.png"
	},
	{
		"name": "Wrapped Matic",
		"address": "0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889",
		"symbol": "WMATIC",
		"decimals": 18,
		"chainId": 80001,
		"logoURI": "https://assets.coingecko.com/coins/images/4713/thumb/matic-token-icon.png?1624446912"
	},
	{
		"chainId": 81457,
		"address": "0xb1a5700fA2358173Fe465e6eA4Ff52E36e88E2ad",
		"name": "Blast",
		"symbol": "BLAST",
		"decimals": 18,
		"logoURI": "https://assets.coingecko.com/coins/images/35494/standard/Blast.jpg?1719385662"
	},
	{
		"chainId": 7777777,
		"address": "0xCccCCccc7021b32EBb4e8C08314bD62F7c653EC4",
		"name": "USD Coin (Bridged from Ethereum)",
		"symbol": "USDzC",
		"decimals": 6,
		"logoURI": "https://assets.coingecko.com/coins/images/35218/large/USDC_Icon.png?1707908537"
	},
	{
		"name": "Uniswap",
		"address": "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
		"symbol": "UNI",
		"decimals": 18,
		"chainId": 11155111,
		"logoURI": "ipfs://QmXttGpZrECX5qCyXbBQiqgQNytVGeZW5Anewvh2jc4psg"
	},
	{
		"name": "Wrapped Ether",
		"address": "0xfFf9976782d46CC05630D1f6eBAb18b2324d6B14",
		"symbol": "WETH",
		"decimals": 18,
		"chainId": 11155111,
		"logoURI": "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2/logo.png"
	}
]`

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

func createUniswapTokenListJsonResponse(name string, timestamp string, major int, minor int, tokens string) string {
	list := strings.ReplaceAll(uniswapTokenListJsonResponseTemplate, "NAME", name)
	list = strings.ReplaceAll(list, "TIMESTAMP", timestamp)
	list = strings.ReplaceAll(list, "MAJOR", fmt.Sprintf("%d", major))
	list = strings.ReplaceAll(list, "MINOR", fmt.Sprintf("%d", minor))
	return strings.ReplaceAll(list, "TOKENS", tokens)
}

var uniswapTokenListJsonResponse = createUniswapTokenListJsonResponse("Uniswap Labs Default", "2025-08-26T21:30:26.717Z", 13, 45, uniswapTokensJsonResponse)
var uniswapTokenListJsonResponse1 = createUniswapTokenListJsonResponse("Uniswap Labs Default", "2025-08-27T21:30:26.717Z", 13, 46, uniswapTokensJsonResponse1)
var uniswapTokenListJsonResponse2 = createUniswapTokenListJsonResponse("Uniswap Labs Default", "2025-08-28T21:30:26.717Z", 13, 47, uniswapTokensJsonResponse2)

var uniswapTokenListInvalidTokensJsonResponse = createStatusTokenListJsonResponse("Uniswap Labs Default", "2025-08-26T21:30:26.717Z", 13, 45, uniswapInvalidTokensJsonResponse)

var uniswapTokenListEmptyTokensJsonResponse = createStatusTokenListJsonResponse("Uniswap Labs Default", "2025-08-26T21:30:26.717Z", 13, 45, "[]")

var fetchedUniswapTokenList = createFetchedTokenListFromResponse(uniswapTokenListJsonResponse)
var fetchedUniswapTokenList1 = createFetchedTokenListFromResponse(uniswapTokenListJsonResponse1)
var fetchedUniswapTokenList2 = createFetchedTokenListFromResponse(uniswapTokenListJsonResponse2)

var fetchedUniswapTokenListInvalidTokens = createFetchedTokenListFromResponse(uniswapTokenListInvalidTokensJsonResponse)

var fetchedUniswapTokenListEmpty = createFetchedTokenListFromResponse(uniswapTokenListEmptyTokensJsonResponse)

var uniswapTokenListEmpty = TokenList{
	Name:             "Uniswap Labs Default",
	Timestamp:        "2025-08-26T21:30:26.717Z",
	FetchedTimestamp: fetchedUniswapTokenListEmpty.Fetched.Format(time.RFC3339),
	Source:           "https://example.com/uniswap-token-list.json",
	Version: Version{
		Major: 13,
		Minor: 45,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens:   []*Token{},
}

var uniswapTokenListInvalidTokens = TokenList{
	Name:             "Uniswap Labs Default",
	Timestamp:        "2025-08-26T21:30:26.717Z",
	FetchedTimestamp: fetchedUniswapTokenListInvalidTokens.Fetched.Format(time.RFC3339),
	Source:           "https://example.com/uniswap-token-list.json",
	Version: Version{
		Major: 13,
		Minor: 45,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens: []*Token{
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

var uniswapTokenList = TokenList{
	Name:             "Uniswap Labs Default",
	Timestamp:        "2025-08-28T21:30:26.717Z",
	FetchedTimestamp: fetchedUniswapTokenList2.Fetched.Format(time.RFC3339),
	Source:           "https://example.com/uniswap-token-list.json",
	Version: Version{
		Major: 13,
		Minor: 47,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens: []*Token{
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

var uniswapTokenList2 = TokenList{
	Name:             "Uniswap Labs Default",
	Timestamp:        "2025-08-28T21:30:26.717Z",
	FetchedTimestamp: fetchedUniswapTokenList2.Fetched.Format(time.RFC3339),
	Source:           "https://example.com/uniswap-token-list.json",
	Version: Version{
		Major: 13,
		Minor: 47,
		Patch: 0,
	},
	Tags:     map[string]interface{}{},
	LogoURI:  "ipfs://QmNa8mQkrNKp1WEEeGjFezDmDeodkWRevGFN8JCV7b4Xir",
	Keywords: []string{"uniswap", "default"},
	Tokens: []*Token{
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
			Address:  common.HexToAddress("0x3E5A19c91266aD8cE2477B91585d1856B84062dF"),
			Name:     "Ancient8",
			Symbol:   "A8",
			Decimals: 18,
			LogoURI:  "https://assets.coingecko.com/coins/images/39170/standard/A8_Token-04_200x200.png?1720798300",
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
		{
			ChainID:  10,
			Address:  common.HexToAddress("0x0b2C639c533813f4Aa9D7837CAf62653d097Ff85"),
			Name:     "USDCoin",
			Symbol:   "USDC",
			Decimals: 6,
			LogoURI:  "https://ethereum-optimism.github.io/data/USDC/logo.png",
		},
		{
			ChainID:  42161,
			Address:  common.HexToAddress("0xaf88d065e77c8cC2239327C5EDb3A432268e5831"),
			Name:     "USDCoin",
			Symbol:   "USDC",
			Decimals: 6,
			LogoURI:  "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
		},
		{
			ChainID:  11155111,
			Address:  common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"),
			Name:     "Uniswap",
			Symbol:   "UNI",
			Decimals: 18,
			LogoURI:  "ipfs://QmXttGpZrECX5qCyXbBQiqgQNytVGeZW5Anewvh2jc4psg",
		},
		{
			ChainID:  11155111,
			Address:  common.HexToAddress("0xfFf9976782d46CC05630D1f6eBAb18b2324d6B14"),
			Name:     "Wrapped Ether",
			Symbol:   "WETH",
			Decimals: 18,
			LogoURI:  "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2/logo.png",
		},
	},
}

// #nosec G101
const uniswapTokenListSchemaResponse = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://uniswap.org/tokenlist.schema.json",
  "title": "Uniswap Token List",
  "description": "Schema for lists of tokens compatible with the Uniswap Interface",
  "definitions": {
    "Version": {
      "type": "object",
      "description": "The version of the list, used in change detection",
      "examples": [
        {
          "major": 1,
          "minor": 0,
          "patch": 0
        }
      ],
      "additionalProperties": false,
      "properties": {
        "major": {
          "type": "integer",
          "description": "The major version of the list. Must be incremented when tokens are removed from the list or token addresses are changed.",
          "minimum": 0,
          "examples": [
            1,
            2
          ]
        },
        "minor": {
          "type": "integer",
          "description": "The minor version of the list. Must be incremented when tokens are added to the list.",
          "minimum": 0,
          "examples": [
            0,
            1
          ]
        },
        "patch": {
          "type": "integer",
          "description": "The patch version of the list. Must be incremented for any changes to the list.",
          "minimum": 0,
          "examples": [
            0,
            1
          ]
        }
      },
      "required": [
        "major",
        "minor",
        "patch"
      ]
    },
    "TagIdentifier": {
      "type": "string",
      "description": "The unique identifier of a tag",
      "minLength": 1,
      "maxLength": 10,
      "pattern": "^[\\w]+$",
      "examples": [
        "compound",
        "stablecoin"
      ]
    },
    "ExtensionIdentifier": {
      "type": "string",
      "description": "The name of a token extension property",
      "minLength": 1,
      "maxLength": 40,
      "pattern": "^[\\w]+$",
      "examples": [
        "color",
        "is_fee_on_transfer",
        "aliases"
      ]
    },
    "ExtensionMap": {
      "type": "object",
      "description": "An object containing any arbitrary or vendor-specific token metadata",
      "maxProperties": 10,
      "propertyNames": {
        "$ref": "#/definitions/ExtensionIdentifier"
      },
      "additionalProperties": {
        "$ref": "#/definitions/ExtensionValue"
      },
      "examples": [
        {
          "color": "#000000",
          "is_verified_by_me": true
        },
        {
          "x-bridged-addresses-by-chain": {
            "1": {
              "bridgeAddress": "0x4200000000000000000000000000000000000010",
              "tokenAddress": "0x4200000000000000000000000000000000000010"
            }
          }
        }
      ]
    },
    "ExtensionPrimitiveValue": {
      "anyOf": [
        {
          "type": "string",
          "minLength": 1,
          "maxLength": 42,
          "examples": [
            "#00000"
          ]
        },
        {
          "type": "boolean",
          "examples": [
            true
          ]
        },
        {
          "type": "number",
          "examples": [
            15
          ]
        },
        {
          "type": "null"
        }
      ]
    },
    "ExtensionValue": {
      "anyOf": [
        {
          "$ref": "#/definitions/ExtensionPrimitiveValue"
        },
        {
          "type": "object",
          "maxProperties": 10,
          "propertyNames": {
            "$ref": "#/definitions/ExtensionIdentifier"
          },
          "additionalProperties": {
            "$ref": "#/definitions/ExtensionValueInner0"
          }
        }
      ]
    },
    "ExtensionValueInner0": {
      "anyOf": [
        {
          "$ref": "#/definitions/ExtensionPrimitiveValue"
        },
        {
          "type": "object",
          "maxProperties": 10,
          "propertyNames": {
            "$ref": "#/definitions/ExtensionIdentifier"
          },
          "additionalProperties": {
            "$ref": "#/definitions/ExtensionValueInner1"
          }
        }
      ]
    },
    "ExtensionValueInner1": {
      "anyOf": [
        {
          "$ref": "#/definitions/ExtensionPrimitiveValue"
        }
      ]
    },
    "TagDefinition": {
      "type": "object",
      "description": "Definition of a tag that can be associated with a token via its identifier",
      "additionalProperties": false,
      "properties": {
        "name": {
          "type": "string",
          "description": "The name of the tag",
          "pattern": "^[ \\w]+$",
          "minLength": 1,
          "maxLength": 20
        },
        "description": {
          "type": "string",
          "description": "A user-friendly description of the tag",
          "pattern": "^[ \\w\\.,:]+$",
          "minLength": 1,
          "maxLength": 200
        }
      },
      "required": [
        "name",
        "description"
      ],
      "examples": [
        {
          "name": "Stablecoin",
          "description": "A token with value pegged to another asset"
        }
      ]
    },
    "TokenInfo": {
      "type": "object",
      "description": "Metadata for a single token in a token list",
      "additionalProperties": false,
      "properties": {
        "chainId": {
          "type": "integer",
          "description": "The chain ID of the Ethereum network where this token is deployed",
          "minimum": 1,
          "examples": [
            1,
            42
          ]
        },
        "address": {
          "type": "string",
          "description": "The checksummed address of the token on the specified chain ID",
          "pattern": "^0x[a-fA-F0-9]{40}$",
          "examples": [
            "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
          ]
        },
        "decimals": {
          "type": "integer",
          "description": "The number of decimals for the token balance",
          "minimum": 0,
          "maximum": 255,
          "examples": [
            18
          ]
        },
        "name": {
          "type": "string",
          "description": "The name of the token",
          "minLength": 0,
          "maxLength": 60,
          "anyOf": [
            {
              "const": ""
            },
            {
              "pattern": "^[ \\S+]+$"
            }
          ],
          "examples": [
            "USD Coin"
          ]
        },
        "symbol": {
          "type": "string",
          "description": "The symbol for the token",
          "minLength": 0,
          "maxLength": 20,
          "anyOf": [
            {
              "const": ""
            },
            {
              "pattern": "^\\S+$"
            }
          ],
          "examples": [
            "USDC"
          ]
        },
        "logoURI": {
          "type": "string",
          "description": "A URI to the token logo asset; if not set, interface will attempt to find a logo based on the token address; suggest SVG or PNG of size 64x64",
          "format": "uri",
          "examples": [
            "ipfs://QmXfzKRvjZz3u5JRgC4v5mGVbm9ahrUiB4DgzHBsnWbTMM"
          ]
        },
        "tags": {
          "type": "array",
          "description": "An array of tag identifiers associated with the token; tags are defined at the list level",
          "items": {
            "$ref": "#/definitions/TagIdentifier"
          },
          "maxItems": 10,
          "examples": [
            "stablecoin",
            "compound"
          ]
        },
        "extensions": {
          "$ref": "#/definitions/ExtensionMap"
        }
      },
      "required": [
        "chainId",
        "address",
        "decimals",
        "name",
        "symbol"
      ]
    }
  },
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "The name of the token list",
      "minLength": 1,
      "maxLength": 30,
      "pattern": "^[\\w ]+$",
      "examples": [
        "My Token List"
      ]
    },
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "The timestamp of this list version; i.e. when this immutable version of the list was created"
    },
    "version": {
      "$ref": "#/definitions/Version"
    },
    "tokens": {
      "type": "array",
      "description": "The list of tokens included in the list",
      "items": {
        "$ref": "#/definitions/TokenInfo"
      },
      "minItems": 1,
      "maxItems": 10000
    },
    "tokenMap": {
      "type": "object",
      "description": "A mapping of key 'chainId_tokenAddress' to its corresponding token object",
      "minProperties": 1,
      "maxProperties": 10000,
      "propertyNames": {
        "type": "string"
      },
      "additionalProperties": {
        "$ref": "#/definitions/TokenInfo"
      },
      "examples": [
        {
          "4_0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984":  {
            "name": "Uniswap",
            "address": "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
            "symbol": "UNI",
            "decimals": 18,
            "chainId": 4,
            "logoURI": "ipfs://QmXttGpZrECX5qCyXbBQiqgQNytVGeZW5Anewvh2jc4psg"
          }
        }
      ]
    },
    "keywords": {
      "type": "array",
      "description": "Keywords associated with the contents of the list; may be used in list discoverability",
      "items": {
        "type": "string",
        "description": "A keyword to describe the contents of the list",
        "minLength": 1,
        "maxLength": 20,
        "pattern": "^[\\w ]+$",
        "examples": [
          "compound",
          "lending",
          "personal tokens"
        ]
      },
      "maxItems": 20,
      "uniqueItems": true
    },
    "tags": {
      "type": "object",
      "description": "A mapping of tag identifiers to their name and description",
      "propertyNames": {
        "$ref": "#/definitions/TagIdentifier"
      },
      "additionalProperties": {
        "$ref": "#/definitions/TagDefinition"
      },
      "maxProperties": 20,
      "examples": [
        {
          "stablecoin": {
            "name": "Stablecoin",
            "description": "A token with value pegged to another asset"
          }
        }
      ]
    },
    "logoURI": {
      "type": "string",
      "description": "A URI for the logo of the token list; prefer SVG or PNG of size 256x256",
      "format": "uri",
      "examples": [
        "ipfs://QmXfzKRvjZz3u5JRgC4v5mGVbm9ahrUiB4DgzHBsnWbTMM"
      ]
    }
  },
  "required": [
    "name",
    "timestamp",
    "version",
    "tokens"
  ]
}`
