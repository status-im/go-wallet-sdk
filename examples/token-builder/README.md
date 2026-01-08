# Token Builder Example

This example demonstrates how to use the `pkg/tokens/builder` package to incrementally build token collections from multiple sources with automatic deduplication and native token support.

## What it demonstrates

- **Incremental building**: Start empty and build up token collections step by step
- **Native token integration**: Automatically generate native tokens (ETH, BNB, etc.)
- **Automatic deduplication**: Prevent duplicate tokens using chain ID and address combinations
- **Raw data processing**: Parse and add token lists from various JSON formats
- **State management**: Track tokens and token lists throughout the building process
- **Flexible API**: Add tokens from parsed lists or raw JSON data
- **Performance**: Efficient token lookup and storage

## Run

```bash
cd examples/token-builder
go run main.go
```

## Example Output

```
🏗️  Token Builder Example
==========================

🚀 Basic Builder Usage
=======================
🏗️  Created builder for 4 chains
📊 Initial state: 0 tokens, 0 lists
🌐 Added native tokens: 4 tokens, 1 lists
➕ Added custom token list: 6 tokens, 2 lists

📋 Final Token Collection:
   📊 Summary: 6 unique tokens from 2 lists
   ⛓️  Tokens per chain:
      • Chain 1 (Ethereum): 2 tokens
      • Chain 56 (BSC): 2 tokens
      • Chain 10 (Optimism): 1 tokens
      • Chain 137 (Polygon): 1 tokens
   📋 Token lists:
      • native: Native tokens (4 tokens)
      • custom-tokens: Sample Token List (2 tokens)

📈 Incremental Building Pattern
===============================
🏗️  Building token collection incrementally...

1️⃣ Adding native tokens...
   ✅ Native tokens added: 4 total tokens

2️⃣ Adding DeFi token list...
   ✅ DeFi tokens added: 6 total tokens

3️⃣ Adding stablecoin list...
   ✅ Stablecoins added: 8 total tokens

4️⃣ Adding exchange token list...
   ✅ Exchange tokens added: 10 total tokens

📊 Building Progress Summary:
   📋 native: 4 tokens
   📋 defi-tokens: 2 tokens
   📋 stablecoins: 2 tokens
   📋 exchange-tokens: 2 tokens

🎯 Final collection: 10 unique tokens across 4 lists

📄 Raw Token List Processing
============================
📄 Processing standard format token list...
   ✅ Standard list processed: 6 total tokens

📄 Processing Status format token list...
   ✅ Status list processed: 10 total tokens

📋 Raw Processing Results:
   📊 Summary: 10 unique tokens from 3 lists
   ⛓️  Tokens per chain:
      • Chain 1 (Ethereum): 3 tokens
      • Chain 56 (BSC): 3 tokens
      • Chain 10 (Optimism): 2 tokens
      • Chain 137 (Polygon): 2 tokens
   📋 Token lists:
      • native: Native tokens (4 tokens)
      • uniswap-example: Uniswap Example List (2 tokens)
      • status-example: Status Example List (4 tokens)

🔄 Token Deduplication
======================
🌐 Initial tokens (native): 4

📄 Adding overlapping token lists...
   ➕ Added list 1: 6 tokens (+2)
   ➕ Added list 2: 7 tokens (+1) - USDC deduplicated!
   ➕ Added list 3: 8 tokens (+1) - Different chain USDC kept

🔍 Deduplication Analysis:
   💰 USDC on chain 1: 0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB
   💰 USDC on chain 56: 0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d
   📊 Total USDC tokens: 2 (different chains = different tokens)

✅ Deduplication complete: 8 unique tokens from 6 lists

🎯 Advanced Builder Patterns
============================
🎯 Advanced Builder Pattern Examples:

1️⃣ Builder with validation:
   ✅ Validation passed: 1 native tokens added

2️⃣ Conditional building based on chain support:
   ✅ Ethereum-only builder: 1 tokens

3️⃣ Builder state inspection:
   📊 Builder state:
      • Total tokens: 4 tokens
      • Total lists: 1
      • Memory efficiency: ~200 bytes per token

4️⃣ Error handling strategies:
   🛠️  Error handling examples:
      📝 Testing empty raw data...
      ✅ Correctly caught error: raw token list data is empty
      📝 Testing nil parser...
      ✅ Correctly caught error: parser is nil
      📝 Testing invalid JSON...
      ✅ Correctly caught error: invalid character 'i' looking for beginning of value
      🎯 Error handling validation complete!

✅ Token Builder examples completed!
```

