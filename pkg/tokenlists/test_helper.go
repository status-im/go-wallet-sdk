package tokenlists

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

func copyRemoteListOfTokenLists(original remoteListOfTokenLists) remoteListOfTokenLists {
	copy := remoteListOfTokenLists{
		Timestamp:  original.Timestamp,
		Version:    original.Version,
		TokenLists: make([]tokenList, len(original.TokenLists)),
	}

	for i, tokenList := range original.TokenLists {
		copy.TokenLists[i] = tokenList
	}

	return copy
}

func createFetchedTokenListFromResponse(response string) fetchedTokenList {
	var list fetchedTokenList
	err := json.Unmarshal([]byte(response), &list)
	if err != nil {
		panic(err)
	}
	list.Fetched = time.Now()
	return list
}

func GetTestServer() (server *httptest.Server, close func()) {
	mux := http.NewServeMux()
	server = httptest.NewServer(mux)

	mux.HandleFunc(emptyTokenListsURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(emptyTokenListsResponse)); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(delayedResponseURL, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("delayed-response")); err != nil {
			log.Println(err.Error())
		}
	})

	mux.HandleFunc(listOfTokenListsWrongSchemaURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := strings.ReplaceAll(listOfTokenListsWrongSchemaResponse, serverURLPlaceholder, server.URL)
		if _, err := w.Write([]byte(resp)); err != nil {
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

	mux.HandleFunc(statusURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(statusTokenListJsonResponse)); err != nil {
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
