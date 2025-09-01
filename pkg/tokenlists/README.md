# TokenLists Package

The `tokenlists` package provides a comprehensive solution for managing and fetching token lists from various sources in a privacy-aware manner. It supports multiple token list formats, automatic refresh capabilities, and cross-chain token management.

## Features

- **Multi-source Support**: Fetch token lists from Status, Uniswap, CoinGecko, and custom sources
- **Privacy-aware**: Respects privacy settings to prevent unwanted network requests
- **Automatic Refresh**: Configurable automatic refresh intervals with ETag support
- **Cross-chain Support**: Manage tokens across multiple blockchain networks
- **Extensible**: Plugin-based parser system for custom token list formats
- **Thread-safe**: Concurrent access support with proper synchronization
- **Caching**: Built-in content caching with ETag support for efficient updates

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/status-im/go-wallet-sdk/pkg/tokenlists"
    "go.uber.org/zap"
)

func main() {
    // Create configuration
    config := tokenlists.DefaultConfig().
        WithChains([]uint64{1, 10, 137}). // Ethereum, Polygon, BSC
        WithRemoteListOfTokenListsURL("https://example.com/token-lists.json").
        WithAutoRefreshInterval(30*time.Minute, 3*time.Minute).
        WithLogger(zap.NewNop())

    // Create token list manager
    tl, err := tokenlists.NewTokensList(config)
    if err != nil {
        log.Fatal(err)
    }

    // Start the service
    ctx := context.Background()
    notifyCh := make(chan struct{}, 1)

    if err := tl.Start(ctx, notifyCh); err != nil {
        log.Fatal(err)
    }
    defer tl.Stop(ctx)

    // Wait for initial fetch
    select {
    case <-notifyCh:
        log.Println("Token lists updated")
    case <-time.After(10 * time.Second):
        log.Println("Timeout waiting for token lists")
    }

    // Get all unique tokens
    tokens := tl.UniqueTokens()
    log.Printf("Found %d unique tokens", len(tokens))

    // Get tokens for a specific chain
    ethereumTokens := tl.GetTokensByChain(1)
    log.Printf("Found %d tokens on Ethereum", len(ethereumTokens))
}
```

## Configuration

The package uses a builder pattern for configuration:

```go
config := &tokenlists.Config{
    // Required fields
    MainList:     []byte(`{"tokens": []}`),
    MainListID:   "status",
    Chains:       []uint64{1, 10, 137},

    // Optional fields with defaults
    RemoteListOfTokenListsURL: "https://example.com/lists.json",
    AutoRefreshInterval:       30 * time.Minute,
    AutoRefreshCheckInterval: 3 * time.Minute,

    // Custom components
    PrivacyGuard:                  tokenlists.NewDefaultPrivacyGuard(false),
    LastTokenListsUpdateTimeStore: tokenlists.NewDefaultLastTokenListsUpdateTimeStore(),
    ContentStore:                  tokenlists.NewDefaultContentStore(),
    Parsers:                       make(map[string]tokenlists.Parser),
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `MainList` | `[]byte` | - | Initial token list data |
| `MainListID` | `string` | - | Identifier for the main token list |
| `Chains` | `[]uint64` | - | Supported blockchain chain IDs |
| `RemoteListOfTokenListsURL` | `string` | - | URL to fetch list of token lists |
| `AutoRefreshInterval` | `time.Duration` | 30 min | How often to refresh token lists |
| `AutoRefreshCheckInterval` | `time.Duration` | 3 min | How often to check if refresh is needed |
| `PrivacyGuard` | `PrivacyGuard` | `NewDefaultPrivacyGuard(false)` | Privacy mode controller |
| `ContentStore` | `ContentStore` | `NewDefaultContentStore()` | Content caching store |
| `Parsers` | `map[string]Parser` | Built-in parsers | Token list format parsers |

## API Reference

### TokensList Interface

```go
type TokensList interface {
    // Lifecycle management
    Start(ctx context.Context, notifyCh chan struct{}) error
    Stop(ctx context.Context) error

    LastRefreshTime() (time.Time, error)
    RefreshNow(ctx context.Context) error

    // Privacy management
    PrivacyModeUpdated(ctx context.Context) error

    // Token queries
    UniqueTokens() []*Token
    GetTokenByChainAddress(chainID uint64, addr common.Address) (*Token, bool)
    GetTokensByChain(chainID uint64) []*Token

    // Token list queries
    TokenLists() []*TokenList
    TokenList(id string) (*TokenList, bool)
}
```

### Token Structure

```go
type Token struct {
    CrossChainID string         `json:"crossChainId"`
    ChainID      uint64         `json:"chainId"`
    Address      common.Address `json:"address"`
    Decimals     uint           `json:"decimals"`
    Name         string         `json:"name"`
    Symbol       string         `json:"symbol"`
    LogoURI      string         `json:"logoUri"`
    CustomToken  bool           `json:"custom"`
}
```

### TokenList Structure

```go
type TokenList struct {
    Name             string                 `json:"name"`
    Timestamp        string                 `json:"timestamp"`
    FetchedTimestamp string                 `json:"fetchedTimestamp"`
    Source           string                 `json:"source"`
    Version          Version                `json:"version"`
    Tags             map[string]interface{} `json:"tags"`
    LogoURI          string                 `json:"logoURI"`
    Keywords         []string               `json:"keywords"`
    Tokens           []*Token               `json:"tokens"`
}
```

## Privacy Mode

The package respects privacy settings to prevent unwanted network requests:

```go
// Enable privacy mode
config := tokenlists.DefaultConfig().
    WithPrivacyGuard(tokenlists.NewDefaultPrivacyGuard(true))

// Privacy mode prevents:
// - Automatic token list fetching
// - RefreshNow() calls from making network requests
// - Background refresh worker from running
```

## Supported Token List Formats

### Status Token List
- **Parser**: `StatusTokenListParser`
- **Format**: Status-specific JSON format

### Standard Token List Formats (uniswap, platform specific coingecko list and others use this format)
- **Parser**: `StandardTokenListParser`
- **Format**: Standard Token List format

### CoinGecko All Token List (doesn't contain decimals)
- **Parser**: `CoinGeckoAllTokensParser`
- **Format**: CoinGecko API format with chain mapping

## Custom Parsers (if the list doesn't follow the standard token list format)

Implement the `Parser` interface to support custom token list formats:

```go
type CustomParser struct{}

func (p *CustomParser) Parse(raw []byte, sourceURL string, fetchedAt time.Time) (*TokenList, error) {
    // Parse your custom format
    var customFormat struct {
        Name   string   `json:"name"`
        Tokens []*Token `json:"tokens"`
    }

    if err := json.Unmarshal(raw, &customFormat); err != nil {
        return nil, err
    }

    return &TokenList{
        Name:             customFormat.Name,
        Timestamp:        time.Now().Format(time.RFC3339),
        FetchedTimestamp: fetchedAt.Format(time.RFC3339),
        Source:           sourceURL,
        Tokens:           customFormat.Tokens,
    }, nil
}

// Register custom parser
config := tokenlists.DefaultConfig().
    WithParsers(map[string]tokenlists.Parser{
        "custom": &CustomParser{},
    })
```

## Error Handling

The package provides comprehensive error handling:

```go
tl, err := tokenlists.NewTokensList(config)
if err != nil {
    // Handle configuration errors
    log.Fatal(err)
}

if err := tl.Start(ctx, notifyCh); err != nil {
    // Handle startup errors
    log.Fatal(err)
}

// Handle refresh errors via notification channel
go func() {
    for range notifyCh {
        // Token lists updated successfully
        log.Println("Token lists refreshed")
    }
}()
```

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test ./pkg/tokenlists/...

# Run specific test
go test ./pkg/tokenlists -run TestTokensList_RefreshNow

# Run with verbose output
go test ./pkg/tokenlists/... -v
```
