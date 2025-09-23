package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/common"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/manager"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

func main() {
	fmt.Println("🎯 Token Manager Example")
	fmt.Println("==========================")

	// Initialize storage backends
	contentStore := newMemoryContentStore()
	customTokenStore := newMemoryCustomTokenStore()

	// Add some custom tokens before creating the manager
	addCustomTokensExample(customTokenStore)

	// Setup configuration
	config := createExampleConfig()

	// Create HTTP fetcher for remote token list fetching
	httpFetcher := fetcher.New(fetcher.DefaultConfig())

	// Create manager
	tokenManager, err := manager.New(config, httpFetcher, contentStore, customTokenStore)
	if err != nil {
		log.Fatalf("Failed to create token manager: %v", err)
	}

	// Create notification channel for auto refresh updates
	notifyCh := make(chan struct{}, 1)

	// Start the manager without auto refresh initially but with notification channel
	ctx := context.Background()
	if err := tokenManager.Start(ctx, false, notifyCh); err != nil {
		log.Fatalf("Failed to start token manager: %v", err)
	}
	defer tokenManager.Stop()

	fmt.Println("\nStarting an already started manager should not have any effect")
	if err := tokenManager.Start(ctx, false, notifyCh); err != nil {
		log.Fatalf("Starting an already started manager should not fail: %v", err)
	} else {
		fmt.Println("✅ Starting an already started manager didn't fail, does nothing")
	}

	// Custom tokens were added before manager creation

	// Demonstrate token operations before auto refresh
	fmt.Println("\n📊 Token State BEFORE Auto Refresh")
	fmt.Println("=====================================")
	demonstrateTokenOperations(tokenManager)

	// Now enable auto refresh to demonstrate dynamic updates
	fmt.Println("\n🔄 Enabling Auto Refresh...")
	fmt.Println("============================")

	// Enable auto refresh
	if err := tokenManager.EnableAutoRefresh(ctx); err != nil {
		log.Printf("Note: Auto refresh failed (expected in example): %v", err)
	} else {
		fmt.Println("✅ Auto refresh enabled")
	}

	ctxWithTimeout, ctxWithTimeoutCancel := context.WithTimeout(ctx, 5*time.Second)
	defer ctxWithTimeoutCancel()

loop:
	for {
		select {
		case <-notifyCh:
			fmt.Println("📢 Token lists updated via auto refresh!")
			break loop
		case <-ctxWithTimeout.Done():
			fmt.Println("❌ Timeout reached, stopping listener")
			return
		}
	}

	// Show token state after enabling auto refresh
	fmt.Println("\n📊 Token State AFTER Auto Refresh Enabled")
	fmt.Println("==========================================")
	demonstrateTokenOperations(tokenManager)

	fmt.Println("\nEnabling an already enabled auto refresh should not have any effect")
	if err := tokenManager.EnableAutoRefresh(ctx); err != nil {
		log.Fatalf("Enabling an already enabled auto refresh should not fail: %v", err)
	} else {
		fmt.Println("✅ Enabling an already enabled auto refresh didn't fail, does nothing")
	}

	// Disable auto refresh
	fmt.Println("\n🛑 Disabling Auto Refresh...")
	if err := tokenManager.DisableAutoRefresh(ctx); err != nil {
		log.Printf("Failed to disable auto refresh: %v", err)
	} else {
		fmt.Println("✅ Auto refresh disabled")
	}

	fmt.Println("\nDisabling an already disabled auto refresh should not have any effect")
	if err := tokenManager.DisableAutoRefresh(ctx); err != nil {
		log.Fatalf("Disabling an already disabled auto refresh should not fail: %v", err)
	} else {
		fmt.Println("✅ Disabling an already disabled auto refresh didn't fail, does nothing")
	}

	fmt.Println("\n✨ Token Manager example completed successfully!")
	fmt.Println("👋 Stopping Token Manager...")
}

func createExampleConfig() *manager.Config {
	// Sample token lists (in production, these would come from files or URLs)
	uniswapTokenList := `{
		"name": "Uniswap Default List",
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
			}
		]
	}`

	compoundTokenList := `{
		"name": "Compound Token List",
		"timestamp": "2025-01-01T00:00:00Z",
		"version": {"major": 1, "minor": 0, "patch": 0},
		"tokens": [
			{
				"chainId": 1,
				"address": "0xc00e94Cb662C3520282E6f5717214004A7f26888",
				"symbol": "COMP",
				"name": "Compound",
				"decimals": 18,
				"logoURI": "https://tokens.1inch.io/0xc00e94cb662c3520282e6f5717214004a7f26888.png"
			},
			{
				"chainId": 1,
				"address": "0x39AA39c021dfbaE8faC545936693aC917d5E7563",
				"symbol": "cUSDC",
				"name": "Compound USD Coin",
				"decimals": 8,
				"logoURI": "https://tokens.1inch.io/0x39aa39c021dfbae8fac545936693ac917d5e7563.png"
			}
		]
	}`

	return &manager.Config{
		AutoFetcherConfig: &autofetcher.ConfigRemoteListOfTokenLists{
			Config: autofetcher.Config{
				AutoRefreshInterval:      5 * time.Second,
				AutoRefreshCheckInterval: 1 * time.Second,
			},
			RemoteListOfTokenListsFetchDetails: types.ListDetails{
				ID:        "status-lists",
				SourceURL: "https://prod.market.status.im/static/lists.json",
				Schema:    fetcher.ListOfTokenListsSchema,
			},
			RemoteListOfTokenListsParser: &parsers.StatusListOfTokenListsParser{},
		},
		MainListID: "uniswap-default",
		InitialLists: map[string][]byte{
			"uniswap-default": []byte(uniswapTokenList),
			"compound":        []byte(compoundTokenList),
		},
		CustomParsers: map[string]parsers.TokenListParser{
			"status": &parsers.StatusTokenListParser{},
		},
		Chains: []uint64{common.EthereumMainnet, common.BSCMainnet, common.OptimismMainnet, common.ArbitrumMainnet},
	}
}

