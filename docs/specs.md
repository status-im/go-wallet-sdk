
## 1. Overview and Goals

Go Wallet SDK is a modular Go library intended to support the development of multi‑chain cryptocurrency wallets and blockchain applications. The SDK exposes self‑contained packages for common wallet tasks such as fetching account balances across many EVM chains and interacting with Ethereum JSON‑RPC.

### 1.1 Main Repository Components

| Component             | Purpose                                                                                                                                                                                                                                                    |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/balance/fetcher` | High‑performance balance fetcher for EVM‑compatible chains. Fetches native and token balances across many addresses. It prefers Multicall3 for maximum efficiency and falls back to standard JSON‑RPC batching when Multicall3 is unavailable. |
| `pkg/multicall`       | Utilities around the Multicall3 contract: call builders for native/`ERC20`/`ERC721`/`ERC1155` balance queries and a job runner that batches calls into as few on‑chain multicalls as possible and processes results. |
| `pkg/contracts`       | Auto‑generated Go bindings for on‑chain contracts used by the SDK, including `erc20`, `erc721`, `erc1155`, and `multicall3` (with a registry of deployments across 250+ chains). |
| `pkg/ethclient`       | Chain‑agnostic Ethereum JSON‑RPC client.  It provides two method sets: a drop‑in replacement compatible with go‑ethereum’s `ethclient` and a custom implementation that follows the Ethereum JSON‑RPC specification without assuming chain‑specific types. It supports JSON‑RPC methods covering `eth_`, `net_` and `web3_` namespace |
| `pkg/common`          | Shared types and constants. Such as canonical chain IDs (e.g., Ethereum Mainnet, Optimism, Arbitrum, BSC, Base). Developers use these values when configuring the SDK or examples.                               |
| `cshared/`            | C shared library bindings that expose core SDK functionality to C applications. |
| `examples/`           | Demonstrations of SDK usage.  Includes `balance-fetcher-web` (a web interface for batch balance fetching), `ethclient‑usage` (an example that exercises the Ethereum client across multiple RPC endpoints), and `c-app` (a C application demonstranting usage of the C library usage).                                             |                                                                                                                                                 |

## 2. Architecture

### 2.1 High‑level Structure

Go Wallet SDK follows a modular architecture where each package encapsulates a specific functional domain. There is no central runtime; instead applications import only the packages they need. The SDK currently focuses on EVM‑compatible chains, leaving room for additional chain types in the future. The packages are:
- **Balance Fetcher** – Provides efficient methods to retrieve account balances (native or ERC‑20) across many addresses and tokens. It abstracts over RPC batch calls and an on‑chain BalanceScanner contract. Developers supply a minimal RPC client interface (`RPCClient` and optionally `BatchCaller`) and the package returns a map of balances
- **Ethereum Client** – Exposes the full Ethereum JSON‑RPC API. It wraps a standard RPC client and offers two sets of methods: chain‑agnostic versions prefixed with `Eth*` and a drop‑in `BalanceAt`, `BlockNumber` etc. that mirror go‑ethereum’s ethclient. The client covers methods including network info, block and transaction queries, account state, contract code and gas estimation
- **Common Utilities** – Houses shared types (e.g., `ChainID`) and enumerated constants for well‑known networks. This allows examples and client code to refer to network IDs without hard‑coding numbers.

The SDK emphasises chain agnosticism: methods do not assume particular transaction formats or gas pricing models and therefore work with Ethereum, L2 networks (Optimism, Arbitrum, Polygon), and other EVM‑compatible chains. Each package hides chain‑specific details behind simple interfaces.

### 2.2 Balance Fetcher Design

The balance fetcher is designed to efficiently query balances for many addresses and tokens. Its design includes:

- **Dual fetch strategies** – The package prefers Multicall3 to aggregate many reads into a small number of on‑chain calls. If Multicall3 is unavailable on a given chain, it falls back to standard JSON‑RPC batching (`eth_getBalance` and `eth_call`). Both strategies are exposed transparently through the same API.
- **Batching and concurrency** – With Multicall3, the job runner batches heterogeneous calls (native/`ERC20`/`ERC721`/`ERC1155`) and anchors all chunks to a single block using `tryBlockAndAggregate` for the first chunk and `tryAggregate` for subsequent chunks. In standard mode, the fetcher groups RPC requests into chunks (`batchSize`) to reduce round‑trips and aggregates results into maps keyed by address and token.
- **Chain‑agnostic** – The logic is unaware of specific chain parameters; it accepts any RPC endpoint and optionally a block number. A `ChainID` from `pkg/common` can be used to label results, but the fetcher does not require it.

### 2.3 Ethereum Client Design

The Ethereum client package (`pkg/ethclient`) wraps a generic RPC client and exposes two categories of methods:

- **Go‑ethereum‑compatible methods** – Methods such as `BlockNumber`, `BalanceAt` and `TransactionByHash` mimic the ethclient interface from go‑ethereum so existing applications can switch to this SDK with minimal changes. These methods require a go‑ethereum RPC client (because they call underlying types) and may not work on Layer 2 chains that diverge from Ethereum’s API.
- **Chain‑agnostic methods** – Methods prefixed with Eth* correspond directly to Ethereum JSON‑RPC calls and accept/return standard Go types. Examples include `EthBlockNumber`, `EthGetBalance`, `EthGasPrice`, `EthGetBlockByNumberWithFullTxs`, `EthGetLogs`, and `EthEstimateGas`. These functions rely only on the JSON‑RPC specification and therefore support any EVM‑compatible chain.

Internally, the client stores a reference to an RPC client and implements each method by calling `rpcClient.CallContext` with the appropriate RPC method name and parameters (see eth.go). It deserialises responses into exported Go types or custom structs (e.g., `BlockWithTxHashes`, `BlockWithFullTxs`). The design includes convenience functions for converting block numbers to RPC arguments and decoding hex‑encoded values.

### 2.4 Common Utilities

The `pkg/common` package defines shared types and enumerations. The main export is `type ChainID uint64` with constants for well‑known networks such as `EthereumMainnet`, `EthereumSepolia`, `OptimismMainnet`, `ArbitrumMainnet`, `BSCMainnet`, `BaseMainnet`, `BaseSepolia` and a custom `StatusNetworkSepolia`. These constants allow the examples to pre‑populate supported chains and label results without repeating numeric IDs.

### 2.5 C Library

At `cshared/lib.go` the library functions are exposed to be used as C bindings for core SDK functionality, enabling integration with C applications and other languages that can interface with C libraries.
The shared library is built using Go's `c-shared` build mode (e.g `go build -buildmode=c-shared -o lib.so lib.go`), which generates both the library file (`.so` on Linux, `.dylib` on macOS) and a corresponding C header file with function declarations and type definitions.

### 2.6 Multicall Utilities

`pkg/multicall` provides:

- Call builders: `BuildNativeBalanceCall`, `BuildERC20BalanceCall`, `BuildERC721BalanceCall`, `BuildERC1155BalanceCall`, plus `Process*Result` helpers to decode `IMulticall3.Result` payloads into `*big.Int` balances.
- A job runner to execute many calls efficiently against Multicall3:
  - `RunSync(ctx, jobs, atBlock, caller, batchSize) []JobResult`
  - `ProcessJobRunners(ctx, jobRunners, atBlock, caller, batchSize)`

The runner uses `tryBlockAndAggregate` to return `(blockNumber, blockHash, results)` on the first chunk and `tryAggregate` for subsequent chunks pinned to the same block, ensuring all results are consistent to a single block.

## 3. API Description

### 3.1 Balance Fetcher API (`pkg/balance/fetcher`)

The balance fetcher exposes two primary functions:

| Function                                                                            | Purpose                                                                                                                                                                                                | Parameters                                                                                                                                                                                                                                              | Returns                                                                                                                                                     |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FetchNativeBalances(ctx, addresses, atBlock, rpcClient, batchSize)`                | Retrieves native token balances (e.g., ETH) for multiple addresses. Prefers Multicall3 when available and falls back to batched `eth_getBalance` RPC calls.                                           | `ctx`: context; `addresses`: slice of addresses; `atBlock`: block number or `nil` for latest; `rpcClient`: implements `RPCClient`; `batchSize`: maximum addresses per batch.                                                                            | A map `map[common.Address]*big.Int` associating each address with its balance.  Errors indicate network issues or RPC failures.                             |
| `FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, rpcClient, batchSize)` | Retrieves ERC‑20 token balances for multiple addresses and tokens. Prefers Multicall3 when available and falls back to batched `eth_call` of `balanceOf` for each (address, token) pair.               | `ctx`: context; `addresses`: slice of account addresses; `tokenAddresses`: slice of ERC‑20 contract addresses; `atBlock`: block number or `nil`; `rpcClient`: implements `RPCClient` and `BatchCaller`; `batchSize`: maximum number of calls per batch. | A nested map `map[address]map[token]*big.Int` where `balances[account][token]` is the token balance.  Errors indicate RPC failures or contract call errors. |

