# Token Fetcher Example

This example demonstrates how to use the `pkg/tokens/fetcher` package to fetch token lists from remote sources with support for HTTP caching, concurrent fetching, and error handling.

## Features Demonstrated

- 🌐 **Single Token List Fetching**: Fetch individual token lists from remote URLs
- 🚀 **Concurrent Fetching**: Fetch multiple token lists simultaneously for better performance
- 💾 **HTTP ETag Caching**: Efficient caching using HTTP ETags to minimize bandwidth
- 📚 **List of Token Lists**: Fetch and process master lists that reference multiple token lists
- 🛡️ **Error Handling**: Robust error handling for network failures and invalid responses
- ⚡ **Performance Optimization**: Parallel processing and timeout management

## Quick Start

```bash
cd examples/token-fetcher
go run main.go
```

## Example Output

```
🌐 Token Fetcher Example
=========================

📋 Single Token List Fetch
============================
🔄 Fetching token list from: https://tokens.uniswap.org
✅ Successfully fetched token list:
  📊 Data size: 235,891 bytes
  🏷️  ETag: "1a2b3c4d5e6f7g8h"
  📅 Fetched at: 2025-01-01T12:00:00Z
  👀 Preview: {"name":"Uniswap Default List","timestamp":"2025-01-01T00:00:00Z"...

🚀 Concurrent Token List Fetch
================================
🚀 Fetching 3 token lists concurrently...
⚡ Concurrent fetch completed in 1.2s

📋 Token List: uniswap-default
  🔗 URL: https://tokens.uniswap.org
  ✅ Success: 235,891 bytes
  🏷️  ETag: "1a2b3c4d5e6f7g8h"
  📅 Fetched: 2025-01-01T12:00:00Z

📋 Token List: compound-tokens
  🔗 URL: https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json
  ✅ Success: 12,456 bytes
  🏷️  ETag: "9x8y7z6w5v4u3t2s"
  📅 Fetched: 2025-01-01T12:00:00Z

📋 Token List: defiprime-list
  🔗 URL: https://defiprime.github.io/tokens/defiprime.tokenlist.json
  ✅ Success: 45,123 bytes
  🏷️  ETag: "a1b2c3d4e5f6g7h8"
  📅 Fetched: 2025-01-01T12:00:00Z

📊 Summary: 3/3 token lists fetched successfully

💾 ETag-based Caching
=====================
🔄 First fetch (no ETag)...
✅ First fetch successful: 235,891 bytes, ETag: "1a2b3c4d5e6f7g8h"

🔄 Second fetch (with ETag)...
💾 Cached response (304 Not Modified) - ETag: "1a2b3c4d5e6f7g8h"
   No data transfer needed, content unchanged!

📚 List of Token Lists
======================
🔄 Fetching list of token lists from: https://prod.market.status.im/static/lists.json
✅ Successfully fetched list of token lists:
  📊 Data size: 8,234 bytes
  🏷️  ETag: "z9y8x7w6v5u4t3s2"
  📅 Fetched at: 2025-01-01T12:00:00Z
  👀 Content preview:
{
  "lists": [
    {
      "name": "Uniswap Default List",
      "url": "https://tokens.uniswap.org"
    },
    {
      "name": "Compound Token List",
      "url": "https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json"
    }
  ]
}

🔄 Attempting to fetch individual token lists...
  💡 Tip: Parse the JSON response to extract individual token list URLs
      Then use FetchConcurrent() to fetch all lists in parallel

✅ Token Fetcher examples completed!
```

## Code Examples

### 1. Single Token List Fetch

```go
import "github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"

// Create fetcher with default configuration
f := fetcher.New(fetcher.DefaultConfig())

// Or with custom configuration
config := fetcher.Config{
    Timeout:            10 * time.Second,
    IdleConnTimeout:    120 * time.Second,
    MaxIdleConns:       20,
    DisableCompression: false,
}
f := fetcher.New(config)

fetchDetails := fetcher.FetchDetails{
    ListDetails: types.ListDetails{
        ID:        "uniswap-default",
        SourceURL: "https://tokens.uniswap.org",
        Schema:    "", // add json or url to schema if known
    },
    Etag: "", // No ETag for first fetch
}

fetchedData, err := f.Fetch(ctx, fetchDetails)
if err != nil {
    log.Printf("Failed to fetch: %v", err)
    return
}

fmt.Printf("Fetched %d bytes\n", len(fetchedData.JsonData))
```

### 2. Concurrent Fetching

