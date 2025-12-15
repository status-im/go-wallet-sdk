# Token Fetcher Package

The `fetcher` package provides functionality for fetching token lists and related data from remote sources with support for HTTP caching, JSON schema validation, and concurrent operations.

## Use it when

- You need to fetch token lists (or a list-of-token-lists) over HTTP.
- You want ETag-based caching to reduce bandwidth.
- You want optional JSON schema validation and concurrent fetching.

## Key entrypoints

- `fetcher.New(config)` / `fetcher.DefaultConfig()`
- `(*fetcher).Fetch(ctx, details)`
- `fetcher.FetchConcurrent(ctx, details)`
- Schemas: `fetcher.TokenListSchema`, `fetcher.ListOfTokenListsSchema`

## Overview

The fetcher package is designed to:
- Fetch individual token lists and token list metadata
- Support HTTP caching with ETags to minimize network traffic
- Validate JSON data against schemas
- Handle concurrent fetching operations safely
- Provide robust error handling with context support

## Core Components

### Configuration

#### Config

The `Config` struct allows customization of HTTP client behavior:

```go
type Config struct {
    Timeout            time.Duration // Request timeout (default: 5s)
    IdleConnTimeout    time.Duration // Connection idle timeout (default: 90s)
    MaxIdleConns       int           // Max idle connections (default: 10)
    DisableCompression bool          // Disable gzip compression (default: false)
}
```

#### DefaultConfig()

Returns the default configuration:

```go
config := fetcher.DefaultConfig()
// Config{
//     Timeout:            5 * time.Second,
//     IdleConnTimeout:    90 * time.Second,
//     MaxIdleConns:       10,
//     DisableCompression: false,
// }
```

### Fetcher Interface

The main interface provides methods for fetching resources:

```go
type Fetcher interface {
    // Fetch fetches a single resource from the URL specified in the details.
    Fetch(ctx context.Context, details FetchDetails) (FetchedData, error)

    // FetchConcurrent fetches multiple resources concurrently from the URLs specified in the details.
    FetchConcurrent(ctx context.Context, details []FetchDetails) ([]FetchedData, error)
}
```

### Data Types

#### FetchDetails

Represents the details needed to fetch a resource:

```go
type FetchDetails struct {
    types.ListDetails  // Embedded: ID, SourceURL, Schema
    Etag string             // HTTP ETag for caching
}
```

Where `types.ListDetails` contains:
```go
type ListDetails struct {
    ID        string `json:"id" validate:"required"`            // Unique identifier
    SourceURL string `json:"sourceUrl" validate:"required,url"` // URL to fetch from
    Schema    string `json:"schema"`                            // Optional JSON schema URL
}
```

#### FetchedData

Represents the result of a fetch operation:

```go
type FetchedData struct {
    FetchDetails           // Original fetch details
    Fetched  time.Time     // Timestamp when the resource was fetched
    JsonData []byte        // Raw JSON data (nil if 304 Not Modified)
    Error    error         // Error that occurred during fetch (if any)
}
```

## API Reference

### Constructor

#### `New(config Config) *fetcher`

Creates a new fetcher instance with the specified configuration.

**Parameters:**
- `config`: Configuration for the HTTP client

**Example with default config:**
```go
f := fetcher.New(fetcher.DefaultConfig())
```

**Example with custom config:**
```go
config := fetcher.Config{
    Timeout:            10 * time.Second,
    IdleConnTimeout:    120 * time.Second,
    MaxIdleConns:       20,
    DisableCompression: false,
}
f := fetcher.New(config)
```

### Core Methods

#### `Fetch(ctx context.Context, details FetchDetails) (FetchedData, error)`

Fetches a single resource.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `details`: Fetch details including URL, schema, and ETag

**Returns:**
- `FetchedData`: Result containing data, metadata, and potential errors
- `error`: Only returned for fundamental errors (validation failures, etc.)

**Example:**
```go
details := fetcher.FetchDetails{
    ListDetails: types.ListDetails{
        ID:        "uniswap",
        SourceURL: "https://tokens.uniswap.org",
        Schema:    "https://uniswap.org/tokenlist.schema.json",
    },
    Etag: "previous-etag", // Use empty string for first fetch
}

result, err := f.Fetch(ctx, details)
if err != nil {
    log.Fatal(err) // Validation or fundamental error
}

if result.Error != nil {
    log.Printf("Failed to fetch: %v", result.Error)
} else if result.JsonData == nil {
    log.Println("No new data (304 Not Modified)")
} else {
    log.Printf("Fetched %d bytes", len(result.JsonData))
}
```

