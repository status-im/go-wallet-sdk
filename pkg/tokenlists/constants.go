package tokenlists

const (
	StatusListOfTokenListsID = "status-list-of-token-lists" // #nosec G101

	StatusListID             = "status"
	UniswapListID            = "uniswap"
	CoingeckoAllTokensListID = "coingecko-all-tokens"
	CoingeckoEthereumListID  = "coingecko-ethereum"
	CoingeckoOptimismListID  = "coingecko-optimism"
	CoingeckoArbitrumListID  = "coingecko-arbitrum"
	CoingeckoBSCListID       = "coingecko-bsc"
	CoingeckoBaseListID      = "coingecko-base"

	LocalSourceURL = "local"

	CustomTokenListID    = "custom"
	CommunityTokenListID = "community"

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
