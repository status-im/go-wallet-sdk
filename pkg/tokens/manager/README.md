# Token Manager Package

The `manager` package provides a high-level, thread-safe interface for managing token collections from multiple sources with automatic refresh capabilities, state management, and comprehensive token operations.

## Use it when

- You want one high-level component to manage tokens across chains and sources.
- You need auto-refresh for remote token lists (with caching/backing stores).
- You need fast lookups (by chain/address/key) in a concurrency-safe way.

## Key entrypoints

- `manager.New(config, fetcher, contentStore, customTokenStore)`
- Lifecycle: `Manager.Start(...)` / `Manager.Stop()`
- Lookups: `GetTokenByChainAddress`, `GetTokensByChain`, `GetTokensByKeys`, `UniqueTokens`

## Overview

The manager package is designed to:
- **Centralize token management** across multiple blockchain networks
- **Merge tokens from various sources** (native, remote lists, local lists, custom tokens)
- **Provide thread-safe access** to token collections with optimized read performance
- **Support automatic refresh** of remote token lists with background fetching
- **Maintain deterministic ordering** for consistent token resolution
- **Handle errors gracefully** with fallback mechanisms

## Key Features

- **🔄 Automatic Refresh**: Background fetching and updating of remote token lists
- **🧵 Thread-Safe**: Concurrent read access with proper synchronization
- **📊 Multi-Source Merging**: Combines native, remote, local, and custom tokens
- **🎯 Rich Query API**: Find tokens by chain, address, or list ID
- **⚡ Optimized Performance**: RWMutex for concurrent reads, atomic operations where beneficial
- **🛡️ Error Resilience**: Graceful handling of network failures and data corruption
- **📋 Deterministic Processing**: Consistent token resolution order across runs
- **🔧 Flexible Configuration**: Pluggable parsers and storage backends

## Architecture

### Core Components

```go
type Manager interface {
    // Lifecycle Management
    Start(ctx context.Context, autoRefreshEnabled bool, notifyCh chan struct{}) error
    Stop() error

    // Auto-Refresh Control
    EnableAutoRefresh(ctx context.Context) error
    DisableAutoRefresh(ctx context.Context) error
    TriggerRefresh(ctx context.Context) error

    // Token Operations
    UniqueTokens() []*types.Token
    GetTokenByChainAddress(chainID uint64, addr common.Address) (*types.Token, bool)
    GetTokensByChain(chainID uint64) []*types.Token
    GetTokensByKeys(keys []string) ([]*types.Token, error)

    // Token List Operations
    TokenLists() []*types.TokenList
    TokenList(id string) (*types.TokenList, bool)
}
```

## Token Processing Order

The manager processes tokens in a **deterministic order** to ensure consistent resolution:

1. **🌐 Native Tokens**: Generated for each supported blockchain (ETH, BNB, etc.)
2. **📋 Main List**: Primary token list (remote if available, fallback to local)
3. **📄 Initial Lists**: Other configured lists (alphabetical order, remote preferred)
4. **☁️ Remote Lists**: Additional remote lists not in initial configuration
5. **👤 Custom Tokens**: User-added tokens with validation

This order ensures that **main lists take precedence** over supplementary lists, and **remote data is preferred** over local fallbacks.

## Configuration

### Parser Configuration

The manager supports **flexible parser configuration** through the `CustomParsers` field:

- **🎯 Explicit Parsers**: Specify custom parsers for specific token lists
- **🔧 Default Fallback**: Lists without custom parsers use `StandardTokenListParser`
- **⚡ Automatic Selection**: No need to specify parsers for standard token lists

```go
config := &manager.Config{
    CustomParsers: map[string]parsers.TokenListParser{
        "status-tokens":   &parsers.StatusTokenListParser{},    // Custom format
        "coingecko-data":  &parsers.CoinGeckoAllTokensParser{}, // CoinGecko format
        // "standard-list" omitted - will use StandardTokenListParser automatically
    },
}
```

### Token Filtering Configuration

The manager supports **token filtering** through the `SkippedTokenKeys` field:

- **🚫 Exclude Invalid Tokens**: Skip tokens with no value or deprecated tokens
- **🔑 Token Key Format**: Use `"{chainID}-{lowercaseAddress}"` format (e.g., `"10-0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000"`)
- **⚡ Automatic Filtering**: Filtered tokens are excluded from all token lists during building
- **📋 List Preservation**: Token lists are still stored even if all their tokens are filtered