## Code Examples

### Basic Builder Pattern

```go
import "github.com/status-im/go-wallet-sdk/pkg/tokens/builder"

// Create builder for specific chains
supportedChains := []uint64{1, 56, 10, 137} // Ethereum, BSC, Optimism, Polygon
tokenBuilder := builder.New(supportedChains, nil) // nil = no tokens to skip

// Start empty - Builder pattern
fmt.Printf("Initial: %d tokens\n", len(tokenBuilder.GetTokens())) // 0

// Add native tokens for all supported chains
err := tokenBuilder.AddNativeTokenList()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("With native: %d tokens\n", len(tokenBuilder.GetTokens())) // 4

// Add custom token list
customList := &types.TokenList{
    Name: "My Custom List",
    Tokens: []*types.Token{...},
}
tokenBuilder.AddTokenList("custom", customList)
fmt.Printf("Final: %d tokens\n", len(tokenBuilder.GetTokens()))
```

### Incremental Building

```go
tokenBuilder := builder.New(supportedChains, nil)

// Build step by step
tokenBuilder.AddNativeTokenList()
fmt.Printf("Step 1: %d tokens\n", len(tokenBuilder.GetTokens()))

tokenBuilder.AddTokenList("defi", defiTokenList)
fmt.Printf("Step 2: %d tokens\n", len(tokenBuilder.GetTokens()))

tokenBuilder.AddTokenList("stablecoins", stablecoinList)
fmt.Printf("Step 3: %d tokens\n", len(tokenBuilder.GetTokens()))

// Each step builds on the previous one
tokens := tokenBuilder.GetTokens()
lists := tokenBuilder.GetTokenLists()
```

### Raw Token List Processing

```go
import "github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"

tokenBuilder := builder.New(supportedChains, nil)

// Process raw JSON with appropriate parser
rawJSON := []byte(`{
    "name": "Uniswap Default List",
    "timestamp": "2025-01-01T00:00:00Z",
    "tokens": [...]
}`)

parser := &parsers.StandardTokenListParser{}
err := tokenBuilder.AddRawTokenList(
    "uniswap-default",
    rawJSON,
    "https://tokens.uniswap.org",
    time.Now(),
    parser,
)

if err != nil {
    log.Printf("Failed to process raw list: %v", err)
}
```

### Automatic Deduplication

```go
tokenBuilder := builder.New([]uint64{1}, nil) // Ethereum only

// Create overlapping lists
list1 := &types.TokenList{
    Name: "List 1",
    Tokens: []*types.Token{
        {
            ChainID: 1,
            Address: common.HexToAddress("0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB"),
            Symbol:  "USDC",
            // ... other fields
        },
    },
}

list2 := &types.TokenList{
    Name: "List 2",
    Tokens: []*types.Token{
        {
            ChainID: 1,
            Address: common.HexToAddress("0xA0b86a33E6441b6d9e4AEda6D7bb57B75FE3f5dB"), // Same!
            Symbol:  "USDC",
            // ... other fields
        },
    },
}

// Add both lists
tokenBuilder.AddTokenList("list1", list1)
fmt.Printf("After list1: %d tokens\n", len(tokenBuilder.GetTokens())) // 1

tokenBuilder.AddTokenList("list2", list2)
fmt.Printf("After list2: %d tokens\n", len(tokenBuilder.GetTokens())) // Still 1 - deduplicated!

// Different chains are NOT deduplicated
list3 := &types.TokenList{
    Name: "BSC List",
    Tokens: []*types.Token{
        {
            ChainID: 56, // Different chain
            Address: common.HexToAddress("0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d"),
            Symbol:  "USDC", // Same symbol, different chain
        },
    },
}

tokenBuilder.AddTokenList("bsc", list3)
fmt.Printf("After BSC: %d tokens\n", len(tokenBuilder.GetTokens())) // 2 - different chains
```

