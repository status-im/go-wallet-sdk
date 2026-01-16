# Token Manager Example

This example demonstrates how to use the `pkg/tokens/manager` package for comprehensive token management across multiple blockchain networks with support for various token sources and custom tokens.

## What it demonstrates

- **Complete token management**: High-level interface for token collections
- **Multi-source integration**: Native tokens, remote lists, local lists, custom tokens
- **Thread-safe operations**: Concurrent access to token data
- **Rich query capabilities**: Find tokens by chain, address, or list ID
- **Custom token support**: Add and manage user-defined tokens
- **State management**: Automatic token deduplication and list management
- **Error resilience**: Graceful handling of failures with fallbacks

## Run

```bash
cd examples/token-manager
go run main.go
```

## What This Example Shows

### 1. Manager Configuration

```go
initialLists := map[string][]byte{
    "uniswap-default": uniswapTokenListData,
    "compound":        compoundTokenListData,
}

config := &manager.Config{
    MainListID: "uniswap-default",
    InitialListIDs:      manager.InitialListIDsFromMap(initialLists),
    InitialListProvider: manager.StaticInitialListProvider(initialLists),
    CustomParsers: map[string]parsers.TokenListParser{
        "status": &parsers.StatusTokenListParser{},
    },
    Chains: []uint64{1, 56, 10, 137}, // Multiple blockchain networks
    AutoFetcherConfig: &autofetcher.ConfigRemoteListOfTokenLists{
        Config: autofetcher.Config{
            AutoRefreshInterval:      24 * time.Hour,
            AutoRefreshCheckInterval: time.Hour,
        },
        RemoteListOfTokenListsFetchDetails: types.ListDetails{
            ID:        "status-lists",
            SourceURL: "https://prod.market.status.im/static/lists.json",
            Schema:    fetcher.ListOfTokenListsSchema,
        },
        RemoteListOfTokenListsParser: &parsers.StatusListOfTokenListsParser{},
    },
}
```

### 2. HTTP Fetcher Setup

The manager requires a fetcher for retrieving remote token lists:

```go
// Create HTTP fetcher with default configuration
httpFetcher := fetcher.New(fetcher.DefaultConfig())

// Or with custom configuration
customConfig := fetcher.Config{
    Timeout:            10 * time.Second,
    IdleConnTimeout:    90 * time.Second,
    MaxIdleConns:       10,
    DisableCompression: false,
}
httpFetcher := fetcher.New(customConfig)
```

### 3. Storage Backend Implementation

The example includes in-memory implementations of required storage interfaces:

- **ContentStore**: For caching remote token lists
- **CustomTokenStore**: For managing user-defined tokens

### 4. Creating the Manager

```go
// Create manager with all dependencies
tokenManager, err := manager.New(
    config,
    httpFetcher,      // Fetcher for remote token lists
    contentStore,     // ContentStore implementation
    customTokenStore, // CustomTokenStore implementation
)
if err != nil {
    log.Fatalf("Failed to create token manager: %v", err)
}

// Start the manager
ctx := context.Background()
notifyCh := make(chan struct{}, 1)
if err := tokenManager.Start(ctx, false, notifyCh); err != nil {
    log.Fatalf("Failed to start token manager: %v", err)
}
defer tokenManager.Stop()
```

### 5. Token Operations

```go
// Get all unique tokens across all chains
allTokens := tokenManager.UniqueTokens()

// Find specific token by chain and address
token, exists := tokenManager.GetTokenByChainAddress(chainID, address)

// Get all tokens for a specific blockchain
chainTokens := tokenManager.GetTokensByChain(chainID)

// Get tokens by their keys
keys := []string{"1-0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB", // USDC
  "1-0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT
}
tokens := tokenManager.GetTokensByKeys(keys)

// Access token lists
allLists := tokenManager.TokenLists()
specificList, exists := tokenManager.TokenList("uniswap-default")
```

### 6. Custom Token Management

```go
// Add custom tokens
customTokens := []*types.Token{
    {
        CrossChainID: "my-custom-token",
        ChainID:      1,
        Address:      common.HexToAddress("0x1111..."),
        Symbol:       "CUSTOM",
        Name:         "My Custom Token",
        Decimals:     18,
    },
}
customStore.setTokens(customTokens)
```

## Example Output