func addCustomTokensExample(customStore *memoryCustomTokenStore) {
	// Add some custom tokens for demonstration
	customTokens := []*types.Token{
		{
			CrossChainID: "",
			ChainID:      1,
			Address:      gethcommon.HexToAddress("0x1111111111111111111111111111111111111111"),
			Symbol:       "CUSTOM",
			Name:         "My Custom Token",
			Decimals:     18,
			LogoURI:      "https://example.com/custom-token.png",
		},
		{
			CrossChainID: "",
			ChainID:      56,
			Address:      gethcommon.HexToAddress("0x2222222222222222222222222222222222222222"),
			Symbol:       "CUSTOM2",
			Name:         "Another Custom Token",
			Decimals:     8,
			LogoURI:      "https://example.com/another-token.png",
		},
	}

	customStore.setTokens(customTokens)
	fmt.Println("➕ Added custom tokens for demonstration")
}

func demonstrateTokenOperations(tokenManager manager.Manager) {
	fmt.Println("\n📊 Token Operations Demo")
	fmt.Println("=========================")

	// Get all unique tokens
	allTokens := tokenManager.UniqueTokens()
	fmt.Printf("📈 Total unique tokens: %d\n", len(allTokens))

	// Show sample tokens by category
	fmt.Println("\n🔍 Sample Tokens by Category:")

	// Native tokens
	fmt.Println("\n  🌐 Native Tokens:")
	for _, token := range allTokens {
		if token.IsNative() {
			fmt.Printf("    • %s (%s) on Chain %d\n", token.Name, token.Symbol, token.ChainID)
		}
	}

	// ERC-20 tokens (first few)
	fmt.Println("\n  🪙 ERC-20 Tokens (sample):")
	count := 0
	for _, token := range allTokens {
		if !token.IsNative() && count < 5 {
			fmt.Printf("    • %s (%s) on Chain %d - %s\n",
				token.Name, token.Symbol, token.ChainID, token.Address.Hex())
			count++
		}
	}

	// Search for specific tokens
	fmt.Println("\n🔎 Token Search Examples:")

	// Find USDC on Ethereum
	usdcAddr := gethcommon.HexToAddress("0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB")
	if token, exists := tokenManager.GetTokenByChainAddress(common.EthereumMainnet, usdcAddr); exists {
		fmt.Printf("  ✅ Found USDC: %s (%s) - %d decimals\n",
			token.Name, token.Symbol, token.Decimals)
	}

	// Get all tokens on BSC
	bscTokens := tokenManager.GetTokensByChain(common.BSCMainnet)
	fmt.Printf("  🌾 BSC tokens: %d found\n", len(bscTokens))

	// Get all tokens on Ethereum
	ethTokens := tokenManager.GetTokensByChain(common.EthereumMainnet)
	fmt.Printf("  ⟠ Ethereum tokens: %d found\n", len(ethTokens))

	// Show token lists
	fmt.Println("\n📋 Token Lists:")
	allLists := tokenManager.TokenLists()
	for _, list := range allLists {
		warningText := ""
		tokensCount := len(list.Tokens)
		if tokensCount == 0 {
			warningText = " (!could be that your config doesn't support any chain from this list)"
		}
		fmt.Printf("  📄 %s: %d tokens  %s\n", list.Name, tokensCount, warningText)
	}

	// Get specific token list
	tokenListIDs := []string{"native", "status", "uniswap", "unexisting"}
	for _, tokenListID := range tokenListIDs {

		fmt.Printf("\n📋 Search for token list with ID: %s\n", tokenListID)
		if nativeList, exists := tokenManager.TokenList(tokenListID); exists {
			fmt.Printf("  🌐 Found - %s contains %d tokens\n", nativeList.Name, len(nativeList.Tokens))
		} else {
			fmt.Printf("  🌐 Not found\n")
		}
	}

	fmt.Println("\n✨ Token operations completed successfully!")
}

// Memory-based implementations for the example
type memoryContentStore struct {
	mu   sync.RWMutex
	data map[string]autofetcher.Content
}

func newMemoryContentStore() *memoryContentStore {
	return &memoryContentStore{
		data: make(map[string]autofetcher.Content),
	}
}

func (m *memoryContentStore) GetEtag(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if content, exists := m.data[id]; exists {
		return content.Etag, nil
	}
	return "", fmt.Errorf("not found")
}

func (m *memoryContentStore) Get(id string) (autofetcher.Content, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if content, exists := m.data[id]; exists {
		return content, nil
	}
	return autofetcher.Content{}, fmt.Errorf("not found")
}

func (m *memoryContentStore) Set(id string, content autofetcher.Content) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[id] = content
	return nil
}

func (m *memoryContentStore) GetAll() (map[string]autofetcher.Content, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]autofetcher.Content)
	for k, v := range m.data {
		result[k] = v
	}
	return result, nil
}

type memoryCustomTokenStore struct {
	tokens []*types.Token
	mu     sync.RWMutex
}

func newMemoryCustomTokenStore() *memoryCustomTokenStore {
	return &memoryCustomTokenStore{
		tokens: make([]*types.Token, 0),
	}
}

func (m *memoryCustomTokenStore) GetAll() ([]*types.Token, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*types.Token, len(m.tokens))
	copy(result, m.tokens)
	return result, nil
}

func (m *memoryCustomTokenStore) setTokens(tokens []*types.Token) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens = tokens
}