More specific functions are also available:

| Function                                                                                                          | Description                                                                                                                                                                                                                  |
| ----------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, multicallCaller, multicallAddress, batchSize)`        | Builds `getEthBalance` calls and executes them via Multicall3; decodes results to balances. The first chunk uses `tryBlockAndAggregate` to return the block number/hash.                                                      |
| `FetchErc20BalancesWithMulticall(ctx, accountAddresses, tokenAddresses, atBlock, multicallCaller, batchSize)`    | Builds `ERC20.balanceOf` calls for each (account, token) and executes them via Multicall3; decodes results to balances.                                                                                                      |
| `FetchNativeBalancesStandard(ctx, addresses, atBlock, batchCaller, batchSize)`                                    | Constructs `eth_getBalance` batch requests using the provided `BatchCaller`; decodes hex strings into big.Int balances.                                                                                                      |
| `FetchErc20BalancesStandard(ctx, addresses, tokenAddresses, atBlock, batchCaller, batchSize)`                     | Builds `eth_call` requests for each account/token pair using the ERC‑20 ABI and sends them in batches.                                                                                                                       |

**Multicall3 Deployments and Usage**

Multicall3 is widely deployed across 250+ EVM chains, commonly at `0xCA11bde05977b3631167028862bE2a173976CA11` (case insensitive). The SDK provides a generated registry at `pkg/contracts/multicall3/deployments.go` with helpers:

- `multicall3.GetMulticall3Address(chainID int64) (common.Address, bool)`
- `multicall3.IsChainSupported(chainID int64) bool`
- `multicall3.GetSupportedChainIDs() []int64`

See `pkg/contracts/multicall3/deployments/` for the generator that ingests upstream deployment metadata.

Within `pkg/multicall`, the `Caller` interface abstracts the `multicall3` view methods used by the job runner:

```12:26:pkg/multicall/runner.go
type Caller interface {
    ViewTryBlockAndAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error)
    ViewTryAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error)
}
```

The balance fetcher integrates with the job runner, e.g. native balances:

```14:46:pkg/balance/fetcher/fetcher_multicall.go
func FetchNativeBalancesWithMulticall(
    ctx context.Context,
    accountAddresses []common.Address,
    atBlock gethrpc.BlockNumber,
    multicallCaller multicall.Caller,
    multicallAddress common.Address,
    batchSize int,
) (BalancePerAccountAddress, error) {
    // builds getEthBalance calls, runs RunSync, decodes results
}
```

**ERC‑20/721/1155 ABI Usage**

- Uses generated bindings in `pkg/contracts/erc20`, `pkg/contracts/erc721`, and `pkg/contracts/erc1155` or packs `balanceOf(...)` via the ABI for `eth_call` in standard mode.
- In standard mode, `balanceOf` is encoded with `abi.Pack("balanceOf", accountAddress)` (and `tokenID` for `ERC1155`) and sent as `input` to the token contract `to` address.

**Regenerating Bindings**

The SDK uses auto-generated Go bindings to interact with smart contracts. These bindings provide type-safe method calls and handle ABI encoding/decoding automatically.

Use `abigen` to regenerate bindings when contract sources are updated:

```bash
# ERC-20 from Solidity interface  
abigen --sol pkg/contracts/erc20/IERC20.sol --pkg erc20 --out pkg/contracts/erc20/erc20.go

