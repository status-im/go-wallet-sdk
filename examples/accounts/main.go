package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/status-im/go-wallet-sdk/pkg/accounts/extkeystore"
	"github.com/status-im/go-wallet-sdk/pkg/accounts/mnemonic"
)

var extKeyStore *extkeystore.KeyStore
var stdKeyStore *keystore.KeyStore

func main() {
	log.Printf("Starting Accounts Example Web Server...")

	// Create temporary directories for keystores
	extKeyDir := filepath.Join(os.TempDir(), "accounts-example-extkeystore")
	stdKeyDir := filepath.Join(os.TempDir(), "accounts-example-stdkeystore")

	if err := os.MkdirAll(extKeyDir, 0700); err != nil {
		log.Fatalf("Failed to create extkeystore directory: %v", err)
	}
	if err := os.MkdirAll(stdKeyDir, 0700); err != nil {
		log.Fatalf("Failed to create stdkeystore directory: %v", err)
	}

	// Initialize extended keystore with light scrypt parameters for faster operations
	extKeyStore = extkeystore.NewKeyStore(extKeyDir, extkeystore.LightScryptN, extkeystore.LightScryptP)

	// Initialize standard keystore with light scrypt parameters
	stdKeyStore = keystore.NewKeyStore(stdKeyDir, keystore.LightScryptN, keystore.LightScryptP)

	r := mux.NewRouter()

	// Add logging middleware
	r.Use(loggingMiddleware)

	// Serve static files and main page
	r.HandleFunc("/", handleHome)

	// API endpoints - Extended Keystore
	r.HandleFunc("/api/ext/generate-mnemonic", handleGenerateMnemonic).Methods("POST")
	r.HandleFunc("/api/ext/create-account", handleExtCreateAccount).Methods("POST")
	r.HandleFunc("/api/ext/derive-account", handleExtDeriveAccount).Methods("POST")
	r.HandleFunc("/api/ext/import-key", handleExtImportKey).Methods("POST")
	r.HandleFunc("/api/ext/export-key", handleExtExportKey).Methods("POST")
	r.HandleFunc("/api/ext/export-priv", handleExtExportPriv).Methods("POST")
	r.HandleFunc("/api/ext/sign-message", handleExtSignMessage).Methods("POST")
	r.HandleFunc("/api/ext/unlock-account", handleExtUnlockAccount).Methods("POST")
	r.HandleFunc("/api/ext/lock-account", handleExtLockAccount).Methods("POST")
	r.HandleFunc("/api/ext/delete-account", handleExtDeleteAccount).Methods("POST")
	r.HandleFunc("/api/ext/accounts", handleExtGetAccounts).Methods("GET")
	r.HandleFunc("/api/ext/account/{address}", handleExtGetAccountInfo).Methods("GET")

	// API endpoints - Standard Keystore
	r.HandleFunc("/api/std/create-account", handleStdCreateAccount).Methods("POST")
	r.HandleFunc("/api/std/import-key", handleStdImportKey).Methods("POST")
	r.HandleFunc("/api/std/export-key", handleStdExportKey).Methods("POST")
	r.HandleFunc("/api/std/sign-message", handleStdSignMessage).Methods("POST")
	r.HandleFunc("/api/std/unlock-account", handleStdUnlockAccount).Methods("POST")
	r.HandleFunc("/api/std/lock-account", handleStdLockAccount).Methods("POST")
	r.HandleFunc("/api/std/delete-account", handleStdDeleteAccount).Methods("POST")
	r.HandleFunc("/api/std/accounts", handleStdGetAccounts).Methods("GET")
	r.HandleFunc("/api/std/account/{address}", handleStdGetAccountInfo).Methods("GET")

	// Start server
	port := ":8081"
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

// sendJSONError sends a JSON error response
func sendJSONError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorMsg := message
	if err != nil {
		if errorMsg != "" {
			errorMsg = fmt.Sprintf("%s: %v", message, err)
		} else {
			errorMsg = err.Error()
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": errorMsg,
	})
}