```go
config := &manager.Config{
    SkippedTokenKeys: []string{
        "10-0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000", // Optimism ETH with no value
    },
}
```

**Note:** Token keys are matched case-insensitively. The filtering applies to all token sources (native tokens, remote lists, local lists, and custom tokens).

### Basic Configuration

```go
config := &manager.Config{
    MainListID: "uniswap-default",
    InitialLists: map[string][]byte{
        "uniswap-default": uniswapTokenListData,
        "compound":        compoundTokenListData,
        "custom-local":    customTokenListData,
    },
    CustomParsers: map[string]parsers.TokenListParser{
        "custom-local": &parsers.StatusTokenListParser{}, // Custom parser needed
        // "uniswap-default" and "compound" will use StandardTokenListParser automatically
    },
    Chains: []uint64{1, 56, 8453}, // Ethereum, BSC, Base
    SkippedTokenKeys: []string{
        "10-0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000", // Skip Optimism ETH with no value
    },
}

// Create HTTP fetcher for remote token list fetching
httpFetcher := fetcher.New(fetcher.DefaultConfig())

// Storage backends
contentStore := &MyContentStore{}     // Implements autofetcher.ContentStore
customTokenStore := &MyCustomStore{}  // Implements CustomTokenStore

manager, err := manager.New(config, httpFetcher, contentStore, customTokenStore)
if err != nil {
    log.Fatal(err)
}
```

### With Auto-Fetcher

```go
config := &manager.Config{
    MainListID: "uniswap-default",
    InitialLists: map[string][]byte{
        "uniswap-default": uniswapData,
    },
    CustomParsers: map[string]parsers.TokenListParser{
        // Optional: only specify if you need non-standard parsers
        // "uniswap-default" will use StandardTokenListParser automatically
    },
    Chains: []uint64{1, 56},
    SkippedTokenKeys: []string{
        // Optional: tokens to exclude from all token lists, format: "{chainID}-{lowercaseAddress}"
    },

    // Auto-fetcher configuration
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

## Usage Patterns

### Basic Usage

```go
// Start the manager
ctx := context.Background()
notifyCh := make(chan struct{}, 1)

err := manager.Start(ctx, true, notifyCh) // Enable auto-refresh
if err != nil {
    log.Fatal(err)
}
defer manager.Stop()

// Listen for token list updates
go func() {
    for range notifyCh {
        log.Println("Token lists updated!") // Refresh your UI
    }
}()
```

### Token Operations

```go
// Get all unique tokens across all chains
allTokens := manager.UniqueTokens()
fmt.Printf("Total tokens: %d\n", len(allTokens))

// Find a specific token
usdc := common.HexToAddress("0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB")
token, exists := manager.GetTokenByChainAddress(1, usdc)
if exists {
    fmt.Printf("Found: %s (%s)\n", token.Name, token.Symbol)
}

// Get all tokens for a specific chain
ethereumTokens := manager.GetTokensByChain(1)
fmt.Printf("Ethereum tokens: %d\n", len(ethereumTokens))

// Get tokens by their keys (efficient batch lookup)
keys := []string{
    "1-0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB", // USDC on Ethereum
    "56-0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d", // USDC on BSC
}
tokens, err := manager.GetTokensByKeys(keys)
if err != nil {
    log.Printf("Error: %v", err)
}
fmt.Printf("Retrieved %d tokens\n", len(tokens))

// Get tokens from a specific list
uniswapList, exists := manager.TokenList("uniswap-default")
if exists {
    fmt.Printf("Uniswap list has %d tokens\n", len(uniswapList.Tokens))
}

// Get all token lists
allLists := manager.TokenLists()
for _, list := range allLists {
    fmt.Printf("List: %s (%d tokens)\n", list.Name, len(list.Tokens))
}
```

### Auto-Refresh Management

```go
// Enable auto-refresh (requires auto-fetcher configuration)
err := manager.EnableAutoRefresh(ctx)
if err != nil {
    log.Printf("Failed to enable auto-refresh: %v", err)
}

// Disable auto-refresh
err = manager.DisableAutoRefresh(ctx)
if err != nil {
    log.Printf("Failed to disable auto-refresh: %v", err)
}

