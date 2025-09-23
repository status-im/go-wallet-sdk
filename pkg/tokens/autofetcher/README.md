# AutoFetcher Package

The `autofetcher` package provides automated background fetching and caching of token lists with configurable refresh intervals. It supports both direct token list fetching and remote list-of-token-lists discovery patterns.

## Overview

The autofetcher package is designed to:
- **Automatically fetch token lists** in the background with configurable intervals
- **Support two modes**: direct token lists and remote list-of-token-lists
- **Provide thread-safe operations** for concurrent usage
- **Handle HTTP caching** with ETags to minimize network traffic
- **Store fetched content** using a pluggable ContentStore interface
- **Graceful lifecycle management** with Start/Stop operations

## Key Features

- **Thread-Safe**: All operations are safe for concurrent access
- **Configurable Intervals**: Set custom refresh and check intervals
- **Error Reporting**: Real-time error notifications via channels
- **ETag Support**: Automatic HTTP caching to reduce bandwidth
- **Flexible Storage**: Pluggable ContentStore interface
- **Context Support**: Full context cancellation and timeout support

## Core Types

### AutoFetcher Interface

```go
type AutoFetcher interface {
    // Start starts the background autofetcher process.
    // Returns a channel that receives errors from refresh operations.
    // If no error is sent (nil), the refresh was successful.
    // Can be called multiple times safely - subsequent calls return the same channel.
    Start(ctx context.Context) (refreshCh chan error)

    // Stop stops the background autofetcher process.
    // Blocks until the background goroutine has finished.
    // Can be called multiple times safely.
    Stop()
}
```

### ContentStore Interface

```go
type ContentStore interface {
    // GetEtag retrieves the Etag for a given ID.
    GetEtag(id string) (string, error)

    // Get retrieves the content for a given ID.
    Get(id string) (Content, error)

    // Set stores the content for a given ID.
    Set(id string, content Content) error

    // GetAll retrieves all content.
    GetAll() (map[string]Content, error)
}
```

**Important**: ContentStore implementations MUST be thread-safe for concurrent access.

## Usage Patterns

### Pattern 1: Direct Token Lists

When you have a known set of token lists to fetch:

```go
import (
    "context"
    "log"
    "time"

    "github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
    "github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
    "github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

// Create configuration for token lists fetching
config := autofetcher.ConfigTokenLists{
    Config: autofetcher.Config{
        LastUpdate:               time.Now().Add(-time.Hour),
        AutoRefreshInterval:      30 * time.Minute,
        AutoRefreshCheckInterval: time.Minute,
    },
    TokenLists: []types.ListDetails{
        {
            ID:        "uniswap",
            SourceURL: "https://tokens.uniswap.org",
            Schema:    "https://uniswap.org/tokenlist.schema.json",
        },
        {
            ID:        "compound",
            SourceURL: "https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json",
        },
    },
}

// Create HTTP fetcher with default configuration
httpFetcher := fetcher.New(fetcher.DefaultConfig())

// Create your ContentStore implementation
myContentStore := &MyContentStore{}

// Create autofetcher
autoFetcher, err := autofetcher.NewAutofetcherFromTokenLists(config, httpFetcher, myContentStore)
if err != nil {
    log.Fatal(err)
}

// Start background fetching
ctx := context.Background()
refreshCh := autoFetcher.Start(ctx)

// Monitor for errors
go func() {
    for err := range refreshCh {
        if err != nil {
            log.Printf("Refresh error: %v", err)
        } else {
            log.Println("Refresh completed successfully")
        }
    }
}()

// Stop when done
defer autoFetcher.Stop()
```

### Pattern 2: Remote List of Token Lists

When you want to discover token lists from a remote registry:

```go
import (
    "context"
    "log"
    "time"

    "github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
    "github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
    "github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
    "github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

// Create configuration for remote list of token lists fetching
config := autofetcher.ConfigRemoteListOfTokenLists{
    Config: autofetcher.Config{
        LastUpdate:               time.Now().Add(-time.Hour),
        AutoRefreshInterval:      time.Hour,
        AutoRefreshCheckInterval: 5 * time.Minute,
    },
    RemoteListOfTokenListsFetchDetails: types.ListDetails{
        ID:        "status-lists",
		SourceURL: "https://prod.market.status.im/static/lists.json",
		Schema:    fetcher.ListOfTokenListsSchema,
    },
    RemoteListOfTokenListsParser: &parsers.StatusListOfTokenListsParser{},
}

// Create HTTP fetcher with default configuration
httpFetcher := fetcher.New(fetcher.DefaultConfig())

// Create your ContentStore implementation
myContentStore := &MyContentStore{}

// Create autofetcher
autoFetcher, err := autofetcher.NewAutofetcherFromRemoteListOfTokenLists(config, httpFetcher, myContentStore)
if err != nil {
    log.Fatal(err)
}

// Start background fetching
ctx := context.Background()
refreshCh := autoFetcher.Start(ctx)

// Monitor refresh operations
go func() {
    for err := range refreshCh {
        if err != nil {
            log.Printf("Auto-refresh failed: %v", err)
        } else {
            log.Println("Auto-refresh completed successfully")
        }
    }
}()

// Stop when done
defer autoFetcher.Stop()
```

