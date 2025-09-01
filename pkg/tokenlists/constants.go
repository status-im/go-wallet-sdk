package tokenlists

const (
	EthereumNativeCrossChainID = "eth-native"
	EthereumNativeSymbol       = "ETH"
	EthereumNativeName         = "Ethereum"

	BinanceSmartChainNativeCrossChainID = "bsc-native"
	BinanceSmartChainNativeSymbol       = "BNB"
	BinanceSmartChainNativeName         = "BNB"

	StatusListOfTokenListsID = "status-list-of-token-lists" // #nosec G101

	NativeTokenListID        = "native"
	StatusListID             = "status"
	UniswapListID            = "uniswap"
	CoingeckoAllTokensListID = "coingeckoAllTokens"
	CoingeckoEthereumListID  = "coingeckoEthereum"
	CoingeckoOptimismListID  = "coingeckoOptimism"
	CoingeckoArbitrumListID  = "coingeckoArbitrum"
	CoingeckoBSCListID       = "coingeckoBsc"
	CoingeckoBaseListID      = "coingeckoBase"

	LocalSourceURL = "local"

	CustomTokenListID = "custom"

	// #nosec G101
	listOfTokenListsSchema = `{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "timestamp": {
      "type": "string",
      "description": "The timestamp of this list version",
      "format": "date-time",
      "additionalProperties": false
    },
    "version": {
      "type": "object",
      "description": "The version of the list, used in change detection",
      "properties": {
        "major": {
          "type": "integer"
        },
        "minor": {
          "type": "integer"
        },
        "patch": {
          "type": "integer"
        }
      },
      "required": [
        "major",
        "minor",
        "patch"
      ],
      "additionalProperties": false
    },
    "tokenLists": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "A unique identifier for the token list source."
          },
          "sourceUrl": {
            "type": "string",
            "format": "uri",
            "description": "URL pointing to the token list source."
          },
          "schema": {
            "type": "string",
            "format": "uri",
            "description": "Optional URL pointing to the schema definition of the token list.",
            "nullable": true
          }
        },
        "required": [
          "id",
          "sourceUrl"
        ],
        "additionalProperties": false
      }
    }
  }
}`
)