# ERC-721 from Solidity interface
abigen --sol pkg/contracts/erc721/IERC721.sol --pkg erc721 --out pkg/contracts/erc721/erc721.go

# ERC-1155 from Solidity interface
abigen --sol pkg/contracts/erc1155/IERC1155.sol --pkg erc1155 --out pkg/contracts/erc1155/erc1155.go

# Alternative: Generate from ABI JSON (if available)
abigen --abi IERC20.abi.json --pkg erc20 --out pkg/contracts/erc20/erc20.go

# Multicall3 bindings are generated from `IMulticall3.sol` and the deployments registry is generated via:
cd pkg/contracts/multicall3/deployments && go run .
```

Ensure the ABI/Solidity sources match the deployed contract versions. Regeneration is needed when contracts are upgraded, ABIs change, or new functionality is added.

### 3.2 Ethereum Client API (`pkg/ethclient`)

This package exports a `Client` type that wraps a lower‑level RPC client and provides both go‑ethereum‑compatible methods and chain‑agnostic methods. Developers construct a client using `NewClient(rpcClient)`. If the provided RPC client is a go‑ethereum `rpc.Client`, the SDK internally also creates a `gethEthClient` for compatibility with existing `ethclient` code

The Ethereum client exposes a large set of methods. They can be grouped into several categories. All methods follow the same pattern of accepting a context and returning typed data or errors.

| Method                   | Description                                          | Example                                                                                |
| ------------------------ | ---------------------------------------------------- | -------------------------------------------------------------------------------------- |
| `Web3ClientVersion(ctx)` | Returns the version of the Ethereum client software | `client.Web3ClientVersion(ctx)` returns a string like `"Geth/v1.16.0-stable/linux"` |

**Net Namespace**

| Method               | Description                                                 | Example                                                                          |
| -------------------- | ----------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `NetListening(ctx)`  | Returns whether the client is actively listening for peers | `client.NetListening(ctx)` returns `true` if listening                          |
| `NetPeerCount(ctx)`  | Returns the number of connected peers                       | `client.NetPeerCount(ctx)` returns `uint64` peer count                          |
| `NetVersion(ctx)`    | Returns the network ID as a string                          | `client.NetVersion(ctx)` returns `"1"` for Mainnet, `"11155111"` for Sepolia   |

**Eth Namespace - Node/Network Information**

| Method                           | Description                                                  | Example                                                                    |
| -------------------------------- | ------------------------------------------------------------ | -------------------------------------------------------------------------- |
| `EthProtocolVersion(ctx)`        | Returns the Ethereum protocol version                        | `client.EthProtocolVersion(ctx)` returns `"0x41"` (protocol version 65)   |
| `EthChainId(ctx)`                | Returns the chain ID as a big integer                       | `client.EthChainId(ctx)` returns `*big.Int` with value `1` for Mainnet    |
| `EthSyncing(ctx)`                | Returns sync status or false if not syncing                 | `client.EthSyncing(ctx)` returns `*ethereum.SyncProgress` or `false`      |
| `EthCoinbase(ctx)`               | Returns the coinbase address (mining reward recipient)      | `client.EthCoinbase(ctx)` returns `common.Address`                        |
| `EthMining(ctx)`                 | Returns whether the client is mining                        | `client.EthMining(ctx)` returns `true` if mining                          |
| `EthHashrate(ctx)`               | Returns the mining hashrate in hashes per second            | `client.EthHashrate(ctx)` returns `uint64` hashrate                       |
| `EthMaxPriorityFeePerGas(ctx)`   | Returns suggested priority fee for EIP‑1559 transactions    | `client.EthMaxPriorityFeePerGas(ctx)` returns `*big.Int` in wei           |
| `EthBlobBaseFee(ctx)`            | Returns the base fee for blob transactions (EIP‑4844)       | `client.EthBlobBaseFee(ctx)` returns `*big.Int` base fee in wei           |

**Eth Namespace - Blocks**

| Method                                          | Description                                                          | Example                                                                        |
| ----------------------------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `EthBlockNumber(ctx)`                           | Returns the number of the most recent block                         | `client.EthBlockNumber(ctx)` returns `uint64` block number                    |
| `EthGetBlockByHashWithTxHashes(ctx, hash)`      | Fetches a block by hash with transaction hashes only                | `client.EthGetBlockByHashWithTxHashes(ctx, blockHash)`                        |
| `EthGetBlockByNumberWithTxHashes(ctx, number)`  | Fetches a block by number with transaction hashes only              | `client.EthGetBlockByNumberWithTxHashes(ctx, big.NewInt(19543210))`           |
| `EthGetBlockByHashWithFullTxs(ctx, hash)`       | Fetches a block by hash with full transaction objects               | `client.EthGetBlockByHashWithFullTxs(ctx, blockHash)`                         |
| `EthGetBlockByNumberWithFullTxs(ctx, number)`   | Fetches a block by number with full transaction objects             | `client.EthGetBlockByNumberWithFullTxs(ctx, big.NewInt(19543210))`            |
| `EthGetBlockReceipts(ctx, number)`              | Returns all transaction receipts for a given block                  | `client.EthGetBlockReceipts(ctx, big.NewInt(19543210))`                       |
| `EthGetBlockTransactionCountByHash(ctx, hash)`  | Returns the number of transactions in a block by hash               | `client.EthGetBlockTransactionCountByHash(ctx, blockHash)`                    |
| `EthGetBlockTransactionCountByNumber(ctx, num)` | Returns the number of transactions in a block by number             | `client.EthGetBlockTransactionCountByNumber(ctx, big.NewInt(19543210))`       |
| `EthGetUncleByBlockHashAndIndex(ctx, hash, i)`  | Returns uncle block by block hash and uncle index                   | `client.EthGetUncleByBlockHashAndIndex(ctx, blockHash, 0)`                    |
| `EthGetUncleByBlockNumberAndIndex(ctx, num, i)` | Returns uncle block by block number and uncle index                 | `client.EthGetUncleByBlockNumberAndIndex(ctx, big.NewInt(19543210), 0)`       |
| `EthGetUncleCountByBlockHash(ctx, hash)`        | Returns the number of uncles in a block by hash                     | `client.EthGetUncleCountByBlockHash(ctx, blockHash)`                          |
| `EthGetUncleCountByBlockNumber(ctx, number)`    | Returns the number of uncles in a block by number                   | `client.EthGetUncleCountByBlockNumber(ctx, big.NewInt(19543210))`             |

**Eth Namespace - Transactions**

| Method                                                | Description                                                   | Example                                                                    |
| ----------------------------------------------------- | ------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `EthSendRawTransaction(ctx, rawTx)`                   | Submits a signed transaction to the network                  | `client.EthSendRawTransaction(ctx, signedTxBytes)`                         |
| `EthSendTransaction(ctx, tx)`                         | Submits a transaction using a managed account                | `client.EthSendTransaction(ctx, txObject)` (requires unlocked account)    |
| `EthGetTransactionByHash(ctx, hash)`                  | Returns transaction details by transaction hash              | `client.EthGetTransactionByHash(ctx, txHash)`                             |
| `EthGetTransactionByBlockHashAndIndex(ctx, hash, i)`  | Returns transaction by block hash and transaction index      | `client.EthGetTransactionByBlockHashAndIndex(ctx, blockHash, 0)`          |
| `EthGetTransactionByBlockNumberAndIndex(ctx, num, i)` | Returns transaction by block number and transaction index    | `client.EthGetTransactionByBlockNumberAndIndex(ctx, big.NewInt(123), 0)`  |
| `EthGetTransactionReceipt(ctx, hash)`                 | Returns the receipt of a transaction by hash                 | `client.EthGetTransactionReceipt(ctx, txHash)`                            |
| `EthGetTransactionCount(ctx, address, atBlock)`       | Returns the nonce (transaction count) for an account        | `client.EthGetTransactionCount(ctx, myAddress, nil)`                      |
| `EthSign(ctx, addr, data)`                            | Signs arbitrary data with an account's private key          | `client.EthSign(ctx, myAddress, dataToSign)`                              |
| `EthSignTransaction(ctx, tx)`                         | Signs a transaction without sending it                       | `client.EthSignTransaction(ctx, txObject)`                                |

**Eth Namespace - Account/State**

| Method                                        | Description                                                     | Example                                                           |
| --------------------------------------------- | --------------------------------------------------------------- | ----------------------------------------------------------------- |
| `EthGetBalance(ctx, address, atBlock)`        | Returns the balance of an account at a given block             | `client.EthGetBalance(ctx, myAddress, nil)`                      |
| `EthGetCode(ctx, address, atBlock)`           | Returns the contract code at an address                        | `client.EthGetCode(ctx, contractAddr, nil)`                      |
| `EthGetStorageAt(ctx, address, key, atBlock)` | Returns the value from a storage position at an address        | `client.EthGetStorageAt(ctx, contractAddr, storageKey, nil)`     |
| `EthGetProof(ctx, address, keys, atBlock)`    | Returns account and storage proofs for Merkle verification     | `client.EthGetProof(ctx, myAddress, []string{storageKey}, nil)`  |

**Eth Namespace - Gas**

| Method                                                   | Description                                                    | Example                                                           |
| -------------------------------------------------------- | -------------------------------------------------------------- | ----------------------------------------------------------------- |
| `EthGasPrice(ctx)`                                       | Returns the current gas price in wei                          | `client.EthGasPrice(ctx)` returns `*big.Int`                     |
| `EthEstimateGas(ctx, callMsg)`                           | Estimates the gas required to execute a transaction           | `client.EthEstimateGas(ctx, callMsg)` returns `uint64`           |
| `EthFeeHistory(ctx, count, lastBlock, rewardPercentiles)` | Returns historical base fee and priority fee data             | `client.EthFeeHistory(ctx, 10, nil, []float64{25, 50, 75})`      |

**Eth Namespace - Call/Logs/Filters**

| Method                              | Description                                                   | Example                                                      |
| ----------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------ |
| `EthCall(ctx, callMsg, atBlock)`    | Executes a read‑only contract call without creating a tx     | `client.EthCall(ctx, callMsg, nil)`                         |
| `EthGetLogs(ctx, filterQuery)`      | Returns event logs matching a filter query                   | `client.EthGetLogs(ctx, filterQuery)`                       |
| `EthNewFilter(ctx, filterQuery)`    | Creates a new log filter and returns its ID                  | `client.EthNewFilter(ctx, filterQuery)` returns filter ID   |
| `EthNewBlockFilter(ctx)`            | Creates a new block filter and returns its ID                | `client.EthNewBlockFilter(ctx)` returns filter ID           |
| `EthGetFilterLogs(ctx, filterID)`   | Returns all logs for a filter (only for log filters)        | `client.EthGetFilterLogs(ctx, filterID)`                    |
| `EthGetFilterChanges(ctx, filterID)`| Returns new entries since last poll for any filter type      | `client.EthGetFilterChanges(ctx, filterID)`                 |
| `EthUninstallFilter(ctx, filterID)` | Uninstalls a filter and stops polling                        | `client.EthUninstallFilter(ctx, filterID)` returns `bool`   |

The chain‑agnostic methods (prefixed with `Eth*`, `Net*`, `Web3*`) correspond directly to Ethereum JSON‑RPC calls and accept/return standard Go types, making them compatible with any EVM‑compatible chain. For backward compatibility, the package also exports go‑ethereum compatible methods such as `BlockNumber(ctx)`, `BalanceAt(ctx, address, nil)`, etc., which call the same RPC methods but use go‑ethereum types.

**RPC Parameter Translation Helpers**

The Ethereum client includes several critical helper functions that bridge the gap between Go types and the specific JSON-RPC parameter formats required by Ethereum nodes. These helpers are essential because:

1. **Ethereum JSON-RPC has strict formatting requirements** - Parameters must be properly encoded as hex strings, structured objects, or special sentinel values
2. **Go types don't directly match RPC expectations** - Standard Go types like `*big.Int`, `ethereum.CallMsg`, and `ethereum.FilterQuery` need transformation
3. **Chain compatibility requires consistent encoding** - Different Ethereum clients expect the same standardized parameter formats

```go
// Block number encoder handling negative sentinel values for latest/finalized/etc.
func toBlockNumArg(number *big.Int) string