## Configuration

### Config Fields

```go
type Config struct {
    LastUpdate               time.Time     // When data was last updated
    AutoRefreshInterval      time.Duration // How often to refresh
    AutoRefreshCheckInterval time.Duration // How often to check if refresh is needed
}
```

**Important**: `AutoRefreshCheckInterval` must be <= `AutoRefreshInterval`

### ConfigTokenLists

Used for direct token list fetching:

```go
type ConfigTokenLists struct {
    Config
    TokenLists []types.ListDetails
}
```

### ConfigRemoteListOfTokenLists

Used for remote list-of-token-lists pattern:

```go
type ConfigRemoteListOfTokenLists struct {
    Config
    RemoteListOfTokenListsFetchDetails types.ListDetails
    RemoteListOfTokenListsParser       parsers.ListOfTokenListsParser
}
```

### Refresh Logic

The autofetcher checks `time.Since(LastUpdate) >= AutoRefreshInterval` every `AutoRefreshCheckInterval` to determine if a refresh is needed.

**Example timing**:
- `AutoRefreshInterval: 30 * time.Minute` - Refresh every 30 minutes
- `AutoRefreshCheckInterval: time.Minute` - Check every minute if 30 minutes have passed
- Setting `LastUpdate: time.Now().Add(-time.Hour)` - Forces immediate refresh on first check

## Error Handling

### Refresh Channel

The channel returned by `Start()` receives:
- `nil` - Successful refresh
- `error` - Refresh failed with specific error, if no error is returned, the refresh was successful

```go
refreshCh := autoFetcher.Start(ctx)

for err := range refreshCh {
    if err != nil {
        switch {
        case errors.Is(err, autofetcher.ErrStoredListOfTokenListsIsEmpty):
            log.Println("No cached data available and fetch failed")
        case errors.Is(err, autofetcher.ErrFetcherNotProvided):
            log.Println("Fetcher not provided")
        case errors.Is(err, autofetcher.ErrContentStoreNotProvided):
            log.Println("Content store not provided")
        default:
            log.Printf("Refresh error: %v", err)
        }
    } else {
        log.Println("Refresh successful")
    }
}
```

### Common Errors

- `ErrAutoRefreshCheckIntervalGreaterThanInterval` - Invalid interval configuration (check interval must be <= refresh interval)
- `ErrRemoteListOfTokenListsParserNotProvided` - Missing parser for remote lists
- `ErrTokenListsNotProvided` - Empty token lists in configuration
- `ErrFetcherNotProvided` - Fetcher is nil
- `ErrContentStoreNotProvided` - ContentStore is nil
- `ErrStoredListOfTokenListsIsEmpty` - No cached data and remote fetch failed

### Validation

Both configuration types provide a `Validate()` method that checks

```go
if err := config.Validate(); err != nil {
    // Handle validation error
    log.Fatal(err) // ErrAutoRefreshCheckIntervalGreaterThanInterval
}
```

## Advanced Usage

### Custom Refresh Intervals

```go
config := autofetcher.ConfigTokenLists{
    Config: autofetcher.Config{
        LastUpdate:               time.Now().Add(-time.Hour),
        AutoRefreshInterval:      15 * time.Minute,  // Refresh every 15 minutes
        AutoRefreshCheckInterval: 30 * time.Second,  // Check every 30 seconds
    },
    TokenLists: tokenLists,
}
```

### Lifecycle Management

```go
// Start fetching
refreshCh := fetcher.Start(ctx)

// Multiple Start calls are safe - returns same channel
anotherRefreshCh := fetcher.Start(ctx)
// refreshCh == anotherRefreshCh

// Stop gracefully
fetcher.Stop()

// Multiple Stop calls are safe
fetcher.Stop() // No-op

// Start again after stopping gets new channel
newRefreshCh := fetcher.Start(ctx)
// newRefreshCh != refreshCh
```

## Thread Safety

The autofetcher is **fully thread-safe**:

- **Multiple goroutines** can call `Start()` and `Stop()` concurrently
- **Background fetching** runs in separate goroutine without race conditions
- **ContentStore access** is synchronized appropriately
- **Channel operations** are properly coordinated

**ContentStore Requirement**: Your ContentStore implementation MUST be thread-safe.

## Testing

The package includes comprehensive tests covering:

- Configuration validation
- Lifecycle management (Start/Stop)
- Concurrent operations
- Error conditions
- Refresh logic for both patterns
- Thread safety
- ETag handling
- Fallback mechanisms

```bash
# Run tests
go test ./pkg/tokens/autofetcher/...

# Run with race detection
go test -race ./pkg/tokens/autofetcher/...

# Run with verbose output
go test -v ./pkg/tokens/autofetcher/...
```