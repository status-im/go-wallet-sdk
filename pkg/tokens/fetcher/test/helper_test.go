package fetcher_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

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
