package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// TokenInfo represents information about an ERC20 token (matches schema)
type TokenInfo struct {
	ChainID    int                    `json:"chainId"`
	Address    string                 `json:"address"`
	Decimals   int                    `json:"decimals"`
	Name       string                 `json:"name"`
	Symbol     string                 `json:"symbol"`
	LogoURI    string                 `json:"logoURI,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// TokenList represents the structure of Uniswap's token list (matches schema)
type TokenList struct {
	Name      string                   `json:"name"`
	Timestamp string                   `json:"timestamp"`
	Version   TokenVersion             `json:"version"`
	Tokens    []TokenInfo              `json:"tokens"`
	TokenMap  map[string]TokenInfo     `json:"tokenMap,omitempty"`
	Keywords  []string                 `json:"keywords,omitempty"`
	Tags      map[string]TagDefinition `json:"tags,omitempty"`
	LogoURI   string                   `json:"logoURI,omitempty"`
}

type TokenVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

type TagDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TokenListService manages token lists for different chains
type TokenListService struct {
	cache     map[int][]TokenInfo // chainID -> tokens
	cacheTime map[int]time.Time   // chainID -> last fetch time
	mutex     sync.RWMutex
	client    *http.Client
	tokenList *TokenList // Local token list
}

// NewTokenListService creates a new token list service
func NewTokenListService() *TokenListService {
	tls := &TokenListService{
		cache:     make(map[int][]TokenInfo),
		cacheTime: make(map[int]time.Time),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Load local token list
	if err := tls.loadLocalTokenList(); err != nil {
		fmt.Printf("Warning: Failed to load local token list: %v\n", err)
	}

	return tls
}

// loadLocalTokenList loads the token list from the local tokenlist.json file
func (tls *TokenListService) loadLocalTokenList() error {
	data, err := os.ReadFile("tokenlist.json")
	if err != nil {
		return fmt.Errorf("failed to read tokenlist.json: %w", err)
	}

	var tokenList TokenList
	if err := json.Unmarshal(data, &tokenList); err != nil {
		return fmt.Errorf("failed to parse tokenlist.json: %w", err)
	}

	tls.tokenList = &tokenList
	fmt.Printf("Loaded token list: %s (version %d.%d.%d) with %d tokens\n",
		tokenList.Name, tokenList.Version.Major, tokenList.Version.Minor, tokenList.Version.Patch, len(tokenList.Tokens))

	return nil
}

// GetTokensForChain returns tokens for a specific chain ID
func (tls *TokenListService) GetTokensForChain(chainID int) ([]TokenInfo, error) {
	tls.mutex.RLock()
	if tokens, exists := tls.cache[chainID]; exists {
		if time.Since(tls.cacheTime[chainID]) < 1*time.Hour { // Cache for 1 hour
			tls.mutex.RUnlock()
			return tokens, nil
		}
	}
	tls.mutex.RUnlock()

	// Try to get tokens from local token list first
	if tls.tokenList != nil {
		var filteredTokens []TokenInfo
		for _, token := range tls.tokenList.Tokens {
			if token.ChainID == chainID {
				filteredTokens = append(filteredTokens, token)
			}
		}

		if len(filteredTokens) > 0 {
			// Update cache
			tls.mutex.Lock()
			tls.cache[chainID] = filteredTokens
			tls.cacheTime[chainID] = time.Now()
			tls.mutex.Unlock()

			return filteredTokens, nil
		}
	}

	// Fallback to remote fetch if local list doesn't have tokens for this chain
	tokens, err := tls.fetchTokensForChain(chainID)
	if err != nil {
		return nil, err
	}

	// Update cache
	tls.mutex.Lock()
	tls.cache[chainID] = tokens
	tls.cacheTime[chainID] = time.Now()
	tls.mutex.Unlock()

	return tokens, nil
}

// fetchTokensForChain fetches tokens for a specific chain from Uniswap's token list
func (tls *TokenListService) fetchTokensForChain(chainID int) ([]TokenInfo, error) {
	// Uniswap default token list URL - using the main list
	url := "https://raw.githubusercontent.com/Uniswap/default-token-list/main/src/tokens/mainnet.json"

	// For now, we'll use the mainnet list and filter by chain ID
	// In a production environment, you might want to fetch different lists for different chains
	resp, err := tls.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch token list: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenList TokenList
	if err := json.Unmarshal(body, &tokenList); err != nil {
		return nil, fmt.Errorf("failed to parse token list: %w", err)
	}

	// Filter tokens by chain ID (mainnet = 1)
	var filteredTokens []TokenInfo
	for _, token := range tokenList.Tokens {
		if token.ChainID == chainID {
			filteredTokens = append(filteredTokens, token)
		}
	}

	// If no tokens found for the specific chain, return common tokens as fallback
	if len(filteredTokens) == 0 {
		return tls.GetCommonTokens(chainID), nil
	}

	return filteredTokens, nil
}

// GetCommonTokens returns a list of commonly used tokens for a chain
func (tls *TokenListService) GetCommonTokens(chainID int) []TokenInfo {
	// Common token addresses for mainnet (chain ID 1)
	commonTokens := map[int][]TokenInfo{
		1: { // Ethereum Mainnet
			{Address: "0xA0b86a33E6441b8C4C8C8C8C8C8C8C8C8C8C8C8", Symbol: "USDC", Name: "USD Coin", Decimals: 6, ChainID: 1},
			{Address: "0xdAC17F958D2ee523a2206206994597C13D831ec7", Symbol: "USDT", Name: "Tether USD", Decimals: 6, ChainID: 1},
			{Address: "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", Symbol: "WBTC", Name: "Wrapped BTC", Decimals: 8, ChainID: 1},
			{Address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", Symbol: "WETH", Name: "Wrapped Ether", Decimals: 18, ChainID: 1},
			{Address: "0x6B175474E89094C44Da98b954EedeAC495271d0F", Symbol: "DAI", Name: "Dai Stablecoin", Decimals: 18, ChainID: 1},
		},
		10: { // Optimism
			{Address: "0x7F5c764cBc14f9669B88837ca1490cCa17c31607", Symbol: "USDC", Name: "USD Coin", Decimals: 6, ChainID: 10},
			{Address: "0x94b008aA00579c1307B0EF2c499aD98a8ce58e58", Symbol: "USDT", Name: "Tether USD", Decimals: 6, ChainID: 10},
			{Address: "0x4200000000000000000000000000000000000006", Symbol: "WETH", Name: "Wrapped Ether", Decimals: 18, ChainID: 10},
		},
		42161: { // Arbitrum
			{Address: "0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8", Symbol: "USDC", Name: "USD Coin", Decimals: 6, ChainID: 42161},
			{Address: "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9", Symbol: "USDT", Name: "Tether USD", Decimals: 6, ChainID: 42161},
			{Address: "0x82aF49447D8a07e3bd95BD0d56f35241523fBab1", Symbol: "WETH", Name: "Wrapped Ether", Decimals: 18, ChainID: 42161},
		},
		137: { // Polygon
			{Address: "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174", Symbol: "USDC", Name: "USD Coin", Decimals: 6, ChainID: 137},
			{Address: "0xc2132D05D31c914a87C6611C10748AEb04B58e8F", Symbol: "USDT", Name: "Tether USD", Decimals: 6, ChainID: 137},
			{Address: "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270", Symbol: "WMATIC", Name: "Wrapped MATIC", Decimals: 18, ChainID: 137},
		},
		// Add more chains as needed
	}

	if tokens, exists := commonTokens[chainID]; exists {
		return tokens
	}

	// Return empty list for unknown chains
	return []TokenInfo{}
}

// ValidateTokenAddress validates if a token address is properly formatted
func (tls *TokenListService) ValidateTokenAddress(address string) bool {
	// Basic validation - should be a hex address
	if len(address) != 42 || address[:2] != "0x" {
		return false
	}

	// Check if all characters are valid hex
	for _, char := range address[2:] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}

// GetTokenInfo returns token information for a specific address and chain
func (tls *TokenListService) GetTokenInfo(chainID int, address string) (*TokenInfo, error) {
	tokens, err := tls.GetTokensForChain(chainID)
	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		if token.Address == address {
			return &token, nil
		}
	}

	return nil, fmt.Errorf("token not found for address %s on chain %d", address, chainID)
}

// SearchTokensBySymbol searches for tokens by symbol across all chains
func (tls *TokenListService) SearchTokensBySymbol(symbol string) []TokenInfo {
	if tls.tokenList == nil {
		return []TokenInfo{}
	}

	var results []TokenInfo
	symbolUpper := strings.ToUpper(symbol)

	for _, token := range tls.tokenList.Tokens {
		if strings.ToUpper(token.Symbol) == symbolUpper {
			results = append(results, token)
		}
	}

	return results
}

// GetTokenCountByChain returns the number of tokens for each chain
func (tls *TokenListService) GetTokenCountByChain() map[int]int {
	if tls.tokenList == nil {
		return map[int]int{}
	}

	counts := make(map[int]int)
	for _, token := range tls.tokenList.Tokens {
		counts[token.ChainID]++
	}

	return counts
}

// GetSupportedChains returns a list of chain IDs that have tokens in the list
func (tls *TokenListService) GetSupportedChains() []int {
	if tls.tokenList == nil {
		return []int{}
	}

	chainSet := make(map[int]bool)
	for _, token := range tls.tokenList.Tokens {
		chainSet[token.ChainID] = true
	}

	chains := make([]int, 0, len(chainSet))
	for chainID := range chainSet {
		chains = append(chains, chainID)
	}

	return chains
}

// GetTokenListInfo returns information about the loaded token list
func (tls *TokenListService) GetTokenListInfo() map[string]interface{} {
	if tls.tokenList == nil {
		return map[string]interface{}{
			"loaded": false,
			"error":  "No token list loaded",
		}
	}

	chainCounts := tls.GetTokenCountByChain()
	supportedChains := tls.GetSupportedChains()

	return map[string]interface{}{
		"loaded":          true,
		"name":            tls.tokenList.Name,
		"version":         fmt.Sprintf("%d.%d.%d", tls.tokenList.Version.Major, tls.tokenList.Version.Minor, tls.tokenList.Version.Patch),
		"timestamp":       tls.tokenList.Timestamp,
		"totalTokens":     len(tls.tokenList.Tokens),
		"supportedChains": supportedChains,
		"chainCounts":     chainCounts,
		"keywords":        tls.tokenList.Keywords,
		"logoURI":         tls.tokenList.LogoURI,
	}
}
