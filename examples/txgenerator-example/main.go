package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	log.Printf("Starting Transaction Generator Web Server...")

	r := mux.NewRouter()

	// Add logging middleware
	r.Use(loggingMiddleware)

	// Serve static files
	r.HandleFunc("/", handleHome)
	r.HandleFunc("/generate", handleGenerateTransaction)

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

// handleGenerateTransaction handles the POST /generate endpoint
func handleGenerateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate the transaction
	tx, err := GenerateTransaction(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Convert transaction to JSON
	txJSON, err := TransactionToJSON(tx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txJSON)
}
