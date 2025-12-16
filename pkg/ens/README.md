# ENS Package

The `ens` package provides Ethereum Name Service (ENS) resolution capabilities. It supports both forward resolution (ENS name to Ethereum address) and reverse resolution (Ethereum address to ENS name).

## Use it when

- You need forward resolution (name → address) like `vitalik.eth`.
- You need reverse resolution (address → name) for display purposes.
- You want to check ENS availability on a chain before attempting resolution.

## Key entrypoints

- `ens.NewResolver(client)`
- `(*Resolver).AddressOf(name)` and `(*Resolver).GetName(address)`
- `ens.IsSupportedChain(chainID)`

## Features

- **Forward Resolution**: Convert ENS names (e.g., `vitalik.eth`) to Ethereum addresses
- **Reverse Resolution**: Convert Ethereum addresses to ENS names
- **Chain Detection**: Check if ENS is available on the connected chain
- **Input Validation**: Validates ENS name structure and Ethereum addresses before resolution

## Supported Chains

ENS is available on:
- Ethereum Mainnet (Chain ID: 1)
- Sepolia Testnet (Chain ID: 11155111)
- Holesky Testnet (Chain ID: 17000)

Use `IsSupportedChain(chainID)` to check if ENS is available for a given chain.

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ethereum/go-ethereum/rpc"
    "github.com/status-im/go-wallet-sdk/pkg/ens"
    "github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

func main() {
    // Connect to Ethereum node
    rpcClient, err := rpc.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
    if err != nil {
        log.Fatal(err)
    }
    defer rpcClient.Close()

    client := ethclient.NewClient(rpcClient)

    // Optional: Check if ENS is available on this chain
    chainID, err := client.ChainID(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    if !ens.IsSupportedChain(chainID.Uint64()) {
        log.Fatal("ENS is not available on this chain")
    }

    // Create ENS resolver
    resolver, err := ens.NewResolver(client)
    if err != nil {
        log.Fatal(err)
    }

    // Forward resolution: name -> address
    address, err := resolver.AddressOf("vitalik.eth")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Address: %s\n", address.Hex())

    // Reverse resolution: address -> name
    name, err := resolver.GetName(address)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Name: %s\n", name)
}
```

### Forward Resolution

Convert an ENS name to an Ethereum address:

```go
address, err := resolver.AddressOf("vitalik.eth")
if err != nil {
    if errors.Is(err, ens.ErrInvalidName) {
        // Handle invalid name format
    }
    // Handle other errors
}
fmt.Printf("Address: %s\n", address.Hex())
```

### Reverse Resolution

Convert an Ethereum address to an ENS name:

```go
address := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
name, err := resolver.GetName(address)
if err != nil {
    // Handle error (address may not have a reverse record)
}
fmt.Printf("Name: %s\n", name)
```

## Limitations

- **Wrapped names are not supported**: This package does not support ENS wrapped names (names using the NameWrapper contract).

## See Also

- [Ethereum Client](../ethclient/README.md) - RPC client for ENS calls
- [Token Manager](../tokens/manager/README.md) - Manage token metadata
- [Account Management](../accounts/extkeystore/README.md) - Manage wallet addresses

## Examples

- [ENS Resolver Example](../../examples/ens-resolver-example/README.md) - CLI tool for ENS resolution
- [Accounts Example](../../examples/accounts/README.md) - Web interface with ENS integration (potential)