// handleGenerateMnemonic generates a random mnemonic phrase
func handleGenerateMnemonic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Length int `json:"length"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if req.Length == 0 {
		req.Length = 12
	}

	phrase, err := mnemonic.CreateRandomMnemonic(req.Length)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to generate mnemonic", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"mnemonic": phrase,
		"length":   req.Length,
	})
}

// handleExtCreateAccount creates a new account from a mnemonic phrase in extended keystore
func handleExtCreateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Mnemonic   string `json:"mnemonic"`
		Passphrase string `json:"passphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if req.Mnemonic == "" {
		sendJSONError(w, http.StatusBadRequest, "mnemonic is required", nil)
		return
	}

	// Create extended key from mnemonic
	extKey, err := mnemonic.CreateExtendedKeyFromMnemonic(req.Mnemonic, "")
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to create extended key", err)
		return
	}

	// Import the extended key into keystore
	account, err := extKeyStore.ImportExtendedKey(extKey, req.Passphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to import key", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": account.Address.Hex(),
		"url":     account.URL.String(),
	})
}

// handleExtDeriveAccount derives a child account from a parent in extended keystore
func handleExtDeriveAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address       string `json:"address"`
		Path          string `json:"path"`
		Passphrase    string `json:"passphrase"`    // Passphrase for the parent account
		NewPassphrase string `json:"newPassphrase"` // Passphrase for the derived account (when pinned)
		Pin           bool   `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}

	// Parse derivation path
	derivationPath, err := accounts.ParseDerivationPath(req.Path)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid derivation path", err)
		return
	}

	// If newPassphrase is empty and pin is true, use the same passphrase as the parent
	newPassphrase := req.NewPassphrase
	if newPassphrase == "" && req.Pin {
		newPassphrase = req.Passphrase
	}

	// Derive account
	derivedAccount, err := extKeyStore.DeriveWithPassphrase(account, derivationPath, req.Pin, req.Passphrase, newPassphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to derive account", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": derivedAccount.Address.Hex(),
		"url":     derivedAccount.URL.String(),
		"pinned":  req.Pin,
	})
}

// handleExtImportKey imports a JSON key file into extended keystore
func handleExtImportKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		KeyJSON       string `json:"keyJson"`
		Passphrase    string `json:"passphrase"`
		NewPassphrase string `json:"newPassphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	keyJSONBytes := []byte(req.KeyJSON)
	account, err := extKeyStore.Import(keyJSONBytes, req.Passphrase, req.NewPassphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to import key", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": account.Address.Hex(),
		"url":     account.URL.String(),
	})
}

// handleExtExportKey exports a key as JSON from extended keystore
func handleExtExportKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address       string `json:"address"`
		Passphrase    string `json:"passphrase"`
		NewPassphrase string `json:"newPassphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}
	keyJSON, err := extKeyStore.ExportExt(account, req.Passphrase, req.NewPassphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to export key", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"keyJson": string(keyJSON),
	})
}

// handleExtExportPriv exports a key as private key JSON (compatible with standard keystore) from extended keystore
func handleExtExportPriv(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address       string `json:"address"`
		Passphrase    string `json:"passphrase"`
		NewPassphrase string `json:"newPassphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}
	keyJSON, err := extKeyStore.ExportPriv(account, req.Passphrase, req.NewPassphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to export private key", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"keyJson": string(keyJSON),
	})
}