## Key Concepts

### Builder Pattern

The builder follows the classic Builder pattern:
- **Start empty**: `New()` creates empty builder
- **Build incrementally**: Add components one at a time
- **Maintain state**: Internal state tracks tokens and lists
- **Get results**: Retrieve final collections when ready

### Token Deduplication

Tokens are deduplicated using a unique key:
```go
key := fmt.Sprintf("%d-%s", token.ChainID, token.Address.Hex())
```

**Same token** (deduplicated):
- Chain ID: 1, Address: 0x123... (first occurrence kept)
- Chain ID: 1, Address: 0x123... (duplicate ignored)

**Different tokens** (both kept):
- Chain ID: 1, Address: 0x123... (Ethereum USDC)
- Chain ID: 56, Address: 0x456... (BSC USDC)

### Token Filtering

The builder supports filtering out specific tokens by their keys. This is useful for excluding invalid or unwanted tokens (e.g., tokens with no value or deprecated tokens).

**Token Key Format:**
Token keys follow the format: `"{chainID}-{lowercaseAddress}"`

```go
// Skip specific tokens (e.g., Optimism ETH with no value)
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

tokens := builder.GetTokens()
// tokens will only contain the second token (0x4200...), the skipped token (0xdead...) is excluded
```

**Note:** Token lists are still stored in the builder even if all their tokens are filtered out. Only the tokens themselves are excluded from the unified token collection.

### Native Token Support

Native tokens are automatically generated for supported chains:
- **Ethereum (1)**: ETH native token
- **BSC (56)**: BNB native token
- **Other chains**: ETH-equivalent native tokens

```go
// Generates native tokens for all supported chains
err := builder.AddNativeTokenList()
```

## Performance Characteristics

### Time Complexity
- **Add token**: O(1) - Hash map insertion
- **Deduplication**: O(1) - Hash map lookup
- **Get tokens**: O(1) - Return map reference
- **Build operations**: O(n) where n = total tokens

### Memory Usage
- **Token storage**: ~200 bytes per unique token
- **Deduplication map**: Key string + pointer per token
- **Lists storage**: Reference to original TokenList objects

### Scalability
- **Tokens**: Handles 100,000+ tokens efficiently
- **Lists**: No practical limit on number of lists
- **Memory**: Linear scaling with number of unique tokens

## Advanced Usage Patterns

### Conditional Building

```go
// Build different collections based on conditions
func buildTokenCollection(includeTestnets bool) *builder.Builder {
    var chains []uint64

    // Always include mainnets
    chains = append(chains, 1, 56, 137) // Ethereum, BSC, Polygon

    // Conditionally include testnets
    if includeTestnets {
        chains = append(chains, 11155111, 97) // Sepolia, BSC Testnet
    }

    builder := builder.New(chains, nil)
    builder.AddNativeTokenList()

    return builder
}
```

### Error-Tolerant Building

```go
func buildWithErrorTolerance(rawLists map[string][]byte) (*builder.Builder, []error) {
    builder := builder.New(supportedChains, nil)
    builder.AddNativeTokenList()

    var errors []error
    parser := &parsers.StandardTokenListParser{}

    for listID, rawData := range rawLists {
        err := builder.AddRawTokenList(listID, rawData, "", time.Now(), parser)
        if err != nil {
            errors = append(errors, fmt.Errorf("failed to add %s: %w", listID, err))
            continue // Continue with other lists
        }
    }

    return builder, errors
}
```

### Builder Factory Pattern

