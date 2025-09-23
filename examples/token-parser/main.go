package main

import (
	"fmt"
	"log"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

func main() {
	fmt.Println("🔍 Token Parser Example")
	fmt.Println("========================")

	supportedChains := []uint64{1, 56, 10, 137} // Ethereum, BSC, Optimism, Polygon

	// Example 1: Standard Uniswap-format token list
	fmt.Println("\n📋 Standard Token List Parser")
	fmt.Println("==============================")
	demonstrateStandardParser(supportedChains)

	// Example 2: Status-format token list
	fmt.Println("\n🟣 Status Token List Parser")
	fmt.Println("============================")
	demonstrateStatusParser(supportedChains)

	// Example 3: CoinGecko all tokens format
	fmt.Println("\n🦎 CoinGecko Token Parser")
	fmt.Println("==========================")
	demonstrateCoinGeckoParser(supportedChains)

	// Example 4: Status List of Token Lists
	fmt.Println("\n📚 Status List of Token Lists Parser")
	fmt.Println("====================================")
	demonstrateStatusListOfTokenListsParser()

	// Example 5: Error handling and validation
	fmt.Println("\n⚠️  Error Handling & Validation")
	fmt.Println("=================================")
	demonstrateErrorHandling(supportedChains)

	fmt.Println("\n✅ Token Parser examples completed!")
}

func demonstrateStandardParser(supportedChains []uint64) {
	// Sample Uniswap-format token list JSON
	standardTokenListJSON := `{
		"name": "Example Standard Token List",
		"timestamp": "2025-01-01T00:00:00Z",
		"version": {
			"major": 1,
			"minor": 0,
			"patch": 0
		},
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
				"chainId": 1,
				"address": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				"symbol": "USDT",
				"name": "Tether USD",
				"decimals": 6,
				"logoURI": "https://tokens.1inch.io/0xdac17f958d2ee523a2206206994597c13d831ec7.png"
			},
			{
				"chainId": 56,
				"address": "0x55d398326f99059fF775485246999027B3197955",
				"symbol": "USDT",
				"name": "Tether USD (BSC)",
				"decimals": 18,
				"logoURI": "https://tokens.1inch.io/0x55d398326f99059ff775485246999027b3197955.png"
			},
			{
				"chainId": 999,
				"address": "0x1234567890123456789012345678901234567890",
				"symbol": "UNSUPPORTED",
				"name": "Unsupported Chain Token",
				"decimals": 18
			}
		]
	}`

	parser := &parsers.StandardTokenListParser{}

	fmt.Printf("🔄 Parsing standard token list with %d chains supported...\n", len(supportedChains))

	tokenList, err := parser.Parse([]byte(standardTokenListJSON), supportedChains)
	if err != nil {
		log.Printf("❌ Failed to parse standard token list: %v", err)
		return
	}

	fmt.Printf("✅ Successfully parsed standard token list:\n")
	fmt.Printf("  📛 ID: %s\n", tokenList.ID)
	fmt.Printf("  📛 Name: %s\n", tokenList.Name)
	fmt.Printf("  📅 Timestamp: %s\n", tokenList.Timestamp)
	fmt.Printf("  🔗 Source: %s\n", tokenList.Source)
	fmt.Printf("  📊 Version: v%d.%d.%d\n", tokenList.Version.Major, tokenList.Version.Minor, tokenList.Version.Patch)
	fmt.Printf("  🪙 Total tokens in list: %d\n", len(tokenList.Tokens))

	// Show parsed tokens
	supportedTokens := 0
	for _, token := range tokenList.Tokens {
		supportedTokens++
		fmt.Printf("    • %s (%s) - Chain %d - %s\n",
			token.Name, token.Symbol, token.ChainID, token.Address.Hex())
	}

	fmt.Printf("  ✅ Supported tokens: %d (unsupported chains filtered out)\n", supportedTokens)
}

func demonstrateStatusParser(supportedChains []uint64) {
	// Sample Status-format token list JSON
	statusTokenListJSON := `{
		"name": "Status Token List",
		"timestamp": "2025-09-01T13:00:00.000Z",
		"version": {
			"major": 0,
			"minor": 0,
			"patch": 0
		},
		"tags": {},
		"logoURI": "https://res.cloudinary.com/dhgck7ebz/image/upload/f_auto,c_limit,w_64,q_auto/Brand/Logo%20Section/Mark/Mark_01",
		"keywords": [
			"status"
		],
		"tokens": [
			{
				"crossChainId": "status",
				"symbol": "SNT",
				"name": "Status",
				"decimals": 18,
				"logoURI": "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
				"contracts": {
					"1": "0x744d70fdbe2ba4cf95131626614a1763df805b9e",
					"10": "0x650af3c15af43dcb218406d30784416d64cfb6b2",
					"8453": "0x662015ec830df08c0fc45896fab726542e8ac09e",
					"42161": "0x707f635951193ddafbb40971a0fcaab8a6415160"
				}
			},
			{
				"crossChainId": "status-test-token",
				"symbol": "STT",
				"name": "Status Test Token",
				"decimals": 18,
				"logoURI": "https://assets.coingecko.com/coins/images/779/thumb/status.png?1548610778",
				"contracts": {
					"84532": "0xfdb3b57944943a7724fcc0520ee2b10659969a06",
					"11155111": "0xe452027cdef746c7cd3db31cb700428b16cd8e51",
					"1660990954": "0x1c3ac2a186c6149ae7cb4d716ebbd0766e4f898a"
				}
			},
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

	parser := &parsers.StatusTokenListParser{}

	fmt.Printf("🔄 Parsing Status token list (chain-grouped format)...\n")

	tokenList, err := parser.Parse([]byte(statusTokenListJSON), supportedChains)
	if err != nil {
		log.Printf("❌ Failed to parse Status token list: %v", err)
		return
	}

	fmt.Printf("✅ Successfully parsed Status token list:\n")
	fmt.Printf("  📛 ID: %s\n", tokenList.ID)
	fmt.Printf("  📛 Name: %s\n", tokenList.Name)
	fmt.Printf("  📅 Timestamp: %s\n", tokenList.Timestamp)
	fmt.Printf("  🔗 Source: %s\n", tokenList.Source)
	fmt.Printf("  📊 Version: v%d.%d.%d\n", tokenList.Version.Major, tokenList.Version.Minor, tokenList.Version.Patch)
	fmt.Printf("  🪙 Tokens found: %d\n", len(tokenList.Tokens))

	// Group tokens by chain for display
	chainTokens := make(map[uint64][]*types.Token)
	for _, token := range tokenList.Tokens {
		chainTokens[token.ChainID] = append(chainTokens[token.ChainID], token)
	}

	for chainID, tokens := range chainTokens {
		fmt.Printf("    ⛓️  Chain %d: %d tokens\n", chainID, len(tokens))
		for _, token := range tokens {
			fmt.Printf("      • %s (%s) - %s\n", token.Name, token.Symbol, token.Address.Hex())
		}
	}
}

func demonstrateCoinGeckoParser(supportedChains []uint64) {
	// Sample CoinGecko all tokens format JSON (simplified)
	coinGeckoJSON := `[
		{
			"id": "bitcoin",
			"symbol": "btc",
			"name": "Bitcoin",
			"platforms": {
				"ethereum": "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
				"binance-smart-chain": "0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c"
			}
		},
		{
			"id": "ethereum",
			"symbol": "eth",
			"name": "Ethereum",
			"platforms": {
				"ethereum": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
				"binance-smart-chain": "0x2170Ed0880ac9A755fd29B2688956BD959F933F8"
			}
		},
		{
			"id": "usd-coin",
			"symbol": "usdc",
			"name": "USD Coin",
			"platforms": {
				"ethereum": "0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB",
				"binance-smart-chain": "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d",
				"polygon-pos": "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
				"unsupported-network": "0x1234567890123456789012345678901234567890"
			}
		}
	]`

	parser := parsers.NewCoinGeckoAllTokensParser(parsers.DefaultCoinGeckoChainsMapper)

	fmt.Printf("🔄 Parsing CoinGecko all tokens format...\n")

	tokenList, err := parser.Parse([]byte(coinGeckoJSON), supportedChains)
	if err != nil {
		log.Printf("❌ Failed to parse CoinGecko tokens: %v", err)
		return
	}

	fmt.Printf("✅ Successfully parsed CoinGecko token list:\n")
	fmt.Printf("  📛 ID: %s\n", tokenList.ID)
	fmt.Printf("  📛 Name: %s\n", tokenList.Name)
	fmt.Printf("  📅 Timestamp: %s\n", tokenList.Timestamp)
	fmt.Printf("  🔗 Source: %s\n", tokenList.Source)
	fmt.Printf("  🪙 Tokens parsed: %d\n", len(tokenList.Tokens))

	// Group by chain for display
	chainTokens := make(map[uint64][]*types.Token)
	for _, token := range tokenList.Tokens {
		chainTokens[token.ChainID] = append(chainTokens[token.ChainID], token)
	}

	for chainID, tokens := range chainTokens {
		fmt.Printf("    ⛓️  Chain %d: %d tokens\n", chainID, len(tokens))
		for _, token := range tokens {
			fmt.Printf("      • %s (%s) - %s\n", token.Name, token.Symbol, token.Address.Hex())
		}
	}

	fmt.Printf("  💡 Note: CoinGecko format automatically generates cross-chain IDs\n")
}

func demonstrateStatusListOfTokenListsParser() {
	// Sample Status list of token lists JSON
	listOfTokenListsJSON := `{
		"timestamp": "2025-09-01T00:00:00.000Z",
		"version": {
			"major": 0,
			"minor": 1,
			"patch": 0
		},
		"tokenLists": [
			{
				"id": "uniswap",
				"sourceUrl": "https://ipfs.io/ipns/tokens.uniswap.org",
				"schema": "https://uniswap.org/tokenlist.schema.json"
			},
			{
				"id": "aave",
				"sourceUrl": "https://raw.githubusercontent.com/bgd-labs/aave-address-book/main/tokenlist.json"
			},
			{
				"id": "kleros",
				"sourceUrl": "https://t2crtokens.eth.link"
			},
			{
				"id": "superchain",
				"sourceUrl": "https://static.optimism.io/optimism.tokenlist.json"
			}
		]
	}`

	parser := &parsers.StatusListOfTokenListsParser{}

	fmt.Printf("🔄 Parsing Status list of token lists...\n")

	listOfTokenLists, err := parser.Parse([]byte(listOfTokenListsJSON))
	if err != nil {
		log.Printf("❌ Failed to parse list of token lists: %v", err)
		return
	}

	fmt.Printf("✅ Successfully parsed list of token lists:\n")
	fmt.Printf("  📅 Timestamp: %s\n", listOfTokenLists.Timestamp)
	fmt.Printf("  📊 Version: v%d.%d.%d\n",
		listOfTokenLists.Version.Major,
		listOfTokenLists.Version.Minor,
		listOfTokenLists.Version.Patch)
	fmt.Printf("  📋 Token lists found: %d\n", len(listOfTokenLists.TokenLists))

	fmt.Println("\n  📄 Individual token lists:")
	for i, listDetails := range listOfTokenLists.TokenLists {
		fmt.Printf("    %d. %s\n", i+1, listDetails.ID)
		fmt.Printf("       🔗 URL: %s\n", listDetails.SourceURL)
		fmt.Printf("       📋 Schema: %s\n", listDetails.Schema)
	}

	fmt.Printf("\n  💡 These %d lists can now be fetched using the token fetcher\n", len(listOfTokenLists.TokenLists))
}

func demonstrateErrorHandling(supportedChains []uint64) {
	fmt.Println("🧪 Testing various error scenarios:")

	// Test 1: Invalid JSON
	fmt.Println("\n1️⃣ Testing invalid JSON:")
	invalidJSON := `{"name": "Invalid List", "tokens": [invalid json}`
	parser := &parsers.StandardTokenListParser{}

	_, err := parser.Parse([]byte(invalidJSON), supportedChains)
	if err != nil {
		fmt.Printf("   ✅ Correctly caught JSON error: %v\n", err)
	} else {
		fmt.Printf("   ❌ Should have failed with JSON error\n")
	}

	// Test 2: Empty supported chains
	fmt.Println("\n4️⃣ Testing empty supported chains:")
	validJSON := `{
		"name": "Test List",
		"timestamp": "2025-01-01T00:00:00Z",
		"version": {"major": 1, "minor": 0, "patch": 0},
		"tokens": [
			{
				"chainId": 1,
				"address": "0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB",
				"symbol": "USDC",
				"name": "USD Coin",
				"decimals": 6
			}
		]
	}`

	tokenList, err := parser.Parse([]byte(validJSON), []uint64{})
	if err != nil {
		fmt.Printf("   ❌ Unexpected error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Parsed successfully with empty chains: %d tokens (all filtered)\n", len(tokenList.Tokens))
	}

	// Test 3: Chain filtering
	fmt.Println("\n5️⃣ Testing chain filtering:")
	multiChainJSON := `{
		"name": "Multi-Chain List",
		"timestamp": "2025-01-01T00:00:00Z",
		"version": {"major": 1, "minor": 0, "patch": 0},
		"tokens": [
			{
				"chainId": 1,
				"address": "0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB",
				"symbol": "USDC",
				"name": "USD Coin",
				"decimals": 6
			},
			{
				"chainId": 56,
				"address": "0x55d398326f99059fF775485246999027B3197955",
				"symbol": "USDT",
				"name": "Tether USD",
				"decimals": 18
			},
			{
				"chainId": 999,
				"address": "0x1234567890123456789012345678901234567890",
				"symbol": "UNKNOWN",
				"name": "Unknown Chain Token",
				"decimals": 18
			}
		]
	}`

	// Test with only Ethereum support
	ethereumOnly := []uint64{1}
	tokenList, err = parser.Parse([]byte(multiChainJSON), ethereumOnly)
	if err != nil {
		fmt.Printf("   ❌ Unexpected error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Chain filtering works: %d tokens (only Ethereum)\n", len(tokenList.Tokens))
		for _, token := range tokenList.Tokens {
			fmt.Printf("      • %s on chain %d\n", token.Symbol, token.ChainID)
		}
	}
}

// Additional helper functions for advanced usage examples

func demonstrateAdvancedParsing() {
	fmt.Println("\n🎯 Advanced Parsing Techniques")
	fmt.Println("===============================")

	// Example: Custom parser selection based on content
	fmt.Println("\n📝 Parser Selection Strategy:")
	fmt.Println("   💡 Tips for choosing the right parser:")
	fmt.Println("      • Standard format: Most common, used by Uniswap, Compound, etc.")
	fmt.Println("      • Status format: Chain-grouped tokens, more efficient for multi-chain")
	fmt.Println("      • CoinGecko format: Cross-platform tokens with automatic cross-chain IDs")
	fmt.Println("      • Auto-detection: Check JSON structure to select parser automatically")

	// Example: Performance considerations
	fmt.Println("\n⚡ Performance Considerations:")
	fmt.Println("   • Standard parser: Fast, straightforward deserialization")
	fmt.Println("   • Status parser: Slightly slower due to chain grouping logic")
	fmt.Println("   • CoinGecko parser: Slower due to cross-platform mapping")
	fmt.Println("   • Memory usage: ~1MB per 1000 tokens during parsing")

	// Example: Validation strategies
	fmt.Println("\n🔍 Token Validation:")
	fmt.Println("   • Address format: Checksummed Ethereum addresses")
	fmt.Println("   • Symbol validation: Non-empty, reasonable length")
	fmt.Println("   • Decimals range: Typically 0-18 for ERC-20 tokens")
	fmt.Println("   • Chain ID validation: Must be in supported chains list")
}

func demonstrateParserComparison() {
	fmt.Println("\n📊 Parser Comparison")
	fmt.Println("====================")

	fmt.Printf("%-20s %-15s %-15s %-20s %-15s\n", "Parser", "Format", "Performance", "Use Case", "Cross-Chain")
	fmt.Printf("%-20s %-15s %-15s %-20s %-15s\n", "------", "------", "-----------", "--------", "-----------")
	fmt.Printf("%-20s %-15s %-15s %-20s %-15s\n", "Standard", "Uniswap", "Fast", "General purpose", "Manual")
	fmt.Printf("%-20s %-15s %-15s %-20s %-15s\n", "Status", "Chain-grouped", "Medium", "Multi-chain apps", "Manual")
	fmt.Printf("%-20s %-15s %-15s %-20s %-15s\n", "CoinGecko", "Platform-based", "Slow", "Cross-platform", "Automatic")
}
