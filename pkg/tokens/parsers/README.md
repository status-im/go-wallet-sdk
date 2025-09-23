# Token List Parsers

The `parsers` package provides implementations for parsing token lists from various formats and sources. It supports multiple token list standards and converts them into a unified internal format for consistent processing.

## Overview

The parsers package provides two main types of parsers:
1. **Token List Parsers**: Parse individual token lists from various providers
2. **List of Token Lists Parsers**: Parse metadata about collections of token lists

## Features

- **Multiple Format Support**: Parse token lists from different providers and standards
- **Chain Filtering**: Filter tokens by supported blockchain networks
- **Address Validation**: Validate Ethereum addresses and skip invalid entries
- **Cross-Chain Support**: Handle tokens that exist across multiple blockchains
- **Unified Output**: Convert all formats to a consistent internal structure
- **Extensible Design**: Easy to add new parsers for additional formats
- **Metadata Parsing**: Parse lists of token lists for discovery and management

## Parser Interfaces

### TokenListParser Interface

All token list parsers implement this interface:

```go
type TokenListParser interface {
    Parse(raw []byte, supportedChains []uint64) (*types.TokenList, error)
}
```

### ListOfTokenListsParser Interface

For parsing metadata about token list collections:

```go
type ListOfTokenListsParser interface {
    Parse(raw []byte) (*types.ListOfTokenLists, error)
}
```

## Token List Parsers

### 1. Standard Token List Parser (`StandardTokenListParser`)