// Call and filter translators to RPC args
func toCallArg(msg ethereum.CallMsg) interface{}
func toFilterArg(q ethereum.FilterQuery) (interface{}, error)
```

**Block Number Translation (`toBlockNumArg`)**

Converts Go `*big.Int` block numbers into proper JSON-RPC format:
- `nil` → `"latest"` (most recent block)
- Positive numbers → hex-encoded strings (e.g., `big.NewInt(12345)` → `"0x3039"`)
- Special negative values → sentinel strings:
  - `-1` → `"pending"` (pending block)
  - `-2` → `"latest"` (latest mined block)  
  - `-3` → `"finalized"` (finalized block)
  - `-4` → `"safe"` (safe block)
  - `-5` → `"earliest"` (genesis block)

This is used by all block-parameter methods like `EthGetBalance`, `EthGetCode`, `EthCall`, etc.

**Call Message Translation (`toCallArg`)**

Converts Go `ethereum.CallMsg` structs into JSON-RPC call objects with proper hex encoding:
- Addresses → hex strings
- Data/input → hex-encoded bytes  
- Gas values → hex-encoded numbers
- Wei amounts → hex-encoded big integers
- EIP-1559 fee fields → properly formatted fee caps
- Access lists and blob parameters → structured objects

This ensures `EthCall`, `EthEstimateGas`, and transaction methods send correctly formatted parameters.

**Filter Query Translation (`toFilterArg`)**

Converts Go `ethereum.FilterQuery` structs into JSON-RPC filter objects:
- Address lists → arrays of hex-encoded addresses
- Topics → arrays of topic hashes with proper null handling
- Block ranges → properly formatted block parameters using `toBlockNumArg`
- Validates mutually exclusive parameters (block hash vs. block range)

This enables `EthGetLogs`, `EthNewFilter`, and other event filtering methods to work correctly across all EVM chains.

### 3.3 C Shared Library API (`cshared/`)

The C shared library provides a minimal but complete interface for blockchain operations from C applications. All functions use consistent patterns for error handling and memory management.

| Function | Description | Parameters | Returns |
| -------- | ----------- | ---------- | ------- |
| `GoWSK_NewClient(rpcURL, errOut)` | Creates a new Ethereum client connected to the specified RPC endpoint | `rpcURL`: null-terminated string with RPC URL; `errOut`: optional double pointer for error message | Opaque client handle (0 on failure) |
| `GoWSK_CloseClient(handle)` | Closes an Ethereum client and releases its resources | `handle`: client handle from `GoWSK_NewClient` | None |
| `GoWSK_ChainID(handle, errOut)` | Retrieves the chain ID for the connected network | `handle`: client handle; `errOut`: optional double pointer for error message | Chain ID as null-terminated string (must be freed) |
| `GoWSK_GetBalance(handle, address, errOut)` | Fetches the native token balance for an address | `handle`: client handle; `address`: hex-encoded Ethereum address; `errOut`: optional double pointer for error message | Balance in wei as null-terminated string (must be freed) |
| `GoWSK_FreeCString(s)` | Frees a string allocated by the library | `s`: string pointer returned by other functions | None |

**Usage Pattern**

All C applications follow the same basic pattern:

```c
#include "libgowalletsdk.h"