```go
type BuilderFactory struct {
    defaultChains []uint64
    parsers       map[string]parsers.TokenListParser
}

func NewBuilderFactory(chains []uint64) *BuilderFactory {
    return &BuilderFactory{
        defaultChains: chains,
        parsers: map[string]parsers.TokenListParser{
            "standard": &parsers.StandardTokenListParser{},
            "status":   &parsers.StatusTokenListParser{},
            "coingecko": &parsers.CoinGeckoAllTokensParser{},
        },
    }
}

func (f *BuilderFactory) CreateBuilder(profile string) *builder.Builder {
    switch profile {
    case "defi":
        return f.createDefiBuilder()
    case "trading":
        return f.createTradingBuilder()
    default:
        return builder.New(f.defaultChains, nil)
    }
}

func (f *BuilderFactory) createDefiBuilder() *builder.Builder {
    builder := builder.New(f.defaultChains, nil)
    builder.AddNativeTokenList()
    // Add DeFi-specific token lists
    return builder
}
```

## Error Handling

### Common Errors

```go
// Empty raw data
err := builder.AddRawTokenList("test", []byte{}, "url", time.Now(), parser)
// Returns: ErrEmptyRawTokenList

// Nil parser
err = builder.AddRawTokenList("test", data, "url", time.Now(), nil)
// Returns: ErrParserIsNil

// Parser error (invalid JSON, missing fields, etc.)
err = builder.AddRawTokenList("test", invalidData, "url", time.Now(), parser)
// Returns: parser-specific error
```

### Error Handling Strategy

```go
func safeAddRawTokenList(builder *builder.Builder, listID string, data []byte, parser parsers.TokenListParser) error {
    if len(data) == 0 {
        return fmt.Errorf("empty data for list %s", listID)
    }

    if parser == nil {
        return fmt.Errorf("nil parser for list %s", listID)
    }

    err := builder.AddRawTokenList(listID, data, "", time.Now(), parser)
    if err != nil {
        return fmt.Errorf("failed to add list %s: %w", listID, err)
    }

    return nil
}
```

## Integration Examples

### With Token Manager

```go
// Build token collection then create manager
builder := builder.New(supportedChains, nil)
builder.AddNativeTokenList()
builder.AddTokenList("uniswap", uniswapList)

// Manager would use builder internally
config := &manager.Config{
    MainListID: "uniswap",
    InitialLists: map[string][]byte{
        "uniswap": uniswapData,
    },
    Parsers: map[string]parsers.TokenListParser{
        "uniswap": &parsers.StandardTokenListParser{},
    },
    Chains: supportedChains,
}
```

## Best Practices

### 1. **Start with Native Tokens**
```go
// Always add native tokens first for completeness
builder := builder.New(chains, nil)
builder.AddNativeTokenList() // ETH, BNB, etc.
```

### 2. **Handle Errors Gracefully**
```go
// Don't fail entire build if one list fails
for listID, data := range tokenLists {
    if err := builder.AddRawTokenList(listID, data, "", time.Now(), parser); err != nil {
        log.Printf("Warning: Failed to add %s: %v", listID, err)
        continue // Keep going with other lists
    }
}
```

### 3. **Use Appropriate Chain Filtering**
```go
// Filter chains based on your use case
mainnetChains := []uint64{1, 56, 137} // Production
testnetChains := []uint64{11155111, 97} // Testing
allChains := append(mainnetChains, testnetChains...) // Development
```

### 4. **Monitor Builder State**
```go
// Track building progress
builder := builder.New(chains, nil)
fmt.Printf("Initial: %d tokens\n", len(builder.GetTokens()))

builder.AddNativeTokenList()
fmt.Printf("With native: %d tokens\n", len(builder.GetTokens()))

for _, list := range tokenLists {
    builder.AddTokenList(list.ID, list)
    fmt.Printf("Added %s: %d total tokens\n", list.ID, len(builder.GetTokens()))
}
```

## Dependencies

- `github.com/status-im/go-wallet-sdk/pkg/tokens/parsers` - Token list parsing
- `github.com/status-im/go-wallet-sdk/pkg/tokens/types` - Core token types
- `github.com/ethereum/go-ethereum/common` - Ethereum address types
- `time` - Timestamp handling
- `fmt` - Error formatting

The token builder provides a flexible, efficient foundation for building token collections in blockchain applications with automatic deduplication and comprehensive format support.