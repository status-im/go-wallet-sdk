# Balance Fetcher Package

High-performance balance fetching for EVM-compatible blockchains.

## Use it when

- You need native balance fetching for many addresses efficiently.
- You need ERC20 balances for many (account, token) pairs with batching.
- You want a chain-agnostic balance fetcher that can use Multicall3 where available.

## Key entrypoints

- `fetcher.FetchNativeBalances(ctx, addresses, atBlock, rpcClient, batchSize)`
- `fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, rpcClient, batchSize)`
- Interfaces: `fetcher.RPCClient`, `fetcher.BatchCaller`, `multicall.Caller`

## Features

- **Batch balance fetching** for multiple addresses and ERC20 tokens in fewer calls
- **Chain-agnostic**: Works with any EVM-compatible chain

## Quick Usage

### Native Token Balances

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

### ERC20 Token Balances

```go
import (
    "context"
    "math/big"
    "github.com/status-im/go-wallet-sdk/pkg/balance/fetcher"
    // ... your RPC client import
)

// addresses: slice of common.Address (account addresses)
// tokenAddresses: slice of common.Address (ERC20 token contract addresses)
// atBlock: block number (use nil for latest)
// rpcClient: must implement fetcher.RPCClient and fetcher.BatchCaller
// batchSize: number of calls per batch (e.g., 10)

balances, err := fetcher.FetchErc20Balances(context.Background(), addresses, tokenAddresses, atBlock, rpcClient, batchSize)
if err != nil {
    // handle error
}

// balances is a map[common.Address]map[common.Address]*big.Int
// balances[accountAddress][tokenAddress] = balance
for accountAddr, tokenBalances := range balances {
    for tokenAddr, balance := range tokenBalances {
        fmt.Printf("Account %s, Token %s: %s\n", accountAddr.Hex(), tokenAddr.Hex(), balance.String())
    }
}
```

## Interfaces

- `RPCClient`: Minimal interface for RPC calls (compatible with go-ethereum clients)
- `BatchCaller`: Interface for batch RPC calls
- `multicall.Caller`: Interface for Multicall3 contract calls

## File Structure

- `fetcher.go` - Main interface and entry point
- `fetcher_multicall.go` - Multicall3 contract implementation
- `fetcher_standard.go` - Standard RPC implementation
- `types.go` - Shared types/interfaces
- `utils.go` - Helper functions
- `mock/` - Mocks for testing

## Dependencies

- [go-ethereum](https://github.com/ethereum/go-ethereum) for types and RPC interfaces

## See Also

- [Multi-Standard Fetcher](../multistandardfetcher/README.md) - Fetch multiple token standards at once
- [Multicall Package](../../multicall/README.md) - Low-level batching control
- [Ethereum Client](../../ethclient/README.md) - RPC client for balance queries

## Examples

- [Balance Fetcher Web](../../../examples/balance-fetcher-web/README.md) - Web interface for balance fetching
- [Multi-Standard Fetcher Example](../../../examples/multistandardfetcher-example/README.md) - Complete example with multiple token types
 