// Create client
char* err = NULL;
unsigned long long client = GoWSK_NewClient("https://mainnet.infura.io/v3/KEY", &err);
if (client == 0) {
    fprintf(stderr, "Error: %s\n", err);
    GoWSK_FreeCString(err);
    return 1;
}

// Use client APIs
char* chainID = GoWSK_ChainID(client, &err);
if (chainID) {
    printf("Chain ID: %s\n", chainID);
    GoWSK_FreeCString(chainID);
}

char* balance = GoWSK_GetBalance(client, "0x...", &err);
if (balance) {
    printf("Balance: %s wei\n", balance);
    GoWSK_FreeCString(balance);
}

// Always close client
GoWSK_CloseClient(client);
```

All string returns from the library are allocated with `malloc` and must be freed using `GoWSK_FreeCString`. Also Error messages returned via `errOut` parameters must also be freed

## 4. Example Applications

### 4.1 Web‑Based Balance Fetcher

The `examples/balance-fetcher-web` folder contains a complete web application that demonstrates how to use the balance fetcher. Key aspects include:
- **Features** – The web UI allows users to specify custom chains (chain ID and RPC URL), enter multiple Ethereum addresses and an optional block number, then fetch balances across chains. It prefers Multicall3 for aggregation (often querying thousands of balances in a single RPC call) and automatically falls back to standard RPC. It displays balances in both ETH and wei. The example pre‑populates common chains such as Ethereum, Optimism, Arbitrum and Polygon.
- **Usage** – Running `go run .` in the example directory starts an HTTP server on `localhost:8080`. Users can configure chains, input addresses and click Fetch Balances. The backend sends a `POST /fetch` request containing a JSON payload with chains, addresses and block number.
- **Project Structure** – The example is organised into `main.go` (entry point), `types.go` (data structures), `rpc_client.go` (custom RPC client), `utils.go`, `templates.go` (HTML/JS templates), and `handlers.go` (HTTP handlers).
- **Security Considerations** – The example warns that it is for demonstration only. Production deployments should secure RPC endpoints, implement authentication, validate user input and add rate‑limiting.

### 4.2 Ethereum Client Example

The `examples/ethclient-usage` folder shows how to use the Ethereum client across multiple networks. It exercises a wide range of RPC methods and demonstrates multi‑endpoint support.

- **Features** – The example tests connectivity and functionality across multiple RPC endpoints, retrieves network and blockchain data, account balances and nonces, contract code, filters events, retrieves transaction details, checks network status, and estimates gas. It highlights the chain‑agnostic benefits of the custom **eth.go** methods, which make no assumptions about transaction types or chain‑specific fields.

- **Usage** – Users specify one or more RPC endpoints via the **ETH_RPC_ENDPOINTS** environment variable and run **go run main.go**. The program iterates through each endpoint, prints client and network information, queries blocks and transactions, and demonstrates various API calls. Example output shows block and transaction details, balances, gas prices and event logs.

- **Configuration** – The example includes defaults for Ethereum Mainnet, Optimism, Arbitrum and Sepolia but can be configured to use Infura, Alchemy or local nodes by setting `ETH_RPC_ENDPOINTS` ENV variable.

- **Code Structure** – The example is split into `main.go`, which loops over endpoints, and helper functions such as `testRPC()` that call various methods and handle errors.

### 4.3 C Application Example

At `examples/c-app` there is a simple app demonstrating how to use the C library.

**usage**

At the root do to create the library:

```bash
make build-c-lib
```

Run the example:

```bash
cd examples/c-app && make build
make
cd bin/
./c-app
```

### 4.4 Multicall3 Example Output

The multicall‑powered flow can fetch thousands of ERC‑20 balances in a single RPC call while pinning all results to the same block:

```
Found 5333 tokens for chain ID 1
Using Multicall3 contract at: 0xcA11bde05977b3631167028862bE2a173976CA11
Prepared 5335 balance calls for multicall

=== Multicall3 Results ===
Block Number: 23275719

ETH (Ethereum): 26.332052548639873000
...
Summary: total tokens queried: 5333, tokens with non-zero balance: 507, success rate: 100%
```

## 5. Testing & Development

### 5.1 Fetching  SDK

Developers can fetch the SDK by running:

```bash
go get github.com/status-im/go-wallet-sdk
```

### 5.2 Running Tests

All packages are fully tested. To run the tests do:

```bash
go test ./...
```

This executes unit tests for the balance fetcher and Ethereum client. The balance fetcher includes a `mock` package to simulate RPC responses. The repository also includes continuous integration workflows (`.github/workflows`) and static analysis configurations (`.golangci.yml`).

### 5.3 Building the C Shared Library

The SDK includes build support for creating C shared libraries that expose core functionality to non-Go applications.

To build the library run:

```bash
make build-c-lib
```

This creates:
- `build/libgowalletsdk.dylib` (macOS) or `build/libgowalletsdk.so` (Linux)
- `build/libgowalletsdk.h` (C header file)

## 6. Limitations & Future Improvements

- 