#### `FetchConcurrent(ctx context.Context, details []FetchDetails) ([]FetchedData, error)`

Fetches multiple resources concurrently using goroutines.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `details`: Slice of fetch details for concurrent fetching

**Returns:**
- `[]FetchedData`: Results for all fetch operations (successful and failed)
- `error`: Only returned for fundamental errors

**Example:**
```go
details := []fetcher.FetchDetails{
    {
        ListDetails: types.ListDetails{
            ID:        "uniswap",
            SourceURL: "https://tokens.uniswap.org",
        },
    },
    {
        ListDetails: types.ListDetails{
            ID:        "compound",
            SourceURL: "https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json",
        },
    },
}

results, err := f.FetchConcurrent(ctx, details)
if err != nil {
    log.Fatal(err)
}

// Process individual results
for _, result := range results {
    if result.Error != nil {
        log.Printf("Failed to fetch %s: %v", result.ID, result.Error)
    } else {
        log.Printf("Successfully fetched %s (%d bytes)", result.ID, len(result.JsonData))
    }
}
```

## Features

### 1. HTTP Caching with ETags

The fetcher leverages HTTP ETags for efficient caching:

- **304 Not Modified**: Returns empty `JsonData` when content hasn't changed
- **Fresh Data**: Downloads new data when ETag differs
- **Automatic Management**: ETags are handled automatically

```go
// First fetch
result1, _ := f.Fetch(ctx, details)
fmt.Printf("ETag: %s\n", result1.Etag)

// Subsequent fetch with ETag
details.Etag = result1.Etag
result2, _ := f.Fetch(ctx, details)
if result2.JsonData == nil {
    fmt.Println("No changes (304 Not Modified)")
}
```

### 2. JSON Schema Validation

Automatic validation against JSON schemas when specified:

```go
details := fetcher.FetchDetails{
    ListDetails: types.ListDetails{
        ID:        "validated-list",
        SourceURL: "https://tokens.uniswap.org",
        Schema:    "https://uniswap.org/tokenlist.schema.json", // Schema URL
    },
}

result, _ := f.Fetch(ctx, details)
if result.Error != nil {
    // Could be network error or schema validation error
    if errors.Is(result.Error, fetcher.ErrTokenListDoesNotMatchSchema) {
        log.Println("Schema validation failed")
    }
}
```

### 3. Context Support

Full support for context cancellation and timeouts:

```go
// Timeout after 30 seconds
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Cancellation will interrupt ongoing HTTP requests
results, err := f.FetchConcurrent(ctx, details)
```

### 4. Concurrent Safety

The fetcher is designed for safe concurrent operations:

- **Goroutine-based**: `FetchConcurrent` uses goroutines with proper synchronization
- **Channel Safety**: Safe handling of channels with context cancellation
- **Individual Errors**: Each fetch operation has independent error handling

## Validation

### URL Validation

FetchDetails are validated using struct tags:

```go
type ListDetails struct {
    ID        string `json:"id" validate:"required"`
    SourceURL string `json:"sourceUrl" validate:"required,url"`
    Schema    string `json:"schema"` // Optional
}
```

Invalid details will cause `Fetch` to return an error immediately.

### Schema Validation

When a schema URL is provided:
1. The schema is fetched from the URL or used as inline JSON
2. The fetched JSON data is validated against the schema
3. Validation failures are returned as `ErrTokenListDoesNotMatchSchema`

## Built-in Schema

The package includes a built-in schema for list-of-token-lists format:

```go
const ListOfTokenListsSchema = `{...}` // JSON Schema for token list metadata
```

## Testing

The package includes comprehensive tests with a test HTTP server:

```bash
# Run all tests
go test ./pkg/tokens/fetcher/...

# Run with verbose output
go test -v ./pkg/tokens/fetcher/...

# Run specific tests
go test -run TestFetch -v ./pkg/tokens/fetcher/...
go test -run TestFetchConcurrent -v ./pkg/tokens/fetcher/...
```

## Thread Safety

**The fetcher implementation is thread-safe** and can be used concurrently from multiple goroutines. The underlying HTTP client and all operations are designed for concurrent access.