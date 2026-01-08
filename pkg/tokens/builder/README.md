# Token Builder Package

The `builder` package provides functionality for building token collections by progressively adding multiple token lists from various sources, creating a unified collection of unique tokens across different blockchain networks.

## Use it when

- You want to merge multiple token lists into one deduplicated token set.
- You want to generate native tokens for selected chains.
- You want a stateful builder for token-list ingestion and merging.

## Key entrypoints

- `builder.New(chains, skippedTokenKeys)`
- `(*Builder).AddNativeTokenList()`
- `(*Builder).AddTokenList(...)` / `(*Builder).AddRawTokenList(...)`
- `(*Builder).GetTokens()` / `(*Builder).GetTokenLists()`

## Overview

The builder package is designed to:
- Build token collections incrementally by adding token lists from various sources
- Ensure token uniqueness across all added lists through automatic deduplication
- Generate native tokens for supported blockchain networks
- Parse and add raw token lists using configurable parsers
- Maintain both individual token lists and a unified token collection
- Follow the Builder pattern for stateful construction

## Key Features

- **Builder Pattern**: Start with empty state and progressively build up token collections
- **Deduplication**: Automatically prevents duplicate tokens using chain ID and address combinations
- **Token Filtering**: Skip specific tokens by their keys (useful for excluding invalid or unwanted tokens)
- **Native Token Support**: Generates native tokens (ETH, BNB, etc.) for supported chains
- **Multiple Formats**: Supports parsing various token list formats through pluggable parsers
- **Stateful Construction**: Maintains internal state between operations
- **Chain-Specific Logic**: Special handling for different blockchain networks
- **Incremental Building**: Add token lists one at a time or in batches

## Types

### Builder

The main struct that manages incremental token list building operations:

```go
type Builder struct {
    chains           []uint64                     // Supported chain IDs
    tokens           map[string]*types.Token      // Unified token collection (deduplicated)
    tokenLists       map[string]*types.TokenList  // Individual token lists by ID
    skippedTokenKeys map[string]bool              // Set of token keys to skip (for fast lookup)
}
```

### Constants

```go
const (
    NativeTokenListID = "native"  // ID for the native token list

    // Ethereum native token constants
    EthereumNativeCrossChainID = "eth-native"
    EthereumNativeSymbol       = "ETH"
    EthereumNativeName         = "Ethereum"

    // Binance Smart Chain native token constants
    BinanceSmartChainNativeCrossChainID = "bsc-native"
    BinanceSmartChainNativeSymbol       = "BNB"
    BinanceSmartChainNativeName         = "BNB"
)
```

### Errors

```go
var (
    ErrEmptyRawTokenList = fmt.Errorf("raw token list data is empty")
    ErrParserIsNil       = fmt.Errorf("parser is nil")
)
```

## API Reference

### Constructor

#### `New(chains []uint64, skippedTokenKeys []string) *Builder`

Creates a new Builder instance with empty token collections.

**Parameters:**
- `chains`: List of supported blockchain network IDs
- `skippedTokenKeys`: Optional list of token keys to exclude from token collections. Token keys are in the format `"{chainID}-{lowercaseAddress}"` (e.g., `"10-0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000"`). Pass `nil` or empty slice to skip no tokens.

**Returns:** New Builder instance ready for incremental construction

**Example:**
```go
chains := []uint64{1, 56, 10} // Ethereum, BSC, Optimism

// Without skipping any tokens
builder := builder.New(chains, nil)

// Skip specific tokens (e.g., invalid Optimism ETH token)
skippedKeys := []string{
    "10-0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000", // Optimism ETH with no value
}
builder := builder.New(chains, skippedKeys)

// Builder starts empty and builds up
```

### Getters

#### `GetTokens() map[string]*types.Token`

Returns the unified collection of unique tokens from all added token lists.

**Returns:** Map of token keys to Token objects

#### `GetTokenLists() map[string]*types.TokenList`

Returns all individual token lists indexed by their IDs.

**Returns:** Map of token list IDs to TokenList objects

### Building Operations

#### `AddNativeTokenList() error`

Generates and adds native tokens for all supported chains.

**Example:**
```go
err := builder.AddNativeTokenList()
if err != nil {
    log.Fatal(err)
}
```

**Supported Native Tokens:**
- **Ethereum & Ethereum-compatible chains**: ETH
- **Binance Smart Chain**: BNB

