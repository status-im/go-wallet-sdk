package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	log.Printf("Starting Balance Fetcher Web Server...")

	r := mux.NewRouter()

	// Add logging middleware
	r.Use(loggingMiddleware)

	// Serve static files
	r.HandleFunc("/", handleHome)
	r.HandleFunc("/fetch", handleFetchBalances)
	r.HandleFunc("/api/chains", handleGetChains)
	r.HandleFunc("/api/tokensearch", handleSearchTokens)
	r.HandleFunc("/api/tokens/{chainID}", handleGetTokens)
	r.HandleFunc("/api/tokenlist/info", handleGetTokenListInfo)

	// Start server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Printf("Access the web interface at: http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, r))
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Request completed: %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
