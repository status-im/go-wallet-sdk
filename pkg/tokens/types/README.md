# Token Types

The `types` package provides core data structures for representing tokens and token lists in a unified format. These types serve as the foundation for all token-related operations across the SDK.

## Use it when

- You need a common `Token` / `TokenList` representation across token sources.
- You need stable token keys for indexing and lookup.
- You need helpers like native-token detection.

## Key entrypoints

- `types.Token`, `types.TokenList`, `types.Version`
- `types.TokenKey(chainID, address)` and `(*Token).Key()`
- `types.ChainAndAddressFromTokenKey(key)`
- `(*Token).IsNative()`

## Overview

This package defines two main types:
- **`Token`**: Represents an individual token with its metadata and blockchain information
- **`TokenList`**: Represents a collection of tokens with metadata about the list itself

## Types

### Token

The `Token` struct represents an individual cryptocurrency token with comprehensive metadata:

```go
type Token struct {
    CrossChainID string             `json:"crossChainId"` // Cross-chain identifier (optional)
    ChainID      uint64             `json:"chainId"`      // Blockchain network ID
    Address      gethcommon.Address `json:"address"`      // Contract address
    Decimals     uint               `json:"decimals"`     // Number of decimal places
    Name         string             `json:"name"`         // Full token name
    Symbol       string             `json:"symbol"`       // Token symbol/ticker
    LogoURI      string             `json:"logoUri"`      // URL to token logo
    CustomToken  bool               `json:"custom"`       // Whether this is a custom user token
}
```

#### Key Features

- **Cross-Chain Support**: `CrossChainID` allows grouping tokens across different blockchains
- **Address Validation**: Uses `gethcommon.Address` for type-safe address handling
- **Custom Token Flag**: Distinguishes between official and user-added tokens
- **Rich Metadata**: Includes name, symbol, decimals, and logo information

### TokenList

The `TokenList` struct represents a collection of tokens with metadata about the list:

```go
type TokenList struct {
    ID               string                 `json:"id"`               // Token list ID
    Name             string                 `json:"name"`             // Human-readable list name
    Timestamp        string                 `json:"timestamp"`        // When list was last updated
    FetchedTimestamp string                 `json:"fetchedTimestamp"` // When list was fetched
    Source           string                 `json:"source"`           // Source URL or identifier
    Version          Version                `json:"version"`          // Semantic version
    Tags             map[string]interface{} `json:"tags"`             // Custom metadata tags
    LogoURI          string                 `json:"logoUri"`          // List logo URL
    Keywords         []string               `json:"keywords"`         // Search keywords
    Tokens           []*Token               `json:"tokens"`           // List of tokens
}
```

### Version

The `Version` struct follows semantic versioning:

```go
type Version struct {
    Major int `json:"major"` // Major version number
    Minor int `json:"minor"` // Minor version number
    Patch int `json:"patch"` // Patch version number
}
```

## Key Generation and Indexing

### Token Keys

The package provides utilities for creating unique token identifiers:

```go
// Create a token key manually
key := types.TokenKey(1, common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"))
fmt.Println(key) // "1-0xa0b86a33e6441e8c"

// Get key from token instance
token := types.Token{ChainID: 1, Address: common.HexToAddress("0xA0b86a33E6441e8C")}
key = token.Key()

// Parse chain ID and address from key
chainID, address, ok := types.ChainAndAddressFromTokenKey(key)
if ok {
    fmt.Printf("Chain: %d, Address: %s\n", chainID, address.Hex())
}
```

### Token Key Format

Token keys follow the format: `{chainId}-{lowercaseAddress}`

Examples:
- Ethereum USDC: `1-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48`
- Optimism USDC: `10-0x0b2c639c533813f4aa9d7837caf62653d097ff85`
- Base USDC: `8453-0x833589fcd6edb6e08f4c7c32d4f71b54bda02913`

## Native Token Detection

### IsNative() Method

The `IsNative()` method identifies native blockchain tokens (ETH, BNB, etc.):

```go
// Native token (zero address)
ethToken := types.Token{
    ChainID: 1,
    Address: common.Address{}, // Zero address
    Name:    "Ether",
    Symbol:  "ETH",
    Decimals: 18,
}

if ethToken.IsNative() {
    fmt.Println("This is ETH - the native token of Ethereum")
}

// ERC-20 token (non-zero address)
usdcToken := types.Token{
    ChainID: 1,
    Address: common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
    Name:    "USD Coin",
    Symbol:  "USDC",
    Decimals: 6,
}

if !usdcToken.IsNative() {
    fmt.Println("This is an ERC-20 token")
}
```