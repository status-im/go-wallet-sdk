package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/go-wallet-sdk/pkg/balance/fetcher"
)

// Global token list service
var tokenListService = NewTokenListService()

// handleGetChains handles the GET /api/chains endpoint
func handleGetChains(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Implement the logic to return the list of available chains
}

// handleGetTokens handles the GET /api/tokens/{chainID} endpoint
func handleGetTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract chain ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	chainIDStr := pathParts[3]
	chainID, err := strconv.Atoi(chainIDStr)
	if err != nil {
		http.Error(w, "Invalid chain ID", http.StatusBadRequest)
		return
	}

	// Get tokens for the chain
	tokens, err := tokenListService.GetTokensForChain(chainID)
	if err != nil {
		log.Printf("Failed to get tokens for chain %d: %v", chainID, err)
		// Return common tokens as fallback
		tokens = tokenListService.GetCommonTokens(chainID)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"tokens":  tokens,
		"chainId": chainID,
	})
}

// handleGetTokenListInfo handles the GET /api/tokenlist/info endpoint
func handleGetTokenListInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	info := tokenListService.GetTokenListInfo()
	json.NewEncoder(w).Encode(info)
}

// handleSearchTokens handles the GET /api/tokens/search?symbol={symbol} endpoint
func handleSearchTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Symbol parameter is required", http.StatusBadRequest)
		return
	}

	tokens := tokenListService.SearchTokensBySymbol(symbol)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol": symbol,
		"tokens": tokens,
		"count":  len(tokens),
	})
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
		Results: make(map[string]map[string]AccountBalances),
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
			response.Results[fmt.Sprintf("%d", chainConfig.ChainID)] = map[string]AccountBalances{
				"__chain_error__": {
					Address:       "__chain_error__",
					NativeBalance: BalanceResult{Error: fmt.Sprintf("Failed to connect: %v", err)},
					ERC20Balances: make(map[string]ERC20BalanceResult),
				},
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
			response.Results[fmt.Sprintf("%d", chainConfig.ChainID)] = map[string]AccountBalances{
				"__chain_error__": {
					Address:       "__chain_error__",
					NativeBalance: BalanceResult{Error: "No valid addresses provided"},
					ERC20Balances: make(map[string]ERC20BalanceResult),
				},
			}
			continue
		}

		log.Printf("Chain ID %d: Fetching balances for %d addresses", chainConfig.ChainID, len(addresses))

		// Fetch native balances
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		nativeBalances, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, rpcClient, 10)
		if err != nil {
			log.Printf("Chain ID %d: Failed to fetch native balances: %v", chainConfig.ChainID, err)
			response.Results[fmt.Sprintf("%d", chainConfig.ChainID)] = map[string]AccountBalances{
				"__chain_error__": {
					Address:       "__chain_error__",
					NativeBalance: BalanceResult{Error: fmt.Sprintf("Failed to fetch native balances: %v", err)},
					ERC20Balances: make(map[string]ERC20BalanceResult),
				},
			}
			continue
		}

		// Fetch ERC20 balances if token addresses are provided
		var erc20Balances fetcher.BalancePerAccountAndTokenAddress
		if len(chainConfig.TokenAddresses) > 0 {
			tokenAddresses := make([]common.Address, 0, len(chainConfig.TokenAddresses))
			for _, tokenAddrStr := range chainConfig.TokenAddresses {
				tokenAddrStr = strings.TrimSpace(tokenAddrStr)
				if tokenAddrStr == "" {
					continue
				}
				if !common.IsHexAddress(tokenAddrStr) {
					response.Errors = append(response.Errors, fmt.Sprintf("Invalid token address: %s", tokenAddrStr))
					continue
				}
				tokenAddresses = append(tokenAddresses, common.HexToAddress(tokenAddrStr))
			}

			if len(tokenAddresses) > 0 {
				log.Printf("Chain ID %d: Fetching ERC20 balances for %d tokens", chainConfig.ChainID, len(tokenAddresses))
				erc20Balances, err = fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, rpcClient, 10)
				if err != nil {
					log.Printf("Chain ID %d: Failed to fetch ERC20 balances: %v", chainConfig.ChainID, err)
					// Continue with native balances only
				}
			}
		}

		log.Printf("Chain ID %d: Successfully fetched balances for %d addresses", chainConfig.ChainID, len(nativeBalances))

		// Store results
		chainResults := make(map[string]AccountBalances)
		for _, addr := range addresses {
			addrStr := addr.Hex()
			nativeBalance := nativeBalances[addr]

			accountBalances := AccountBalances{
				Address:       addrStr,
				ERC20Balances: make(map[string]ERC20BalanceResult),
			}

			// Set native balance
			if nativeBalance == nil {
				accountBalances.NativeBalance = BalanceResult{
					Address: addrStr,
					Balance: "0",
					Wei:     "0",
				}
			} else {
				accountBalances.NativeBalance = BalanceResult{
					Address: addrStr,
					Balance: weiToEther(nativeBalance),
					Wei:     nativeBalance.String(),
				}
			}

			// Set ERC20 balances
			if erc20Balances != nil {
				if accountTokenBalances, exists := erc20Balances[addr]; exists {
					for tokenAddr, tokenBalance := range accountTokenBalances {
						tokenAddrStr := tokenAddr.Hex()

						// Get token info for display
						tokenInfo := getTokenInfo(int(chainConfig.ChainID), tokenAddrStr)

						accountBalances.ERC20Balances[tokenAddrStr] = ERC20BalanceResult{
							TokenAddress: tokenAddrStr,
							TokenSymbol:  tokenInfo.Symbol,
							TokenName:    tokenInfo.Name,
							Balance:      formatTokenBalance(tokenBalance, tokenInfo.Decimals),
							Wei:          tokenBalance.String(),
							Decimals:     tokenInfo.Decimals,
						}
					}
				}
			}

			chainResults[addrStr] = accountBalances
		}

		response.Results[fmt.Sprintf("%d", chainConfig.ChainID)] = chainResults
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getTokenInfo retrieves token information from the token list service
func getTokenInfo(chainID int, tokenAddress string) TokenInfo {
	tokens, err := tokenListService.GetTokensForChain(chainID)
	if err != nil {
		// Return fallback token info
		return TokenInfo{
			Address:  tokenAddress,
			Symbol:   "UNKNOWN",
			Name:     "Unknown Token",
			Decimals: 18,
			ChainID:  chainID,
		}
	}

	// Find the token in the list
	for _, token := range tokens {
		if strings.EqualFold(token.Address, tokenAddress) {
			return token
		}
	}

	// Return fallback token info if not found
	return TokenInfo{
		Address:  tokenAddress,
		Symbol:   "UNKNOWN",
		Name:     "Unknown Token",
		Decimals: 18,
		ChainID:  chainID,
	}
}

// formatTokenBalance formats a token balance based on its decimals
func formatTokenBalance(balance *big.Int, decimals int) string {
	if balance == nil {
		return "0"
	}

	// Convert to string with proper decimal places
	balanceStr := balance.String()

	if decimals == 0 {
		return balanceStr
	}

	if len(balanceStr) <= decimals {
		// Pad with leading zeros
		padded := strings.Repeat("0", decimals-len(balanceStr)+1) + balanceStr
		return "0." + padded[1:]
	}

	// Insert decimal point
	return balanceStr[:len(balanceStr)-decimals] + "." + balanceStr[len(balanceStr)-decimals:]
}
