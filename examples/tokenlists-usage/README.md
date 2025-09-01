# Token Lists Usage Example

This example demonstrates how to use the `tokenlists` package from the go-wallet-sdk to manage and query cryptocurrency token lists.

## Features Demonstrated

1. **Basic Setup**: Creating and configuring a TokensList service with default settings
2. **Token Queries**: Various methods to query tokens from the lists
3. **Chain-specific Queries**: Getting tokens for specific blockchain networks
4. **Address Lookups**: Finding specific tokens by their contract addresses
5. **Token List Management**: Working with different token lists and their metadata
6. **Utility Functions**: Using token key generation and parsing utilities
7. **Refresh Operations**: Manual refresh of token lists from remote sources
8. **Event Notifications**: Handling update notifications when token lists change

## What the TokensList Package Does

The `tokenlists` package provides a comprehensive solution for managing cryptocurrency token metadata across multiple blockchain networks:

The example works with a minimal configuration for demonstration purposes, a real configuration will provide more tokens/token lists.

## Key Components

### Token Structure
```go
type Token struct {
    CrossChainID string             // Unique identifier across chains
    ChainID      uint64             // Blockchain network ID
    Address      common.Address     // Contract address
    Decimals     uint               // Token decimal places
    Name         string             // Full token name
    Symbol       string             // Token symbol (e.g., "ETH", "USDT")
    LogoURI      string             // URL to token logo
    CustomToken  bool               // Whether it's a user-added token
}
```

### TokensList Interface
```go
type TokensList interface {
    Start(ctx context.Context, notifyCh chan struct{}) error
    Stop() error

    UniqueTokens() []*Token
    GetTokenByChainAddress(chainID uint64, addr common.Address) (*Token, bool)
    GetTokensByChain(chainID uint64) []*Token

    TokenLists() []*TokenList
    TokenList(id string) (*TokenList, bool)

    RefreshNow(ctx context.Context) error
    LastRefreshTime() (time.Time, error)
}
```

## Running the Example

```bash
cd examples/tokenlists-usage
go run main.go
```

## Output Example

The example will output information about:

- Total number of tokens across all supported chains
- Sample tokens with their metadata (name, symbol, address, etc.)
- Chain-specific token counts (Ethereum, Optimism, Arbitrum, etc.)
- Token lookup by specific contract address
- Available token lists and their sources
- Native tokens for each supported network
- Token key generation and parsing utilities
- Last refresh timestamp and manual refresh operations

## Configuration Options

The TokensList can be configured with:

- **Supported Chains**: Which blockchain networks to include
- **Token List Sources**: Remote URLs for fetching token lists
- **Refresh Intervals**: How often to check for updates
- **Privacy Settings**: Whether to fetch data from remote sources
- **Custom Storage**: Implement custom storage for caching token data
- **Logging**: Custom logger for debugging and monitoring
