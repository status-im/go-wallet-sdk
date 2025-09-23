package main

import (
	"fmt"
	"log"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/common"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/builder"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

func main() {
	fmt.Println("🏗️  Token Builder Example")
	fmt.Println("==========================")

	// Define supported blockchain networks
	supportedChains := []uint64{common.EthereumMainnet, common.BSCMainnet, common.OptimismMainnet, common.ArbitrumMainnet}

	// Example 1: Basic builder usage
	fmt.Println("\n🚀 Basic Builder Usage")
	fmt.Println("=======================")
	demonstrateBasicBuilder(supportedChains)

	// Example 2: Incremental building pattern
	fmt.Println("\n📈 Incremental Building Pattern")
	fmt.Println("===============================")
	demonstrateIncrementalBuilding(supportedChains)

	// Example 3: Adding raw token lists with parsers
	fmt.Println("\n📄 Raw Token List Processing")
	fmt.Println("============================")
	demonstrateRawTokenListProcessing(supportedChains)

	// Example 4: Deduplication demonstration
	fmt.Println("\n🔄 Token Deduplication")
	fmt.Println("======================")
	demonstrateDeduplication(supportedChains)

	// Example 5: Advanced builder patterns
	fmt.Println("\n🎯 Advanced Builder Patterns")
	fmt.Println("============================")
	demonstrateAdvancedPatterns(supportedChains)

	fmt.Println("\n✅ Token Builder examples completed!")
}

func demonstrateBasicBuilder(supportedChains []uint64) {
	// Create new builder starting empty
	tokenBuilder := builder.New(supportedChains)

	fmt.Printf("🏗️  Created builder for %d chains\n", len(supportedChains))
	fmt.Printf("📊 Initial state: %d tokens, %d lists\n",
		len(tokenBuilder.GetTokens()), len(tokenBuilder.GetTokenLists()))

	// Add native tokens first
	err := tokenBuilder.AddNativeTokenList()
	if err != nil {
		log.Printf("❌ Failed to add native tokens: %v", err)
		return
	}

	fmt.Printf("🌐 Added native tokens: %d tokens, %d lists\n",
		len(tokenBuilder.GetTokens()), len(tokenBuilder.GetTokenLists()))

	// Create and add a custom token list
	customList := createSampleTokenList()
	tokenBuilder.AddTokenList("custom-tokens", customList)

	fmt.Printf("➕ Added custom token list: %d tokens, %d lists\n",
		len(tokenBuilder.GetTokens()), len(tokenBuilder.GetTokenLists()))

	// Display final results
	fmt.Println("\n📋 Final Token Collection:")
	displayTokenSummary(tokenBuilder)
}

func demonstrateIncrementalBuilding(supportedChains []uint64) {
	// Start with empty builder
	tokenBuilder := builder.New(supportedChains)

	fmt.Println("🏗️  Building token collection incrementally...")

	// Step 1: Add native tokens
	fmt.Println("\n1️⃣ Adding native tokens...")
	err := tokenBuilder.AddNativeTokenList()
	if err != nil {
		log.Printf("❌ Failed: %v", err)
		return
	}
	fmt.Printf("   ✅ Native tokens added: %d total tokens\n", len(tokenBuilder.GetTokens()))

	// Step 2: Add DeFi tokens
	fmt.Println("\n2️⃣ Adding DeFi token list...")
	defiList := createDefiTokenList()
	tokenBuilder.AddTokenList("defi-tokens", defiList)
	fmt.Printf("   ✅ DeFi tokens added: %d total tokens\n", len(tokenBuilder.GetTokens()))

	// Step 3: Add stablecoin list
	fmt.Println("\n3️⃣ Adding stablecoin list...")
	stablecoinList := createStablecoinTokenList()
	tokenBuilder.AddTokenList("stablecoins", stablecoinList)
	fmt.Printf("   ✅ Stablecoins added: %d total tokens\n", len(tokenBuilder.GetTokens()))

	// Step 4: Add exchange tokens
	fmt.Println("\n4️⃣ Adding exchange token list...")
	exchangeList := createExchangeTokenList()
	tokenBuilder.AddTokenList("exchange-tokens", exchangeList)
	fmt.Printf("   ✅ Exchange tokens added: %d total tokens\n", len(tokenBuilder.GetTokens()))

	// Show building progress
	fmt.Println("\n📊 Building Progress Summary:")
	lists := tokenBuilder.GetTokenLists()
	for listID, list := range lists {
		fmt.Printf("   📋 %s: %d tokens\n", listID, len(list.Tokens))
	}

	fmt.Printf("\n🎯 Final collection: %d unique tokens across %d lists\n",
		len(tokenBuilder.GetTokens()), len(lists))
}

func demonstrateRawTokenListProcessing(supportedChains []uint64) {
	tokenBuilder := builder.New(supportedChains)

	// Add native tokens first
	err := tokenBuilder.AddNativeTokenList()
	if err != nil {
		log.Printf("❌ Failed to add native tokens: %v", err)
		return
	}

	// Example 1: Process standard format raw data
	fmt.Println("📄 Processing standard format token list...")
	standardJSON := `{
		"name": "Uniswap Example List",
		"timestamp": "2025-01-01T00:00:00Z",
		"version": {"major": 1, "minor": 0, "patch": 0},
		"tokens": [
			{
				"chainId": 1,
				"address": "0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB",
				"symbol": "USDC",
				"name": "USD Coin",
				"decimals": 6,
				"logoURI": "https://tokens.1inch.io/0xa0b86a33e6441b6d9e4aeda6d7bb57b75fe3f5db.png"
			},
			{
				"chainId": 56,
				"address": "0x55d398326f99059fF775485246999027B3197955",
				"symbol": "USDT",
				"name": "Tether USD",
				"decimals": 18
			}
		]
	}`

	standardParser := &parsers.StandardTokenListParser{}
	err = tokenBuilder.AddRawTokenList(
		"uniswap-example",
		[]byte(standardJSON),
		"https://example.com/uniswap-list.json",
		time.Now(),
		standardParser,
	)
	if err != nil {
		log.Printf("❌ Failed to add standard list: %v", err)
	} else {
		fmt.Printf("   ✅ Standard list processed: %d total tokens\n", len(tokenBuilder.GetTokens()))
	}

	// Example 2: Process Status format raw data
	fmt.Println("\n📄 Processing Status format token list...")
	statusJSON := `{
		"name": "Status Example List",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"version": {"major": 2, "minor": 1, "patch": 0},
		"tokens": [
			{
				"crossChainId": "usd-coin",
				"symbol": "USDC",
				"name": "USDC (EVM)",
				"decimals": 6,
				"logoURI": "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
				"contracts": {
					"1": "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
					"10": "0x0b2c639c533813f4aa9d7837caf62653d097ff85",
					"8453": "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
					"42161": "0xaf88d065e77c8cc2239327c5edb3a432268e5831",
					"84532": "0x036cbd53842c5426634e7929541ec2318f3dcf7e",
					"421614": "0x75faf114eafb1bdbe2f0316df893fd58ce46aa4d",
					"11155111": "0x1c7d4b196cb0c7b01d743fbc6116a902379c7238",
					"11155420": "0x5fd84259d66cd46123540766be93dfe6d43130d7",
					"1660990954": "0xc445a18ca49190578dad62fba3048c07efc07ffe"
				}
			},
			{
				"crossChainId": "usd-coin-bsc",
				"symbol": "USDC",
				"name": "USDC (BSC)",
				"decimals": 18,
				"logoURI": "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48/logo.png",
				"contracts": {
					"56": "0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d"
				}
			}
		]
	}`

	statusParser := &parsers.StatusTokenListParser{}
	err = tokenBuilder.AddRawTokenList(
		"status-example",
		[]byte(statusJSON),
		"https://example.com/status-list.json",
		time.Now(),
		statusParser,
	)
	if err != nil {
		log.Printf("❌ Failed to add Status list: %v", err)
	} else {
		fmt.Printf("   ✅ Status list processed: %d total tokens\n", len(tokenBuilder.GetTokens()))
	}

	// Show final results
	fmt.Println("\n📋 Raw Processing Results:")
	displayTokenSummary(tokenBuilder)
}

func demonstrateDeduplication(supportedChains []uint64) {
	tokenBuilder := builder.New(supportedChains)

	// Add native tokens
	err := tokenBuilder.AddNativeTokenList()
	if err != nil {
		log.Printf("❌ Failed to add native tokens: %v", err)
		return
	}

	initialCount := len(tokenBuilder.GetTokens())
	fmt.Printf("🌐 Initial tokens (native): %d\n", initialCount)

	// Create overlapping token lists
	fmt.Println("\n📄 Adding overlapping token lists...")

	// List 1: Popular tokens
	list1 := &types.TokenList{
		Name: "Popular Tokens List 1",
		Tokens: []*types.Token{
			{
				CrossChainID: "usdc-ethereum",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB"),
				Symbol:       "USDC",
				Name:         "USD Coin",
				Decimals:     6,
			},
			{
				CrossChainID: "usdt-ethereum",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
				Symbol:       "USDT",
				Name:         "Tether USD",
				Decimals:     6,
			},
		},
	}

	// List 2: Duplicate USDC + additional token
	list2 := &types.TokenList{
		Name: "Popular Tokens List 2",
		Tokens: []*types.Token{
			{
				CrossChainID: "usdc-ethereum", // Same as list 1
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB"),
				Symbol:       "USDC",
				Name:         "USD Coin",
				Decimals:     6,
			},
			{
				CrossChainID: "weth-ethereum",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				Symbol:       "WETH",
				Name:         "Wrapped Ether",
				Decimals:     18,
			},
		},
	}

	// List 3: USDC on different chain (should NOT be deduplicated)
	list3 := &types.TokenList{
		Name: "BSC Tokens",
		Tokens: []*types.Token{
			{
				CrossChainID: "usdc-bsc",
				ChainID:      56,
				Address:      gethcommon.HexToAddress("0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d"),
				Symbol:       "USDC",
				Name:         "USD Coin (BSC)",
				Decimals:     18,
			},
		},
	}

	// Add lists and show deduplication in action
	tokenBuilder.AddTokenList("popular-1", list1)
	count1 := len(tokenBuilder.GetTokens())
	fmt.Printf("   ➕ Added list 1: %d tokens (+%d)\n", count1, count1-initialCount)

	tokenBuilder.AddTokenList("popular-2", list2)
	count2 := len(tokenBuilder.GetTokens())
	fmt.Printf("   ➕ Added list 2: %d tokens (+%d) - USDC deduplicated!\n", count2, count2-count1)

	tokenBuilder.AddTokenList("bsc-tokens", list3)
	count3 := len(tokenBuilder.GetTokens())
	fmt.Printf("   ➕ Added list 3: %d tokens (+%d) - Different chain USDC kept\n", count3, count3-count2)

	// Show deduplication analysis
	fmt.Println("\n🔍 Deduplication Analysis:")
	tokens := tokenBuilder.GetTokens()
	usdcCount := 0
	for _, token := range tokens {
		if token.Symbol == "USDC" {
			usdcCount++
			fmt.Printf("   💰 USDC on chain %d: %s\n", token.ChainID, token.Address.Hex())
		}
	}
	fmt.Printf("   📊 Total USDC tokens: %d (different chains = different tokens)\n", usdcCount)

	fmt.Printf("\n✅ Deduplication complete: %d unique tokens from %d lists\n",
		len(tokens), len(tokenBuilder.GetTokenLists()))
}

func demonstrateAdvancedPatterns(supportedChains []uint64) {
	fmt.Println("🎯 Advanced Builder Pattern Examples:")

	// Pattern 1: Builder with validation
	fmt.Println("\n1️⃣ Builder with validation:")
	validationBuilder := builder.New(supportedChains)
	err := validationBuilder.AddNativeTokenList()
	if err != nil {
		log.Printf("❌ Validation failed: %v", err)
	} else {
		fmt.Printf("   ✅ Validation passed: %d native tokens added\n", len(validationBuilder.GetTokens()))
	}

	// Pattern 2: Conditional building
	fmt.Println("\n2️⃣ Conditional building based on chain support:")
	conditionalBuilder := builder.New([]uint64{1}) // Only Ethereum

	// This will only include Ethereum native token
	err = conditionalBuilder.AddNativeTokenList()
	if err != nil {
		log.Printf("❌ Failed: %v", err)
	} else {
		fmt.Printf("   ✅ Ethereum-only builder: %d tokens\n", len(conditionalBuilder.GetTokens()))
	}

	// Pattern 3: Builder state inspection
	fmt.Println("\n3️⃣ Builder state inspection:")
	inspectionBuilder := builder.New(supportedChains)
	inspectionBuilder.AddNativeTokenList()

	lists := inspectionBuilder.GetTokenLists()
	tokens := inspectionBuilder.GetTokens()

	fmt.Printf("   📊 Builder state:\n")
	fmt.Printf("      • Total tokens: %d\n", len(tokens))
	fmt.Printf("      • Total lists: %d\n", len(lists))
	fmt.Printf("      • Memory efficiency: ~%d bytes per token\n",
		estimateTokenMemoryUsage(tokens))

	// Pattern 4: Error handling strategies
	fmt.Println("\n4️⃣ Error handling strategies:")
	demonstrateErrorHandling()
}

func createSampleTokenList() *types.TokenList {
	return &types.TokenList{
		Name:      "Sample Token List",
		Timestamp: "2025-01-01T00:00:00Z",
		Source:    "internal",
		Version:   types.Version{Major: 1, Minor: 0, Patch: 0},
		Tokens: []*types.Token{
			{
				CrossChainID: "sample-token-1",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0x1234567890123456789012345678901234567890"),
				Symbol:       "SAMPLE1",
				Name:         "Sample Token 1",
				Decimals:     18,
				LogoURI:      "https://example.com/sample1.png",
			},
			{
				CrossChainID: "sample-token-2",
				ChainID:      56,
				Address:      gethcommon.HexToAddress("0x2345678901234567890123456789012345678901"),
				Symbol:       "SAMPLE2",
				Name:         "Sample Token 2",
				Decimals:     8,
				LogoURI:      "https://example.com/sample2.png",
			},
		},
	}
}

func createDefiTokenList() *types.TokenList {
	return &types.TokenList{
		Name:      "DeFi Tokens",
		Timestamp: "2025-01-01T00:00:00Z",
		Source:    "defi-protocols",
		Tokens: []*types.Token{
			{
				CrossChainID: "compound-token",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0xc00e94Cb662C3520282E6f5717214004A7f26888"),
				Symbol:       "COMP",
				Name:         "Compound",
				Decimals:     18,
			},
			{
				CrossChainID: "aave-token",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9"),
				Symbol:       "AAVE",
				Name:         "Aave",
				Decimals:     18,
			},
		},
	}
}

func createStablecoinTokenList() *types.TokenList {
	return &types.TokenList{
		Name:      "Stablecoins",
		Timestamp: "2025-01-01T00:00:00Z",
		Source:    "stablecoin-registry",
		Tokens: []*types.Token{
			{
				CrossChainID: "dai-ethereum",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"),
				Symbol:       "DAI",
				Name:         "Dai Stablecoin",
				Decimals:     18,
			},
			{
				CrossChainID: "busd-bsc",
				ChainID:      56,
				Address:      gethcommon.HexToAddress("0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"),
				Symbol:       "BUSD",
				Name:         "Binance USD",
				Decimals:     18,
			},
		},
	}
}

func createExchangeTokenList() *types.TokenList {
	return &types.TokenList{
		Name:      "Exchange Tokens",
		Timestamp: "2025-01-01T00:00:00Z",
		Source:    "exchange-registry",
		Tokens: []*types.Token{
			{
				CrossChainID: "binance-coin",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0xB8c77482e45F1F44dE1745F52C74426C631bDD52"),
				Symbol:       "BNB",
				Name:         "Binance Coin",
				Decimals:     18,
			},
			{
				CrossChainID: "ftx-token",
				ChainID:      1,
				Address:      gethcommon.HexToAddress("0x50D1c9771902476076eCFc8B2A83Ad6b9355a4c9"),
				Symbol:       "FTT",
				Name:         "FTX Token",
				Decimals:     18,
			},
		},
	}
}

func displayTokenSummary(tokenBuilder *builder.Builder) {
	tokens := tokenBuilder.GetTokens()
	lists := tokenBuilder.GetTokenLists()

	fmt.Printf("   📊 Summary: %d unique tokens from %d lists\n", len(tokens), len(lists))

	// Group tokens by chain
	chainCounts := make(map[uint64]int)
	for _, token := range tokens {
		chainCounts[token.ChainID]++
	}

	fmt.Println("   ⛓️  Tokens per chain:")
	for chainID, count := range chainCounts {
		chainName := getChainName(chainID)
		fmt.Printf("      • Chain %d (%s): %d tokens\n", chainID, chainName, count)
	}

	fmt.Println("   📋 Token lists:")
	for listID, list := range lists {
		fmt.Printf("      • %s: %s (%d tokens)\n", listID, list.Name, len(list.Tokens))
	}
}

func getChainName(chainID uint64) string {
	switch chainID {
	case common.EthereumMainnet:
		return "Ethereum"
	case common.BSCMainnet:
		return "BSC"
	case common.OptimismMainnet:
		return "Optimism"
	case common.ArbitrumMainnet:
		return "Arbitrum"
	default:
		return "Unknown"
	}
}

func estimateTokenMemoryUsage(tokens map[string]*types.Token) int {
	if len(tokens) == 0 {
		return 0
	}
	// Rough estimate: ~200 bytes per token (including maps, strings, etc.)
	return 200
}

func demonstrateErrorHandling() {
	fmt.Println("   🛠️  Error handling examples:")

	builder := builder.New([]uint64{1})

	// Test 1: Empty raw data
	fmt.Println("      📝 Testing empty raw data...")
	err := builder.AddRawTokenList("empty", []byte{}, "test", time.Now(), &parsers.StandardTokenListParser{})
	if err != nil {
		fmt.Printf("      ✅ Correctly caught error: %v\n", err)
	}

	// Test 2: Nil parser
	fmt.Println("      📝 Testing nil parser...")
	err = builder.AddRawTokenList("nil-parser", []byte(`{}`), "test", time.Now(), nil)
	if err != nil {
		fmt.Printf("      ✅ Correctly caught error: %v\n", err)
	}

	// Test 3: Invalid JSON
	fmt.Println("      📝 Testing invalid JSON...")
	invalidJSON := []byte(`{"name": invalid json}`)
	err = builder.AddRawTokenList("invalid", invalidJSON, "test", time.Now(), &parsers.StandardTokenListParser{})
	if err != nil {
		fmt.Printf("      ✅ Correctly caught error: %v\n", err)
	}

	fmt.Println("      🎯 Error handling validation complete!")
}