Parses token lists following the [Token Lists standard](https://tokenlists.org/) used by Uniswap and many other DeFi protocols.

**Format**:
```json
{
  "name": "My Token List",
  "timestamp": "2023-01-01T00:00:00.000Z",
  "version": {
    "major": 1,
    "minor": 0,
    "patch": 0
  },
  "tags": {},
  "logoURI": "https://example.com/logo.png",
  "keywords": ["default", "verified"],
  "tokens": [
    {
      "chainId": 1,
      "address": "0xA0b86a33E6441e8C8F60Ec4E9e29464b40507Dac",
      "name": "Compound",
      "symbol": "COMP",
      "decimals": 18,
      "logoURI": "https://example.com/comp.png"
    }
  ]
}
```

**Usage**:
```go
parser := &parsers.StandardTokenListParser{}

supportedChains := []uint64{1, 10, 56} // Ethereum, Optimism, BSC

tokenList, err := parser.Parse(
    jsonData,        // Raw JSON bytes
    supportedChains, // Supported chain IDs
)
```

### 2. Status Token List Parser (`StatusTokenListParser`)

Parses token lists in Status format, which extends the standard format with cross-chain token support.

**Key Features**:
- **Cross-chain tokens**: Single token entry with multiple chain deployments
- **Contracts mapping**: Maps chain IDs to contract addresses

**Format**:
```json
{
  "name": "Status Token List",
  "timestamp": "2023-01-01T00:00:00.000Z",
  "version": {
    "major": 1,
    "minor": 0,
    "patch": 0
  },
  "tokens": [
    {
      "crossChainId": "SNT",
      "symbol": "SNT",
      "name": "Status Network Token",
      "decimals": 18,
      "logoURI": "https://example.com/snt.png",
      "contracts": {
        "1": "0x744d70FDBE2Ba4CF95131626614a1763DF805B9E",
        "10": "0x650AF55D5877F289837c30b94af91538a7504b76",
        "42161": "0x707f635951193ddafbb40971a0fcaab8a6415160"
      }
    }
  ]
}
```

**Usage**:
```go
statusParser := &parsers.StatusTokenListParser{}

tokenList, err := statusParser.Parse(
    statusJsonData,
    supportedChains,
)

// Status format creates multiple Token entries for cross-chain tokens
for _, token := range tokenList.Tokens {
    fmt.Printf("Token: %s on chain %d (cross-chain ID: %s)\n",
               token.Symbol, token.ChainID, token.CrossChainID)
}
```

### 3. CoinGecko All Tokens Parser (`CoinGeckoAllTokensParser`)

Parses tokens from CoinGecko's comprehensive token database format.

**Key Features**:
- **Extensive token database**: Access to CoinGecko's large token collection
- **Platform mapping**: Configurable mapping from platform names to chain IDs
- **Default mappings**: Built-in support for major chains

**Format**:
```json
[
  {
    "id": "bitcoin",
    "symbol": "btc",
    "name": "Bitcoin",
    "platforms": {
      "ethereum": "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
      "binance-smart-chain": "0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c",
      "polygon-pos": "0x1BFD67037B42Cf73acF2047067bd4F2C47D9BfD6"
    }
  }
]
```

**Important Note**: CoinGecko format doesn't include token decimals, so all tokens will have `decimals: 0`. Consider using multicall3 to fetch decimals from contracts.

## List of Token Lists Parsers

### Status List of Token Lists Parser (`StatusListOfTokenListsParser`)

Parses metadata about collections of token lists, useful for token list discovery and management.

**Format**:
```json
{
  "timestamp": "2025-01-01T00:00:00.000Z",
  "version": {
    "major": 1,
    "minor": 0,
    "patch": 0
  },
  "tokenLists": [
    {
      "id": "uniswap",
      "sourceUrl": "https://tokens.uniswap.org",
      "schema": "https://uniswap.org/tokenlist.schema.json"
    },
    {
      "id": "compound",
      "sourceUrl": "https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json"
    }
  ]
}
```

**Usage**:
```go
parser := &parsers.StatusListOfTokenListsParser{}

listOfLists, err := parser.Parse(rawJsonData)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d token lists updated at %s\n",
           len(listOfLists.TokenLists), listOfLists.Timestamp)

for _, tokenListInfo := range listOfLists.TokenLists {
    fmt.Printf("- %s: %s\n", tokenListInfo.ID, tokenListInfo.SourceURL)
}
```


## Token Filtering

All token list parsers automatically filter tokens based on:

1. **Address Validation**: Only tokens with valid Ethereum addresses are included
2. **Chain Support**: Only tokens on supported chains are included
3. **Format Validation**: Malformed token entries are skipped with logging

```go
supportedChains := []uint64{1, 10, 56} // Ethereum, Optimism, BSC

// Parser will only include tokens from these chains
tokenList, err := parser.Parse(data, supportedChains)

// Tokens on unsupported chains (e.g., 42161 for Arbitrum) will be filtered out
```

## Output Format

### TokenList Structure

All token list parsers convert input data to the unified `types.TokenList` structure:

```go
type TokenList struct {
    Name             string                 `json:"name"`             // List name
    Timestamp        string                 `json:"timestamp"`        // Original list timestamp
    FetchedTimestamp string                 `json:"fetchedTimestamp"` // When fetched
    Source           string                 `json:"source"`           // Source URL
    Version          Version                `json:"version"`          // Semantic version
    Tags             map[string]interface{} `json:"tags"`             // Token tags
    LogoURI          string                 `json:"logoUri"`          // List logo
    Keywords         []string               `json:"keywords"`         // Keywords
    Tokens           []*Token               `json:"tokens"`           // Token array
}

type Token struct {
    CrossChainID string             `json:"crossChainId"` // Cross-chain identifier (Status format)
    ChainID      uint64             `json:"chainId"`      // Blockchain network ID
    Address      gethcommon.Address `json:"address"`      // Token contract address
    Decimals     uint               `json:"decimals"`     // Token decimals
    Name         string             `json:"name"`         // Token name
    Symbol       string             `json:"symbol"`       // Token symbol
    LogoURI      string             `json:"logoUri"`      // Token logo URL
    CustomToken  bool               `json:"custom"`       // Whether it's a custom token
}
```

### ListOfTokenLists Structure

For metadata about token list collections:

```go
type ListOfTokenLists struct {
    Timestamp  string        `json:"timestamp"`   // When the metadata was created
    Version    Version       `json:"version"`     // Metadata version
    TokenLists []ListDetails `json:"tokenLists"`  // Token list references
}

type ListDetails struct {
    ID        string `json:"id"`        // Unique identifier
    SourceURL string `json:"sourceUrl"` // URL to fetch the token list
    Schema    string `json:"schema"`    // Optional JSON schema URL
}
```

## Key Concepts

### Timestamp Handling

- **`Timestamp`**: When the original list was created/updated by the provider
- **`FetchedTimestamp`**: When the list was fetched by your application

### Cross-Chain Tokens (Status Format)

Status format supports tokens that exist on multiple chains:
- Single token entry with `crossChainId`
- `contracts` field maps chain IDs to addresses
- Parser creates separate `Token` objects for each chain

### Address Validation

All parsers validate Ethereum addresses:
- Invalid addresses are skipped
- Empty addresses are only allowed for native tokens
- Case normalization is handled automatically

## Default Chain Mappings

The package provides default mappings for CoinGecko platform names:

```go
var DefaultCoinGeckoChainsMapper = map[string]common.ChainID{
    "ethereum":            common.EthereumMainnet,    // 1
    "optimistic-ethereum": common.OptimismMainnet,    // 10
    "arbitrum-one":        common.ArbitrumMainnet,    // 42161
    "binance-smart-chain": common.BSCMainnet,         // 56
    "base":                common.BaseMainnet,        // 8453
}
```

## Testing

The package includes comprehensive tests for all parsers:

```bash
# Run all parser tests
go test ./pkg/tokens/parsers/...

# Run with verbose output
go test -v ./pkg/tokens/parsers/...

# Run specific parser tests
go test -run TestStandardTokenListParser -v ./pkg/tokens/parsers/...
go test -run TestStatusTokenListParser -v ./pkg/tokens/parsers/...
go test -run TestCoinGeckoAllTokensParser -v ./pkg/tokens/parsers/...
go test -run TestStatusListOfTokenListsParser -v ./pkg/tokens/parsers/...
```