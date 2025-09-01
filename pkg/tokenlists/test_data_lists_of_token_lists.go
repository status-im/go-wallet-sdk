package tokenlists

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	serverURLPlaceholder = "SERVER-URL"

	listOfTokenListsWrongSchemaURL = "/list-of-token-lists-with-wrong-schema.json" // #nosec G101

	emptyTokenListsURL = "/empty.json"

	delayedResponseURL = "/delayed-response.json"

	listOfTokenListsEtag             = "lotlEtag"
	listOfTokenListsNewEtag          = "lotlNewEtag"
	listOfTokenListsURL              = "/list-of-token-lists.json"                 // #nosec G101
	listOfTokenListsSomeWrongUrlsURL = "/list-of-token-lists-some-wrong-urls.json" // #nosec G101
	listOfTokenListsWithEtagURL      = "/list-of-token-lists-with-etag.json"       // #nosec G101
	listOfTokenListsWithSameEtagURL  = "/list-of-token-lists-with-same-etag.json"  // #nosec G101
	listOfTokenListsWithNewEtagURL   = "/list-of-token-lists-with-new-etag.json"   // #nosec G101
)

// #nosec G101
const emptyTokenListsResponse = `{
	"timestamp": "2025-09-01T00:00:00.000Z",
	"version": {
		"major": 0,
		"minor": 1,
		"patch": 0
	}
}`

// #nosec G101
const listOfTokenListsWrongSchemaResponse = `{
	"timestamp": "2025-09-01T00:00:00.000Z",
	{
		"id": "uniswap"
	}
}`

// #nosec G101
const listOfTokenListsJsonResponseTemplate = `{
  "timestamp": "TIMESTAMP",
  "version": {
    "major": 0,
    "minor": MINOR,
    "patch": 0
  },
  "tokenLists": TOKEN_LISTS
}`

// #nosec G101
const tokenListsJsonResponse = `[
	{
		"id": "status",
		"sourceUrl": "SERVER-URL/status-token-list.json"
	},
	{
		"id": "uniswap",
		"sourceUrl": "SERVER-URL/uniswap.json"
	}
]`

// #nosec G101
const tokenListsJsonResponse1 = `[
	{
		"id": "status",
		"sourceUrl": "SERVER-URL/status-token-list.json"
	},
	{
		"id": "uniswap",
		"sourceUrl": "SERVER-URL/uniswap.json"
	},
	{
		"id": "coingecko",
		"sourceUrl": "SERVER-URL/coingecko.json"
	}
]`

// #nosec G101
const tokenListsJsonResponse2 = `[
	{
		"id": "status",
		"sourceUrl": "SERVER-URL/status-token-list.json"
	},
	{
		"id": "uniswap",
		"sourceUrl": "SERVER-URL/uniswap.json"
	},
	{
		"id": "coingecko",
		"sourceUrl": "SERVER-URL/coingecko.json"
	},
	{
		"id": "aave",
		"sourceUrl": "SERVER-URL/aave.json"
	}
]`

// #nosec G101
const listOfTokenListsSomeWrongUrlsResponse = `[
	{
		"id": "status",
		"sourceUrl": "SERVER-URL/status-token-list.json"
	},
	{
		"id": "invalid-list",
		"sourceUrl": "SERVER-URL/invalid-url-tokens.json"
	}
]`

func createListOfTokenListsJsonResponse(timestamp string, minor int, tokenLists string) string {
	list := strings.ReplaceAll(listOfTokenListsJsonResponseTemplate, "TIMESTAMP", timestamp)
	list = strings.ReplaceAll(list, "MINOR", fmt.Sprintf("%d", minor))
	return strings.ReplaceAll(list, "TOKEN_LISTS", tokenLists)
}

func createListOfTokenListsFromResponse(response string) remoteListOfTokenLists {
	var list remoteListOfTokenLists
	err := json.Unmarshal([]byte(response), &list)
	if err != nil {
		panic(err)
	}
	return list
}

var listOfTokenListsJsonResponse = createListOfTokenListsJsonResponse("2025-09-01T00:00:00.000Z", 1, tokenListsJsonResponse)
var listOfTokenListsJsonResponse1 = createListOfTokenListsJsonResponse("2025-09-02T00:00:00.000Z", 2, tokenListsJsonResponse1)
var listOfTokenListsJsonResponse2 = createListOfTokenListsJsonResponse("2025-09-03T00:00:00.000Z", 3, tokenListsJsonResponse2)
var listOfTokenListsWrongUrlsJsonResponse = createListOfTokenListsJsonResponse("2025-09-01T00:00:00.000Z", 4, listOfTokenListsSomeWrongUrlsResponse)

var expectedListOfTokenLists = createListOfTokenListsFromResponse(listOfTokenListsJsonResponse)
var expectedListOfTokenLists1 = createListOfTokenListsFromResponse(listOfTokenListsJsonResponse1)
var expectedListOfTokenLists2 = createListOfTokenListsFromResponse(listOfTokenListsJsonResponse2)
