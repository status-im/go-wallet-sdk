package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/common"
	statuscommon "github.com/status-im/go-wallet-sdk/pkg/common"
	"github.com/status-im/go-wallet-sdk/pkg/tokenlists"
)

func printToken(token *tokenlists.Token) {
	fmt.Printf("Token Key: %s:\n", token.Key())
	fmt.Printf("  Name: %s (%s)\n", token.Name, token.Symbol)
	fmt.Printf("  Chain ID: %d\n", token.ChainID)
	fmt.Printf("  Address: %s\n", token.Address.Hex())
	fmt.Printf("  Decimals: %d\n", token.Decimals)
	fmt.Printf("  Native: %t\n", token.IsNative())
	fmt.Printf("  Custom: %t\n", token.CustomToken)
	if token.LogoURI != "" {
		fmt.Printf("  Logo: %s\n", token.LogoURI)
	}
	fmt.Println()
}

func main() {
	fmt.Println("🪙 Token Lists Management Example")
	fmt.Println("=====================================")

	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create configuration with defaults
	config := tokenlists.DefaultConfig()

	// Configure with logger, main list, and specific chains
	config.WithLogger(logger).
		WithMainList(tokenlists.StatusListID, []byte(tokenlists.StatusTokenListJSON)).
		WithChains([]uint64{
			statuscommon.EthereumMainnet,
			statuscommon.OptimismMainnet,
			statuscommon.ArbitrumMainnet,
			statuscommon.BSCMainnet,
			statuscommon.BaseMainnet,
		}).
		WithParsers(tokenlists.DefaultParsers).
		WithAutoRefreshInterval(5*time.Minute, time.Minute)

	// Create token lists manager
	tokensList, err := tokenlists.NewTokensList(config)
	if err != nil {
		log.Fatalf("Failed to create basic token list: %v", err)
	}

	// Start the service
	notifyCh := make(chan struct{}, 1)
	if err := tokensList.Start(ctx, notifyCh); err != nil {
		log.Fatalf("Failed to start token list service: %v", err)
	}
	defer tokensList.Stop()

	// Example 1: Query tokens by various methods
	fmt.Println("\n🔍 Example 1: Querying Tokens")
	fmt.Println("-----------------------------")

	// Get all unique tokens
	allTokens := tokensList.UniqueTokens()
	fmt.Printf("Total unique tokens: %d\n", len(allTokens))

	// Show first few tokens
	for _, token := range allTokens {
		printToken(token)
	}

	// Example 2: Get tokens by specific chain
	fmt.Println("\n⛓️  Example 2: Tokens by Chain")
	fmt.Println("-----------------------------")

	// Get Ethereum mainnet tokens
	ethTokens := tokensList.GetTokensByChain(statuscommon.EthereumMainnet)
	fmt.Printf("Ethereum mainnet tokens: %d\n", len(ethTokens))

	// Show first few Ethereum tokens
	for _, token := range ethTokens {
		printToken(token)
	}

	// Get Optimism tokens
	opTokens := tokensList.GetTokensByChain(statuscommon.OptimismMainnet)
	fmt.Printf("Optimism tokens: %d\n", len(opTokens))

	// Example 3: Get specific token by address
	fmt.Println("\n🎯 Example 3: Get Token by Address")
	fmt.Println("----------------------------------")

	// Look for USDT on Ethereum (example address)
	usdtAddress := common.HexToAddress("0x1234")
	token, found := tokensList.GetTokenByChainAddress(statuscommon.EthereumMainnet, usdtAddress)
	if found {
		fmt.Printf("Found token at address: %s\n", usdtAddress.Hex())
		printToken(token)
	} else {
		fmt.Printf("Token not found at address: %s\n", usdtAddress.Hex())
	}

	sntOptimismAddress := common.HexToAddress("0x650af3c15af43dcb218406d30784416d64cfb6b2")
	token, found = tokensList.GetTokenByChainAddress(statuscommon.OptimismMainnet, sntOptimismAddress)
	if found {
		fmt.Printf("Found token at address: %s\n", sntOptimismAddress.Hex())
		printToken(token)
	} else {
		fmt.Printf("Token not found at address: %s\n", sntOptimismAddress.Hex())
	}

	// Example 4: Working with Token Lists
	fmt.Println("\n📝 Example 4: Token Lists Information")
	fmt.Println("------------------------------------")

	allTokenLists := tokensList.TokenLists()
	fmt.Printf("Total token lists: %d\n", len(allTokenLists))

	for _, tokenList := range allTokenLists {
		fmt.Printf("List: %s\n", tokenList.Name)
		fmt.Printf("  Source: %s\n", tokenList.Source)
		fmt.Printf("  Tokens: %d\n", len(tokenList.Tokens))
		fmt.Printf("  Version: %s\n", tokenList.Version.String())
		if tokenList.Timestamp != "" {
			fmt.Printf("  Timestamp: %s\n", tokenList.Timestamp)
		}
		if tokenList.FetchedTimestamp != "" {
			fmt.Printf("  Fetched: %s\n", tokenList.FetchedTimestamp)
		}

		// Show sample tokens from this list
		if len(tokenList.Tokens) > 0 && len(tokenList.Tokens) <= 10 {
			fmt.Printf("  Tokens in this list:\n")
			for _, token := range tokenList.Tokens {
				fmt.Printf("    - %s (%s) on chain %d at %s\n", token.Name, token.Symbol, token.ChainID, token.Address.Hex())
			}
		} else if len(tokenList.Tokens) > 10 {
			fmt.Printf("  Sample tokens from this list:\n")
			for i, token := range tokenList.Tokens {
				if i >= 3 { // Show first 3
					break
				}
				fmt.Printf("    - %s (%s) on chain %d at %s\n", token.Name, token.Symbol, token.ChainID, token.Address.Hex())
			}
		}
		fmt.Println()
	}

	// Example 5: Get specific token list
	fmt.Println("\n📄 Example 5: Specific Token List")
	fmt.Println("---------------------------------")

	if nativeList, found := tokensList.TokenList(tokenlists.NativeTokenListID); found {
		fmt.Printf("Native token list: %s\n", nativeList.Name)
		fmt.Printf("Native tokens count: %d\n", len(nativeList.Tokens))
		for _, token := range nativeList.Tokens {
			fmt.Printf("  %s (%s) on chain %d\n", token.Name, token.Symbol, token.ChainID)
		}
	}

	// Try to get Status token list (may not be available in this basic example)
	if statusList, found := tokensList.TokenList(tokenlists.StatusListID); found {
		fmt.Printf("\nStatus token list: %s\n", statusList.Name)
		fmt.Printf("Status tokens count: %d\n", len(statusList.Tokens))
		for _, token := range statusList.Tokens {
			fmt.Printf("  %s (%s) on chain %d at %s\n", token.Name, token.Symbol, token.ChainID, token.Address.Hex())
		}
	} else {
		fmt.Printf("\nNote: Status token list not loaded in this basic example\n")
		fmt.Printf("In production, token lists would be fetched from remote sources\n")
	}

	// Example 6: Token key utilities
	fmt.Println("\n🔑 Example 6: Token Key Utilities")
	fmt.Println("---------------------------------")

	// Create token key
	testAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	tokenKey := tokenlists.TokenKey(statuscommon.EthereumMainnet, testAddress)
	fmt.Printf("Token key: %s\n", tokenKey)

	// Parse token key back
	if chainID, address, valid := tokenlists.ChainAndAddressFromTokenKey(tokenKey); valid {
		fmt.Printf("Parsed - Chain ID: %d, Address: %s\n", chainID, address.Hex())
	} else {
		fmt.Printf("Failed to parse token key: %s\n", tokenKey)
	}

	// Example 7: Check last refresh time
	fmt.Println("\n🔄 Example 7: Refresh Information")
	fmt.Println("---------------------------------")

	if lastRefresh, err := tokensList.LastRefreshTime(); err == nil {
		if lastRefresh.IsZero() {
			fmt.Println("Token lists have not been refreshed yet")
		} else {
			fmt.Printf("Last refresh: %s\n", lastRefresh.Format(time.RFC3339))
		}
	} else {
		fmt.Printf("Error getting last refresh time: %v\n", err)
	}

	// Example 8: Manual refresh (if privacy is off)
	fmt.Println("\n🔃 Example 8: Manual Refresh")
	fmt.Println("----------------------------")

	fmt.Println("Triggering manual refresh...")
	refreshCtx, refreshCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer refreshCancel()

	if err := tokensList.RefreshNow(refreshCtx); err != nil {
		fmt.Printf("Refresh error: %v\n", err)
	} else {
		fmt.Println("✅ Refresh triggered successfully")
	}

	// Wait a bit to see if we get a notification
	select {
	case <-notifyCh:
		fmt.Println("📬 Received update notification!")
		// Show updated counts
		updatedTokens := tokensList.UniqueTokens()
		fmt.Printf("Updated token count: %d\n", len(updatedTokens))
	case <-time.After(2 * time.Second):
		fmt.Println("No notification received (might be using cached data)")
	}

	fmt.Println("\n✅ Example completed successfully!")
}