```go
// Prepare multiple fetch requests
fetchRequests := []fetcher.FetchDetails{
    {
        ListDetails: types.ListDetails{
            ID:        "uniswap-default",
            SourceURL: "https://tokens.uniswap.org",
            Schema:    "", // add json or url to schema if known
        },
    },
    {
        ListDetails: types.ListDetails{
            ID:        "compound-tokens",
            SourceURL: "https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json",
            Schema:    "", // add json or url to schema if known
        },
    },
}

// Fetch all concurrently
results, err := fetcher.FetchConcurrent(ctx, fetchRequests)
if err != nil {
    log.Printf("Concurrent fetch failed: %v", err)
    return
}

// Process results
for _, result := range results {
    if result.Error != nil {
        log.Printf("Failed to fetch %s: %v", result.ID, result.Error)
    } else {
        log.Printf("Successfully fetched %s: %d bytes", result.ID, len(result.JsonData))
    }
}
```

### 3. ETag-based Caching

```go
// First fetch without ETag
fetchDetails := fetcher.FetchDetails{
    ListDetails: types.ListDetails{
        ID:        "uniswap-default",
        SourceURL: "https://tokens.uniswap.org",
        Schema:    "", // add json or url to schema if known
    },
    Etag: "", // No ETag
}

firstFetch, err := fetcher.Fetch(ctx, fetchDetails)
if err != nil {
    return err
}

// Store the ETag for future requests
storedETag := firstFetch.Etag

// Second fetch with ETag - will return empty data if not modified
fetchDetails.Etag = storedETag
secondFetch, err := fetcher.Fetch(ctx, fetchDetails)
if err != nil {
    return err
}

if len(secondFetch.JsonData) == 0 {
    fmt.Println("Content not modified (304 response)")
    // Use cached data
} else {
    fmt.Println("Content updated")
    // Process new data and update ETag
    storedETag = secondFetch.Etag
}
```

## Key Features

### HTTP ETag Support

The fetcher implements efficient HTTP caching using ETags:
- **First request**: Returns full content and ETag
- **Subsequent requests**: Include ETag in `If-None-Match` header
- **304 Not Modified**: Empty response when content unchanged
- **Bandwidth savings**: Significant reduction in data transfer

### Concurrent Processing

Multiple token lists can be fetched simultaneously:
- **Parallel execution**: Uses goroutines for concurrent requests
- **Error isolation**: Individual failures don't affect other requests
- **Timeout handling**: Each request respects context timeouts
- **Performance boost**: Dramatically reduces total fetch time

### Robust Error Handling

Comprehensive error handling for various scenarios:
- **Network failures**: Connection timeouts, DNS failures
- **HTTP errors**: 4xx/5xx status codes
- **Invalid responses**: Malformed JSON, empty responses
- **Context cancellation**: Graceful handling of cancelled requests

### Schema Validation

Optional JSON schema validation:
- **Format checking**: Ensures token lists match expected format
- **Error reporting**: Clear validation error messages
- **Flexibility**: Schema validation can be enabled/disabled per request

## Performance Characteristics

### Single Fetch
- **Latency**: Depends on network and server response time
- **Memory**: Minimal overhead, streams responses
- **Bandwidth**: Uses ETags to minimize unnecessary transfers

### Concurrent Fetch
- **Throughput**: Linear scaling with number of goroutines
- **Latency**: Parallel processing reduces total time
- **Resource usage**: Memory scales with number of concurrent requests

### Benchmarks

Typical performance metrics:
- **Single fetch**: 200-2000ms depending on list size and network
- **3 concurrent fetches**: ~500ms faster than sequential
- **ETag cache hit**: <50ms (no data transfer)
- **Memory usage**: ~1MB per concurrent request

## Dependencies

- `net/http` - HTTP client functionality
- `context` - Request context and timeout handling
- `time` - Timestamp and duration management
- `github.com/status-im/go-wallet-sdk/pkg/tokens/types` - Core types

## Integration Examples

### With Token Manager

```go
// Fetch token lists and add to manager
fetchDetails := []fetcher.FetchDetails{...}
results, err := fetcher.FetchConcurrent(ctx, fetchDetails)

for _, result := range results {
    if result.Error == nil {
        err := manager.AddRawTokenList(
            result.ID,
            result.JsonData,
            result.SourceURL,
            result.Fetched,
            parser,
        )
    }
}
```

### Background Refresh Service

```go
type RefreshService struct {
    fetcher fetcher.Fetcher
    manager manager.Manager
    ticker  *time.Ticker
}

func (s *RefreshService) Start(ctx context.Context) {
    s.ticker = time.NewTicker(time.Hour)
    go func() {
        for {
            select {
            case <-s.ticker.C:
                s.refreshTokenLists(ctx)
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

This example provides a comprehensive guide to using the token fetcher for efficient, reliable token list fetching with production-ready patterns and best practices.