// handleExtSignMessage signs a message with an account in extended keystore
func handleExtSignMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address    string `json:"address"`
		Message    string `json:"message"`
		Passphrase string `json:"passphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}

	// Hash the message using Ethereum message prefix
	messageHash := accounts.TextHash([]byte(req.Message))

	// Sign the hash
	var signature []byte
	var err error
	if req.Passphrase != "" {
		signature, err = extKeyStore.SignHashWithPassphrase(account, req.Passphrase, messageHash)
	} else {
		signature, err = extKeyStore.SignHash(account, messageHash)
	}

	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to sign message", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"signature":   hex.EncodeToString(signature),
		"message":     req.Message,
		"messageHash": hex.EncodeToString(messageHash),
	})
}

// handleExtUnlockAccount unlocks an account in extended keystore
func handleExtUnlockAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address    string `json:"address"`
		Passphrase string `json:"passphrase"`
		Timeout    int    `json:"timeout"` // in seconds, 0 for indefinite
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}
	timeout := time.Duration(req.Timeout) * time.Second

	err := extKeyStore.TimedUnlock(account, req.Passphrase, timeout)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to unlock account", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"unlocked": true,
		"timeout":  req.Timeout,
	})
}

// handleExtLockAccount locks an account in extended keystore
func handleExtLockAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	err := extKeyStore.Lock(common.HexToAddress(req.Address))
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to lock account", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"locked": true,
	})
}

// handleExtGetAccounts returns all accounts in the extended keystore
func handleExtGetAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountsList := extKeyStore.Accounts()
	accountsData := make([]map[string]interface{}, len(accountsList))

	for i, acc := range accountsList {
		accountsData[i] = map[string]interface{}{
			"address": acc.Address.Hex(),
			"url":     acc.URL.String(),
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accountsData,
		"count":    len(accountsData),
	})
}

// handleExtGetAccountInfo returns detailed information about an account in extended keystore
func handleExtGetAccountInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	addressStr := vars["address"]

	if !common.IsHexAddress(addressStr) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	address := common.HexToAddress(addressStr)
	account := accounts.Account{Address: address}

	// Find the account
	foundAccount, err := extKeyStore.Find(account)
	if err != nil {
		sendJSONError(w, http.StatusNotFound, "account not found", err)
		return
	}

	// Check if account is unlocked by trying to get wallet status
	isUnlocked := false
	wallets := extKeyStore.Wallets()
	for _, wallet := range wallets {
		if wallet.Contains(foundAccount) {
			status, err := wallet.Status()
			if err == nil && status == "Unlocked" {
				isUnlocked = true
			}
			break
		}
	}

	// Read the keystore file contents
	filePath := foundAccount.URL.Path
	var fileContents string
	if filePath != "" {
		contents, err := os.ReadFile(filePath)
		if err != nil {
			// If file can't be read, still return other info but indicate error
			fileContents = fmt.Sprintf("Error reading file: %v", err)
		} else {
			fileContents = string(contents)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address":      foundAccount.Address.Hex(),
		"url":          foundAccount.URL.String(),
		"filePath":     filePath,
		"fileContents": fileContents,
		"unlocked":     isUnlocked,
		"hasAddress":   extKeyStore.HasAddress(address),
	})
}

// handleExtDeleteAccount deletes an account from extended keystore
func handleExtDeleteAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address    string `json:"address"`
		Passphrase string `json:"passphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}
	err := extKeyStore.Delete(account, req.Passphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to delete account", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": true,
		"address": req.Address,
	})
}

// Standard Keystore Handlers

// handleStdCreateAccount creates a new account in standard keystore
func handleStdCreateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Passphrase string `json:"passphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	account, err := stdKeyStore.NewAccount(req.Passphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to create account", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": account.Address.Hex(),
		"url":     account.URL.String(),
	})
}

// handleStdImportKey imports a JSON key file into standard keystore
func handleStdImportKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		KeyJSON       string `json:"keyJson"`
		Passphrase    string `json:"passphrase"`
		NewPassphrase string `json:"newPassphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	keyJSONBytes := []byte(req.KeyJSON)
	account, err := stdKeyStore.Import(keyJSONBytes, req.Passphrase, req.NewPassphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to import key", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": account.Address.Hex(),
		"url":     account.URL.String(),
	})
}

