# Balance Fetcher Package

High-performance balance fetching for EVM-compatible blockchains.

## Features

- 🚀 **Batch balance fetching** for multiple addresses
- 🔄 **Automatic fallback**: BalanceScanner contract → standard RPC calls
- 🧪 **Testable**: Includes interfaces and mock support
- ⛓️ **Chain-agnostic**: Works with any EVM-compatible chain

## Quick Usage

```go
import (
    "context"
    "github.com/status-im/go-wallet-sdk/pkg/balance/fetcher"
)

// Fetch balances for multiple addresses
balances, err := fetcher.FetchNativeBalances(
    context.Background(), 
    addresses,           // []common.Address
    atBlock,            // block number (nil for latest)
    rpcClient,          // must implement fetcher.RPCClient
    batchSize,          // addresses per batch (e.g., 10)
)

if err != nil {
    // handle error
}

for addr, bal := range balances {
    fmt.Printf("%s: %s wei\n", addr.Hex(), bal.String())
}
```

## Interfaces

- `RPCClient`: Minimal interface for RPC calls (compatible with go-ethereum clients)
- `BatchCaller`: Interface for batch RPC calls
- `BalanceScanner`: Interface for BalanceScanner contract calls

## Testing

```go
import "github.com/status-im/go-wallet-sdk/pkg/balance/fetcher/mock"

mockRPC := mock.NewMockRPCClient(ctrl)
// Configure mock as needed for your tests
```

## File Structure

- `fetcher.go` - Main interface and entry point
- `fetcher_balancescanner.go` - BalanceScanner contract implementation
- `fetcher_standard.go` - Standard RPC implementation
- `types.go` - Shared types/interfaces
- `utils.go` - Helper functions
- `mock/` - Mocks for testing

## Dependencies

- [go-ethereum](https://github.com/ethereum/go-ethereum) for types and RPC interfaces 