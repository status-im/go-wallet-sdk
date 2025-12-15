# ENS Resolver Example

A command-line tool demonstrating how to use the `pkg/ens` package for Ethereum Name Service (ENS) resolution.

## Features

- **Forward Resolution**: Convert ENS names (e.g., `vitalik.eth`) to Ethereum addresses
- **Reverse Resolution**: Convert Ethereum addresses to ENS names
- **Chain Detection**: Automatically detects and displays the connected chain
- **Error Handling**: Comprehensive error messages for common issues
- **Timeout Control**: Configurable timeout for operations

## Prerequisites

- Go 1.23.0 or later
- Access to an Ethereum RPC endpoint (Mainnet or Sepolia)
  - Get one from [Infura](https://infura.io) or [Alchemy](https://alchemy.com)
  - Or run your own Ethereum node

## Build

```bash
cd examples/ens-resolver-example
go build
```

This will create an executable named `ens-resolver-example` (or `ens-resolver-example.exe` on Windows).

## Run

### Basic Usage

**Forward Resolution** (ENS name → address):
```bash
./ens-resolver-example -rpc https://mainnet.infura.io/v3/YOUR-PROJECT-ID -name vitalik.eth
```

**Reverse Resolution** (address → ENS name):
```bash
./ens-resolver-example -rpc https://mainnet.infura.io/v3/YOUR-PROJECT-ID -address 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
```

### Command-Line Flags

| Flag | Description | Required | Default |
|------|-------------|----------|---------|
| `-rpc` | Ethereum RPC endpoint URL | Yes | - |
| `-name` | ENS name for forward resolution | No* | - |
| `-address` | Ethereum address for reverse resolution | No* | - |
| `-timeout` | Operation timeout duration | No | 30s |
| `-help` | Show help message | No | false |

*Note: Either `-name` or `-address` must be provided, but not both.

### Examples

#### Forward Resolution

Resolve an ENS name to an Ethereum address:

```bash
./ens-resolver-example -rpc https://mainnet.infura.io/v3/YOUR-PROJECT-ID -name vitalik.eth
```

Output:
```
Connecting to RPC endpoint: https://mainnet.infura.io/v3/YOUR-PROJECT-ID
Connected to Ethereum Mainnet (Chain ID: 1)

Resolving ENS name: vitalik.eth
----------------------------------------
✓ Successfully resolved
  ENS Name: vitalik.eth
  Address:  0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
```

#### Reverse Resolution

Resolve an Ethereum address to an ENS name:

```bash
./ens-resolver-example -rpc https://mainnet.infura.io/v3/YOUR-PROJECT-ID -address 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
```

Output:
```
Connecting to RPC endpoint: https://mainnet.infura.io/v3/YOUR-PROJECT-ID
Connected to Ethereum Mainnet (Chain ID: 1)

Performing reverse resolution for: 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
----------------------------------------
✓ Successfully resolved
  Address:  0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
  ENS Name: vitalik.eth
```


#### Using Sepolia Testnet

```bash
./ens-resolver-example -rpc https://sepolia.infura.io/v3/YOUR-PROJECT-ID -name test.eth
```

#### Custom Timeout

Set a custom timeout for slow networks:

```bash
./ens-resolver-example -rpc https://mainnet.infura.io/v3/YOUR-PROJECT-ID -name vitalik.eth -timeout 1m
```

## Supported Chains

ENS resolution is only available on:

- **Ethereum Mainnet** (Chain ID: 1)
- **Sepolia Testnet** (Chain ID: 11155111)
- **Holesky Testnet** (Chain ID: 17000)

Attempting to use other chains will result in an error.

## Error Handling

The tool provides clear error messages for common issues:

- Invalid ENS Name
- No Reverse Record
- Unsupported Chain
- Connection Error