// handleStdExportKey exports a key as JSON from standard keystore
func handleStdExportKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address       string `json:"address"`
		Passphrase    string `json:"passphrase"`
		NewPassphrase string `json:"newPassphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}
	keyJSON, err := stdKeyStore.Export(account, req.Passphrase, req.NewPassphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to export key", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"keyJson": string(keyJSON),
	})
}

// handleStdSignMessage signs a message with an account in standard keystore
func handleStdSignMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address    string `json:"address"`
		Message    string `json:"message"`
		Passphrase string `json:"passphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}

	// Hash the message using Ethereum message prefix
	messageHash := accounts.TextHash([]byte(req.Message))

	// Sign the hash
	var signature []byte
	var err error
	if req.Passphrase != "" {
		signature, err = stdKeyStore.SignHashWithPassphrase(account, req.Passphrase, messageHash)
	} else {
		signature, err = stdKeyStore.SignHash(account, messageHash)
	}

	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to sign message", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"signature":   hex.EncodeToString(signature),
		"message":     req.Message,
		"messageHash": hex.EncodeToString(messageHash),
	})
}

// handleStdUnlockAccount unlocks an account in standard keystore
func handleStdUnlockAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address    string `json:"address"`
		Passphrase string `json:"passphrase"`
		Timeout    int    `json:"timeout"` // in seconds, 0 for indefinite
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}
	timeout := time.Duration(req.Timeout) * time.Second

	err := stdKeyStore.TimedUnlock(account, req.Passphrase, timeout)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to unlock account", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"unlocked": true,
		"timeout":  req.Timeout,
	})
}

// handleStdLockAccount locks an account in standard keystore
func handleStdLockAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	err := stdKeyStore.Lock(common.HexToAddress(req.Address))
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to lock account", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"locked": true,
	})
}

// handleStdGetAccounts returns all accounts in the standard keystore
func handleStdGetAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountsList := stdKeyStore.Accounts()
	accountsData := make([]map[string]interface{}, len(accountsList))

	for i, acc := range accountsList {
		accountsData[i] = map[string]interface{}{
			"address": acc.Address.Hex(),
			"url":     acc.URL.String(),
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accountsData,
		"count":    len(accountsData),
	})
}

// handleStdGetAccountInfo returns detailed information about an account in standard keystore
func handleStdGetAccountInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	addressStr := vars["address"]

	if !common.IsHexAddress(addressStr) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	address := common.HexToAddress(addressStr)
	account := accounts.Account{Address: address}

	// Find the account
	foundAccount, err := stdKeyStore.Find(account)
	if err != nil {
		sendJSONError(w, http.StatusNotFound, "account not found", err)
		return
	}

	// Check if account is unlocked by trying to get wallet status
	isUnlocked := false
	wallets := stdKeyStore.Wallets()
	for _, wallet := range wallets {
		if wallet.Contains(foundAccount) {
			status, err := wallet.Status()
			if err == nil && status == "Unlocked" {
				isUnlocked = true
			}
			break
		}
	}

	// Read the keystore file contents
	filePath := foundAccount.URL.Path
	var fileContents string
	if filePath != "" {
		contents, err := os.ReadFile(filePath)
		if err != nil {
			// If file can't be read, still return other info but indicate error
			fileContents = fmt.Sprintf("Error reading file: %v", err)
		} else {
			fileContents = string(contents)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address":      foundAccount.Address.Hex(),
		"url":          foundAccount.URL.String(),
		"filePath":     filePath,
		"fileContents": fileContents,
		"unlocked":     isUnlocked,
		"hasAddress":   stdKeyStore.HasAddress(address),
	})
}

// handleStdDeleteAccount deletes an account from standard keystore
func handleStdDeleteAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Address    string `json:"address"`
		Passphrase string `json:"passphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to decode request", err)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendJSONError(w, http.StatusBadRequest, "invalid address", nil)
		return
	}

	account := accounts.Account{Address: common.HexToAddress(req.Address)}
	err := stdKeyStore.Delete(account, req.Passphrase)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, "failed to delete account", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": true,
		"address": req.Address,
	})
}