```
🎯 Token Manager Example
==========================
➕ Added custom tokens for demonstration

📊 Token Operations Demo
=========================
📈 Total unique tokens: 12

🔍 Sample Tokens by Category:

  🌐 Native Tokens:
    • Ethereum (ETH) on Chain 1
    • BNB (BNB) on Chain 56

  🪙 ERC-20 Tokens (sample):
    • USD Coin (USDC) on Chain 1 - 0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB
    • Tether USD (USDT) on Chain 1 - 0xdAC17F958D2ee523a2206206994597C13D831ec7
    • Compound (COMP) on Chain 1 - 0xc00e94Cb662C3520282E6f5717214004A7f26888

🔎 Token Search Examples:
  ✅ Found USDC: USD Coin (USDC) - 6 decimals
  🌾 BSC tokens: 3 found
  ⟠ Ethereum tokens: 7 found

📋 Token Lists:
  📄 Native tokens: 4 tokens
  📄 Uniswap Default List: 3 tokens
  📄 Compound Token List: 2 tokens
  📄 Custom tokens: 2 tokens
  🌐 Native token list contains 4 tokens
  👤 Custom token list contains 2 tokens

✨ Token operations completed successfully!

✨ Token Manager is running. Press Ctrl+C to exit.
```

## Supported Networks

The example demonstrates multi-chain support:

- **Ethereum Mainnet** (Chain ID: 1)
- **BSC (Binance Smart Chain)** (Chain ID: 56)
- **Optimism** (Chain ID: 10)
- **Polygon** (Chain ID: 137)

## Key Concepts

### Token Processing Order

The manager processes tokens in a deterministic order:

1. **Native Tokens**: Generated for each supported chain (ETH, BNB, etc.)
2. **Main List**: Primary token list specified in configuration
3. **Additional Lists**: Other configured lists processed alphabetically
4. **Custom Tokens**: User-defined tokens added through CustomTokenStore

### Thread Safety

All operations are thread-safe and optimized for concurrent access:
- **Read operations** (GetTokens, GetTokensByChain) allow multiple concurrent readers
- **Write operations** (Start, Stop) use exclusive locks
- **State updates** are atomic and consistent

### Error Handling

The manager implements graceful error handling:
- Falls back to cached data when remote fetches fail
- Continues processing other sources if one fails
- Maintains core functionality even with partial failures

## Production Considerations

### Storage Backends

In production, implement persistent storage:

```go
// Database-backed content store
type dbContentStore struct {
    db *sql.DB
}

func (s *dbContentStore) Get(id string) (autofetcher.Content, error) {
    // Query database for cached token list
}

// File-based custom token store
type fileCustomTokenStore struct {
    filepath string
}

func (s *fileCustomTokenStore) GetAll() ([]*types.Token, error) {
    // Read custom tokens from file
}
```

### Auto-Refresh Management

Control auto-refresh dynamically at runtime:

```go
// Enable auto-refresh
if err := tokenManager.EnableAutoRefresh(ctx); err != nil {
    log.Printf("Failed to enable auto refresh: %v", err)
}

// Disable auto-refresh
if err := tokenManager.DisableAutoRefresh(ctx); err != nil {
    log.Printf("Failed to disable auto refresh: %v", err)
}

// Manually trigger a refresh
if err := tokenManager.TriggerRefresh(ctx); err != nil {
    log.Printf("Failed to trigger refresh: %v", err)
}
```

### Monitoring and Observability

```go
// Listen for token list updates
go func() {
    for {
        select {
        case <-notifyCh:
            // Log update, trigger cache refresh, notify clients
            log.Println("Token lists updated")
            metrics.IncrementTokenListUpdates()
        case <-ctx.Done():
            return
        }
    }
}()
```

## Dependencies

- `github.com/status-im/go-wallet-sdk/pkg/tokens/manager` - High-level token management
- `github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher` - HTTP fetcher for remote token lists
- `github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher` - Automatic background fetching
- `github.com/status-im/go-wallet-sdk/pkg/tokens/parsers` - Token list parsing
- `github.com/status-im/go-wallet-sdk/pkg/tokens/types` - Core types
- `github.com/ethereum/go-ethereum/common` - Ethereum address types

## Integration Patterns

### Wallet Integration

```go
// Wallet service integration
type WalletService struct {
    tokenManager manager.Manager
}

func (s *WalletService) GetUserTokenBalances(userAddr common.Address) ([]TokenBalance, error) {
    allTokens := s.tokenManager.UniqueTokens()
    // Fetch balances for all tokens...
}
```

This example provides a comprehensive introduction to the token management system and demonstrates its integration in a realistic application context.