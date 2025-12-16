package fetcher_test

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

const (
	serverURLPlaceholder = "SERVER-URL"

	wrongSchemaURL = "/wrong-schema.json" // #nosec G101

	delayedResponseURL = "/delayed-response.json"

	listOfTokenListsEtag             = "lotlEtag"
	listOfTokenListsNewEtag          = "lotlNewEtag"
	listOfTokenListsURL              = "/list-of-token-lists.json"                 // #nosec G101
	listOfTokenListsSomeWrongUrlsURL = "/list-of-token-lists-some-wrong-urls.json" // #nosec G101
	listOfTokenListsWithEtagURL      = "/list-of-token-lists-with-etag.json"       // #nosec G101
	listOfTokenListsWithSameEtagURL  = "/list-of-token-lists-with-same-etag.json"  // #nosec G101
	listOfTokenListsWithNewEtagURL   = "/list-of-token-lists-with-new-etag.json"   // #nosec G101

	uniswapSchemaURL = "/uniswap.schema.json"

	uniswapEtag    = "uniswapEtag"
	uniswapNewEtag = "uniswapNewEtag"
	uniswapURL     = "/uniswap.json"

	uniswapWithEtagURL = "/uniswap-with-etag.json"      // #nosec G101
	uniswapSameEtagURL = "/uniswap-with-same-etag.json" // #nosec G101
	uniswapNewEtagURL  = "/uniswap-with-new-etag.json"  // #nosec G101
)

//go:embed test/list_of_token_lists_wrong_schema.json
var wrongSchemaResponse string

//go:embed test/list_of_token_lists_response_template.json
var listOfTokenListsJsonResponseTemplate string

//go:embed test/token_lists_response.json
var tokenListsJsonResponse string

//go:embed test/token_lists_response_1.json
var tokenListsJsonResponse1 string

//go:embed test/token_lists_response_2.json
var tokenListsJsonResponse2 string

//go:embed test/list_of_token_lists_some_wrong_urls_response.json
var listOfTokenListsSomeWrongUrlsResponse string

//go:embed test/uniswap_token_list_response_template.json
var uniswapTokenListJsonResponseTemplate string

//go:embed test/uniswap_tokens_response.json
var uniswapTokensJsonResponse string

//go:embed test/uniswap_tokens_response_1.json
var uniswapTokensJsonResponse1 string

//go:embed test/uniswap_tokens_response_2.json
var uniswapTokensJsonResponse2 string

//go:embed test/uniswap_token_list_schema_response.json
var uniswapTokenListSchemaResponse string

func GetTestServer() (server *httptest.Server, close func()) {
	mux := http.NewServeMux()
	server = httptest.NewServer(mux)

	mux.HandleFunc(delayedResponseURL, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("delayed-response")); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(wrongSchemaURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(wrongSchemaResponse)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(listOfTokenListsURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := strings.ReplaceAll(listOfTokenListsJsonResponse, serverURLPlaceholder, server.URL)
		if _, err := w.Write([]byte(resp)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(listOfTokenListsSomeWrongUrlsURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := strings.ReplaceAll(listOfTokenListsWrongUrlsJsonResponse, serverURLPlaceholder, server.URL)
		if _, err := w.Write([]byte(resp)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(listOfTokenListsWithEtagURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", listOfTokenListsEtag)
		w.WriteHeader(http.StatusOK)
		resp := strings.ReplaceAll(listOfTokenListsJsonResponse1, serverURLPlaceholder, server.URL)
		if _, err := w.Write([]byte(resp)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(listOfTokenListsWithSameEtagURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", listOfTokenListsEtag)
		w.WriteHeader(http.StatusNotModified)
		if _, err := w.Write([]byte(listOfTokenListsJsonResponse1)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(listOfTokenListsWithNewEtagURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", listOfTokenListsNewEtag)
		w.WriteHeader(http.StatusOK)
		resp := strings.ReplaceAll(listOfTokenListsJsonResponse2, serverURLPlaceholder, server.URL)
		if _, err := w.Write([]byte(resp)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(uniswapSchemaURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(uniswapTokenListSchemaResponse)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(uniswapURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(uniswapTokenListJsonResponse)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(uniswapWithEtagURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", uniswapEtag)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(uniswapTokenListJsonResponse1)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(uniswapSameEtagURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", uniswapEtag)
		w.WriteHeader(http.StatusNotModified)
		if _, err := w.Write([]byte(uniswapTokenListJsonResponse1)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(uniswapNewEtagURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", uniswapNewEtag)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(uniswapTokenListJsonResponse2)); err != nil {
			log.Println(err.Error())
		}
	})

	return server, server.Close
}

func createListOfTokenListsJsonResponse(timestamp string, minor int, tokenLists string) string {
	list := strings.ReplaceAll(listOfTokenListsJsonResponseTemplate, "TIMESTAMP", timestamp)
	list = strings.ReplaceAll(list, "MINOR", fmt.Sprintf("%d", minor))
	return strings.ReplaceAll(list, "TOKEN_LISTS", tokenLists)
}

var listOfTokenListsJsonResponse = createListOfTokenListsJsonResponse("2025-09-01T00:00:00.000Z", 1, tokenListsJsonResponse)
var listOfTokenListsJsonResponse1 = createListOfTokenListsJsonResponse("2025-09-02T00:00:00.000Z", 2, tokenListsJsonResponse1)
var listOfTokenListsJsonResponse2 = createListOfTokenListsJsonResponse("2025-09-03T00:00:00.000Z", 3, tokenListsJsonResponse2)

var listOfTokenListsWrongUrlsJsonResponse = createListOfTokenListsJsonResponse("2025-09-01T00:00:00.000Z", 4, listOfTokenListsSomeWrongUrlsResponse)

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
