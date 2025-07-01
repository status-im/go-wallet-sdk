package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/go-wallet-sdk/pkg/balance/fetcher"
)

// handleGetChains handles the GET /api/chains endpoint
func handleGetChains(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Implement the logic to return the list of available chains
}

// handleFetchBalances handles the POST /fetch endpoint
func handleFetchBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FetchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := FetchResponse{
		Results: make(map[string]map[string]BalanceResult),
		Errors:  []string{},
	}

	// Parse block number
	var atBlock gethrpc.BlockNumber
	if req.BlockNum == "" {
		atBlock = gethrpc.LatestBlockNumber
	} else {
		blockNum, err := strconv.ParseInt(req.BlockNum, 10, 64)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Invalid block number: %s", req.BlockNum))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		atBlock = gethrpc.BlockNumber(blockNum)
	}

	// Process each chain
	for _, chainConfig := range req.Chains {
		log.Printf("Processing chain ID %d with RPC URL: %s", chainConfig.ChainID, chainConfig.RPCURL)

		// Create RPC client
		rpcClient, err := NewCustomRPCClient(chainConfig.RPCURL)
		if err != nil {
			log.Printf("Chain ID %d: Failed to create RPC client: %v", chainConfig.ChainID, err)
			// Chain-level error
			response.Results[fmt.Sprintf("%d", chainConfig.ChainID)] = map[string]BalanceResult{
				"__chain_error__": {Error: fmt.Sprintf("Failed to connect: %v", err)},
			}
			continue
		}
		defer rpcClient.client.Close()

		// Convert addresses
		addresses := make([]common.Address, 0, len(req.Addresses))
		for _, addrStr := range req.Addresses {
			addrStr = strings.TrimSpace(addrStr)
			if addrStr == "" {
				continue
			}
			if !common.IsHexAddress(addrStr) {
				response.Errors = append(response.Errors, fmt.Sprintf("Invalid address: %s", addrStr))
				continue
			}
			addresses = append(addresses, common.HexToAddress(addrStr))
		}

		if len(addresses) == 0 {
			log.Printf("Chain ID %d: No valid addresses provided", chainConfig.ChainID)
			response.Results[fmt.Sprintf("%d", chainConfig.ChainID)] = map[string]BalanceResult{
				"__chain_error__": {Error: "No valid addresses provided"},
			}
			continue
		}

		log.Printf("Chain ID %d: Fetching balances for %d addresses", chainConfig.ChainID, len(addresses))

		// Fetch balances
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		balances, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, rpcClient, 10)
		if err != nil {
			log.Printf("Chain ID %d: Failed to fetch balances: %v", chainConfig.ChainID, err)
			response.Results[fmt.Sprintf("%d", chainConfig.ChainID)] = map[string]BalanceResult{
				"__chain_error__": {Error: fmt.Sprintf("Failed to fetch balances: %v", err)},
			}
			continue
		}

		log.Printf("Chain ID %d: Successfully fetched balances for %d addresses", chainConfig.ChainID, len(balances))

		// Store results
		chainResults := make(map[string]BalanceResult)
		for _, addr := range addresses {
			addrStr := addr.Hex()
			balance := balances[addr]

			if balance == nil {
				chainResults[addrStr] = BalanceResult{
					Address: addrStr,
					Balance: "0",
					Wei:     "0",
				}
			} else {
				chainResults[addrStr] = BalanceResult{
					Address: addrStr,
					Balance: weiToEther(balance),
					Wei:     balance.String(),
				}
			}
		}

		response.Results[fmt.Sprintf("%d", chainConfig.ChainID)] = chainResults
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