// Trigger immediate refresh
err = manager.TriggerRefresh(ctx)
if err != nil {
    log.Printf("Failed to trigger refresh: %v", err)
}
```

### Custom Token Integration

```go
type MyCustomTokenStore struct {
    tokens []*types.Token
}

func (s *MyCustomTokenStore) GetAll() ([]*types.Token, error) {
    // Return user's custom tokens
    return s.tokens, nil
}

// Add custom tokens
customStore.tokens = append(customStore.tokens, &types.Token{
    CrossChainID: "my-custom-token",
    ChainID:      1,
    Address:      common.HexToAddress("0x..."),
    Symbol:       "CUSTOM",
    Name:         "My Custom Token",
    Decimals:     18,
})
```

## Thread Safety

The manager is **fully thread-safe** and optimized for **concurrent access**:

### Read Operations (Concurrent Safe)
- `UniqueTokens()`
- `GetTokenByChainAddress()`
- `GetTokensByChain()`
- `GetTokensByKeys()`
- `TokenLists()`
- `TokenList()`

### Write Operations (Exclusive)
- `Start()`
- `Stop()`
- `EnableAutoRefresh()`
- `DisableAutoRefresh()`
- `TriggerRefresh()`
- Internal state updates

### Synchronization Strategy

The manager uses **two separate mutexes** for optimal concurrency:

```go
type manager struct {
    mu        sync.RWMutex  // Protects manager state (lifecycle, config)
    builderMu sync.RWMutex  // Protects builder access (token queries)
}
```

**Benefits:**
- **Separate read locks**: Token queries don't block lifecycle operations
- **Better concurrency**: Multiple readers can access different resources simultaneously
- **Reduced collisions**: State updates and token queries are independent

**Read operations** use `RLock()` allowing **multiple concurrent readers**.
**Write operations** use `Lock()` for **exclusive access** during updates.

## Error Handling

The manager implements **graceful error handling** with fallback mechanisms:

## Error Reference

```go
var (
    ErrContentStoreNotProvided                       = fmt.Errorf("content store not provided")
    ErrStoredTokenListIsEmpty                        = fmt.Errorf("stored token list is empty")
    ErrAutoFetcherNotProvided                        = fmt.Errorf("auto fetcher not provided")
    ErrAutoRefreshEnabledButNotifyChannelNotProvided = fmt.Errorf("auto refresh enabled but notify channel not provided")
    ErrManagerNotConfiguredForAutoRefresh            = fmt.Errorf("manager not configured for auto refresh")
    ErrNotFoundInInitialLists                        = fmt.Errorf("not found in initial lists")
)
```

## Dependencies

### Required Packages
- `github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher` - HTTP fetcher for remote token lists
- `github.com/status-im/go-wallet-sdk/pkg/tokens/builder` - Token collection building
- `github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher` - Background refresh
- `github.com/status-im/go-wallet-sdk/pkg/tokens/parsers` - Token list parsing
- `github.com/status-im/go-wallet-sdk/pkg/tokens/types` - Core types

## Testing

The package includes comprehensive tests covering:

- ✅ **Basic Operations**: All CRUD operations
- ✅ **Concurrency**: Race condition testing
- ✅ **Error Handling**: Network failures, data corruption
- ✅ **Auto-Refresh**: Background update mechanisms
- ✅ **Edge Cases**: Empty states, invalid configurations
- ✅ **Integration**: Multi-source token merging

Run tests with race detection:
```bash
go test -race ./pkg/tokens/manager/...
```

## See Also

- [Token Types](../types/README.md) - Core token data structures
- [Token Parsers](../parsers/README.md) - Parse different formats
- [Token Fetcher](../fetcher/README.md) - HTTP fetching with caching
- [Token AutoFetcher](../autofetcher/README.md) - Background refresh mechanism
- [Token Builder](../builder/README.md) - Incremental collection building
- [Balance Fetcher](../../balance/fetcher/README.md) - Fetch token balances

## Examples

- [Token Manager Example](../../../examples/token-manager/README.md) - Complete manager usage
- [Token Builder Example](../../../examples/token-builder/README.md) - Building token collections
- [Token Fetcher Example](../../../examples/token-fetcher/README.md) - Fetching token lists