#### `AddTokenList(tokenListID string, tokenList *types.TokenList)`

Adds a parsed token list to the builder. Tokens with keys in the `skippedTokenKeys` list (provided during construction) will be automatically excluded from the token collection, but the token list itself will still be stored.

**Parameters:**
- `tokenListID`: Unique identifier for the token list
- `tokenList`: Parsed TokenList object

**Example:**
```go
tokenList := &types.TokenList{
    Name: "Uniswap Token List",
    Tokens: []*types.Token{
        // ... token objects
    },
}

builder.AddTokenList("uniswap", tokenList) // Tokens matching skippedTokenKeys are filtered out automatically
```

#### `AddRawTokenList(tokenListID string, raw []byte, sourceURL string, fetchedAt time.Time, parser parsers.TokenListParser) error`

Parses and adds raw token list data to the builder.

**Parameters:**
- `tokenListID`: Unique identifier for the token list
- `raw`: Raw JSON data of the token list
- `sourceURL`: Source URL where the list was fetched from
- `fetchedAt`: Timestamp when the list was fetched
- `parser`: Parser implementation for the specific token list format

**Returns:** Error if parsing fails or data is invalid

**Example:**
```go
rawData := []byte(`{"name": "Custom List", "tokens": [...]}`)
parser := &parsers.StandardTokenListParser{}

err := builder.AddRawTokenList(
    "custom-list",
    rawData,
    "https://example.com/tokens.json",
    time.Now(),
    parser,
)
if err != nil {
    log.Printf("Failed to add token list: %v", err)
}
```

## Deduplication Logic

The builder automatically deduplicates tokens using the token's key (combination of chain ID and address):

```go
// These would be deduplicated (same token on same chain)
token1 := &types.Token{ChainID: 1, Address: "0x123..."}
token2 := &types.Token{ChainID: 1, Address: "0x123..."} // duplicate

// These would NOT be deduplicated (different chains)
token3 := &types.Token{ChainID: 1, Address: "0x123..."}
token4 := &types.Token{ChainID: 56, Address: "0x123..."} // different chain

builder.AddTokenList("list1", &types.TokenList{Tokens: []*types.Token{token1}})
builder.AddTokenList("list2", &types.TokenList{Tokens: []*types.Token{token2}}) // ignored
builder.AddTokenList("list3", &types.TokenList{Tokens: []*types.Token{token3, token4}})

// Result: 2 unique tokens (token1 on chain 1, token4 on chain 56)
```

## Token Filtering

The builder supports filtering out specific tokens by their keys. This is useful for excluding invalid or unwanted tokens (e.g., tokens with no value or deprecated tokens).

**Token Key Format:**
Token keys follow the format: `"{chainID}-{lowercaseAddress}"`

**Example:**
```go
// Skip Optimism ETH token that has no value
skippedKeys := []string{
    "10-0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000", // Optimism ETH
}

builder := builder.New([]uint64{10}, skippedKeys)

// Add a token list containing the skipped token
tokenList := &types.TokenList{
    Tokens: []*types.Token{
        {ChainID: 10, Address: common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000")},
        {ChainID: 10, Address: common.HexToAddress("0x4200000000000000000000000000000000000006")},
    },
}

builder.AddTokenList("test-list", tokenList)

tokens := builder.GetTokens() // tokens will only contain the second token (0x4200...), the skipped token (0xdead...) is excluded
```

**Note:** Token lists are still stored in the builder even if all their tokens are filtered out. Only the tokens themselves are excluded from the unified token collection.

## Thread Safety

**The Builder struct is NOT thread-safe.** It performs direct map operations without synchronization, which can cause race conditions in concurrent environments.

### Recommendations:
- **Single-threaded usage**: Use the builder in a single goroutine
- **External synchronization**: If concurrent access is needed, wrap operations with mutex locks
- **Build-then-share pattern**: Complete all building operations, then share the results read-only

### Example with external synchronization:
```go
type SafeBuilder struct {
    builder *builder.Builder
    mu      sync.RWMutex
}

func (s *SafeBuilder) AddTokenList(id string, list *types.TokenList) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.builder.AddTokenList(id, list)
}

func (s *SafeBuilder) GetTokens() map[string]*types.Token {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.builder.GetTokens()
}
```