
## 1. Overview and Goals

Go Wallet SDK is a modular Go library intended to support the development of multi‑chain cryptocurrency wallets and blockchain applications. The SDK exposes self‑contained packages for common wallet tasks such as fetching account balances across many EVM chains and interacting with Ethereum JSON‑RPC.

### 1.1 Main Repository Components

| Component             | Purpose                                                                                                                                                                                                                                                    |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/balance/fetcher` | High‑performance balance fetcher for EVM‑compatible chains.  The package can fetch native token balances or ERC‑20 balances for multiple addresses in batches using smart fallback strategies. It includes automatic fallback strategies (Multicall3 contract or standard RPC batching) and exposes simple APIs to fetch balances for many addresses or tokens                                                             |
| `pkg/multicall`       | Efficient batching of multiple Ethereum contract calls into single transactions using Multicall3. Supports native ETH, ERC20, ERC721, and ERC1155 balance queries with chunked processing, error handling, and both synchronous and asynchronous execution modes. |
| `pkg/ethclient`       | Chain‑agnostic Ethereum JSON‑RPC client.  It provides two method sets: a drop‑in replacement compatible with go‑ethereum's `ethclient` and a custom implementation that follows the Ethereum JSON‑RPC specification without assuming chain‑specific types. It supports JSON‑RPC methods covering `eth_`, `net_` and `web3_` namespace |
| `pkg/eventfilter`     | Efficient filtering for Ethereum transfer events across ERC20, ERC721, and ERC1155 tokens. Minimizes `eth_getLogs` API calls while capturing all relevant transfers involving specified addresses with optimized query generation and direction-based filtering. |
| `pkg/eventlog`        | Ethereum event log parser for ERC20, ERC721, and ERC1155 events. Automatically detects and parses token events with type-safe access to event data, supporting Transfer, Approval, and other standard token events. |
| `pkg/common`          | Shared types and constants. Such as canonical chain IDs (e.g., Ethereum Mainnet, Optimism, Arbitrum, BSC, Base). Developers use these values when configuring the SDK or examples.                               |
| `pkg/contracts/`      | Solidity contracts and Go bindings for smart contract interactions. Includes Multicall3, ERC20, ERC721, and ERC1155 contracts with deployment addresses for multiple chains. |
| `examples/`           | Demonstrations of SDK usage.  Includes `balance-fetcher-web` (a web interface for batch balance fetching), `ethclient‑usage` (an example that exercises the Ethereum client across multiple RPC endpoints), `multiclient3-usage` (demonstrates multicall functionality), and `eventfilter-example` (shows event filtering and parsing capabilities).                                             |                                                                                                                                                 |

## 2. Architecture

### 2.1 High‑level Structure

Go Wallet SDK follows a modular architecture where each package encapsulates a specific functional domain. There is no central runtime; instead applications import only the packages they need. The SDK currently focuses on EVM‑compatible chains, leaving room for additional chain types in the future. The packages are:
- **Balance Fetcher** – Provides efficient methods to retrieve account balances (native or ERC‑20) across many addresses and tokens. It abstracts over RPC batch calls and Multicall3 contract calls. Developers supply a minimal RPC client interface (`RPCClient` and optionally `BatchCaller`) and the package returns a map of balances
- **Multicall** – Efficiently batches multiple Ethereum contract calls into single transactions using Multicall3. Supports native ETH, ERC20, ERC721, and ERC1155 balance queries with automatic chunking, error handling, and both synchronous and asynchronous execution modes. Provides call builders and result processors for different token types.
- **Ethereum Client** – Exposes the full Ethereum JSON‑RPC API. It wraps a standard RPC client and offers two sets of methods: chain‑agnostic versions prefixed with `Eth*` and a drop‑in `BalanceAt`, `BlockNumber` etc. that mirror go‑ethereum's ethclient. The client covers methods including network info, block and transaction queries, account state, contract code and gas estimation
- **Event Filter** – Efficiently filters Ethereum transfer events across ERC20, ERC721, and ERC1155 tokens. Minimizes `eth_getLogs` API calls through optimized query generation and supports direction-based filtering (send, receive, or both). Uses intelligent query merging to reduce the number of RPC calls required.
- **Event Log Parser** – Automatically detects and parses Ethereum event logs for ERC20, ERC721, and ERC1155 tokens. Provides type-safe access to event data with support for Transfer, Approval, and other standard token events. Works seamlessly with the Event Filter package.
- **Common Utilities** – Houses shared types (e.g., `ChainID`) and enumerated constants for well‑known networks. This allows examples and client code to refer to network IDs without hard‑coding numbers.
- **Contract Bindings** – Provides Go bindings for smart contracts including Multicall3, ERC20, ERC721, and ERC1155. Includes deployment addresses for multiple chains and utilities for contract interaction.

The SDK emphasises chain agnosticism: methods do not assume particular transaction formats or gas pricing models and therefore work with Ethereum, L2 networks (Optimism, Arbitrum, Polygon), and other EVM‑compatible chains. Each package hides chain‑specific details behind simple interfaces.

### 2.2 Balance Fetcher Design

The balance fetcher is designed to efficiently query balances for many addresses and tokens. Its design includes:

- **Dual fetch strategies** – The package first attempts to use Multicall3 contract calls to retrieve multiple balances in a single transaction. If Multicall3 is unavailable on a given chain or the call fails, it falls back to batch RPC calls that iterate through addresses/tokens. Both strategies are exposed transparently through the same API.
- **Batching and concurrency** – When using Multicall3, the fetcher groups requests into batches (configurable `batchSize`) to reduce the number of round‑trips. When falling back to RPC, it also groups requests into batches and processes them in parallel when possible, aggregating results into a map keyed by address and token.
- **Chain‑agnostic** – The logic is unaware of specific chain parameters; it accepts any RPC endpoint and optionally a block number. A `ChainID` from `pkg/common` can be used to label results, but the fetcher does not require it.

### 2.3 Multicall Design

The multicall package is designed to efficiently batch multiple Ethereum contract calls into single transactions using Multicall3. Its design includes:

- **Call Builders** – Provides functions to build calls for different token types (native ETH, ERC20, ERC721, ERC1155). Each builder creates properly encoded call data for the specific contract function.
- **Job Runner System** – Supports both synchronous (`RunSync`) and asynchronous (`ProcessJobRunners`) execution modes. The system automatically chunks large call sets into manageable batches to avoid transaction size limits.
- **Error Handling** – Graceful failure handling with detailed error reporting. Individual call failures don't cause the entire batch to fail, allowing partial results to be processed.
- **Result Processing** – Provides dedicated result processors for each token type that decode the raw return data into appropriate Go types (`*big.Int` for balances).
- **Chain Support** – Works with any EVM-compatible chain that has Multicall3 deployed, with automatic address resolution based on chain ID.

### 2.4 Ethereum Client Design

The Ethereum client package (`pkg/ethclient`) wraps a generic RPC client and exposes two categories of methods:

- **Go‑ethereum‑compatible methods** – Methods such as `BlockNumber`, `BalanceAt` and `TransactionByHash` mimic the ethclient interface from go‑ethereum so existing applications can switch to this SDK with minimal changes. These methods require a go‑ethereum RPC client (because they call underlying types) and may not work on Layer 2 chains that diverge from Ethereum’s API.
- **Chain‑agnostic methods** – Methods prefixed with Eth* correspond directly to Ethereum JSON‑RPC calls and accept/return standard Go types. Examples include `EthBlockNumber`, `EthGetBalance`, `EthGasPrice`, `EthGetBlockByNumberWithFullTxs`, `EthGetLogs`, and `EthEstimateGas`. These functions rely only on the JSON‑RPC specification and therefore support any EVM‑compatible chain.

Internally, the client stores a reference to an RPC client and implements each method by calling `rpcClient.CallContext` with the appropriate RPC method name and parameters (see eth.go). It deserialises responses into exported Go types or custom structs (e.g., `BlockWithTxHashes`, `BlockWithFullTxs`). The design includes convenience functions for converting block numbers to RPC arguments and decoding hex‑encoded values.

### 2.5 Common Utilities

The `pkg/common` package defines shared types and enumerations. The main export is `type ChainID uint64` with constants for well‑known networks such as `EthereumMainnet`, `EthereumSepolia`, `OptimismMainnet`, `ArbitrumMainnet`, `BSCMainnet`, `BaseMainnet`, `BaseSepolia` and a custom `StatusNetworkSepolia`. These constants allow the examples to pre‑populate supported chains and label results without repeating numeric IDs.

### 2.6 Event Filter Design

The event filter package (`pkg/eventfilter`) is designed to efficiently query Ethereum transfer events while minimizing API calls:

- **Multi-Token Support** – Supports ERC20, ERC721, and ERC1155 transfer events through a unified interface. Uses standardized event signatures from the `eventlog` package for consistent event detection.
- **Direction-Based Filtering** – Allows filtering by transfer direction (send, receive, or both) by constructing appropriate topic filters. Send direction queries match addresses in the 'from' field, while receive direction queries match addresses in the 'to' field.
- **Query Optimization** – Intelligently merges compatible queries to minimize the number of `eth_getLogs` calls. ERC20 and ERC721 transfers share the same event signature and can be combined in single queries. The package uses OR operations to merge multiple event types when possible.
- **Topic Structure Optimization** – Constructs efficient topic filters by omitting empty trailing topics and using appropriate topic positions for different event types. ERC20/ERC721 transfers use 2-3 topics while ERC1155 transfers use 3-4 topics depending on direction.
- **Chain-Agnostic** – Works with any EVM-compatible chain by using standard event signatures and topic structures. No chain-specific logic or assumptions.

### 2.7 Event Log Parser Design

The event log parser package (`pkg/eventlog`) provides automatic detection and parsing of Ethereum event logs:

- **Multi-Contract Support** – Automatically detects and parses events from ERC20, ERC721, and ERC1155 contracts using a registry-based approach. Each contract type has dedicated parsers that handle their specific event structures.
- **Type-Safe Access** – Provides strongly-typed access to parsed event data through the `Unpacked` field. Each event type is parsed into its corresponding Go struct (e.g., `Erc20Transfer`, `Erc721Transfer`, `Erc1155TransferSingle`).
- **Event Detection** – Uses event signatures and topic patterns to identify event types. Supports all standard token events including Transfer, Approval, ApprovalForAll, and URI events.
- **Integration** – Designed to work seamlessly with the Event Filter package. The `FilterTransfers` function returns parsed events ready for application use.
- **Error Handling** – Gracefully handles unknown or malformed events by returning empty slices. Safe to use with any log data without causing panics.

### 2.8 Contract Bindings

The `pkg/contracts` package provides Go bindings for smart contracts and deployment utilities:

- **Multicall3** – Provides bindings for the Multicall3 contract with deployment addresses for 200+ chains. Includes utilities for address resolution and chain support checking.
- **Token Standards** – ERC20, ERC721, and ERC1155 contract bindings with standard interface implementations.
- **Deployment Management** – Automated deployment address management with utilities to regenerate addresses from official deployment lists.

## 3. API Description

### 3.1 Multicall API (`pkg/multicall`)

The multicall package provides efficient batching of multiple Ethereum contract calls into single transactions using Multicall3.

#### 3.1.1 Call Builders

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `BuildNativeBalanceCall(accountAddress, multicall3Address)` | Builds a call to get native ETH balance | `accountAddress`: `common.Address`, `multicall3Address`: `common.Address` | `multicall3.IMulticall3Call` |
| `BuildERC20BalanceCall(accountAddress, tokenAddress)` | Builds a call to get ERC20 token balance | `accountAddress`: `common.Address`, `tokenAddress`: `common.Address` | `multicall3.IMulticall3Call` |
| `BuildERC721BalanceCall(accountAddress, tokenAddress)` | Builds a call to get ERC721 NFT balance | `accountAddress`: `common.Address`, `tokenAddress`: `common.Address` | `multicall3.IMulticall3Call` |
| `BuildERC1155BalanceCall(accountAddress, tokenAddress, tokenID)` | Builds a call to get ERC1155 token balance | `accountAddress`: `common.Address`, `tokenAddress`: `common.Address`, `tokenID`: `*big.Int` | `multicall3.IMulticall3Call` |

#### 3.1.2 Execution Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `RunSync(ctx, jobs, atBlock, caller, batchSize)` | Execute calls synchronously | `ctx`: `context.Context`, `jobs`: `[][]multicall3.IMulticall3Call`, `atBlock`: `*big.Int`, `caller`: `Caller`, `batchSize`: `int` | `[]JobResult` |
| `ProcessJobRunners(ctx, jobRunners, atBlock, caller, batchSize)` | Execute calls asynchronously | `ctx`: `context.Context`, `jobRunners`: `[]JobRunner`, `atBlock`: `*big.Int`, `caller`: `Caller`, `batchSize`: `int` | `void` |

#### 3.1.3 Result Processing

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `ProcessNativeBalanceResult(result)` | Parse ETH balance from result | `result`: `multicall3.IMulticall3Result` | `*big.Int`, `error` |
| `ProcessERC20BalanceResult(result)` | Parse ERC20 balance from result | `result`: `multicall3.IMulticall3Result` | `*big.Int`, `error` |
| `ProcessERC721BalanceResult(result)` | Parse ERC721 balance from result | `result`: `multicall3.IMulticall3Result` | `*big.Int`, `error` |
| `ProcessERC1155BalanceResult(result)` | Parse ERC1155 balance from result | `result`: `multicall3.IMulticall3Result` | `*big.Int`, `error` |

#### 3.1.4 Types

```go
type JobResult struct {
    Results     []multicall3.IMulticall3Result
    Err         error
    BlockNumber *big.Int
    BlockHash   common.Hash
}

type JobRunner struct {
    Job      []multicall3.IMulticall3Call
    ResultCh chan<- JobResult
}

type Caller interface {
    ViewTryBlockAndAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error)
    ViewTryAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error)
}
```

### 3.2 Balance Fetcher API (`pkg/balance/fetcher`)

The balance fetcher exposes two primary functions:

| Function                                                                            | Purpose                                                                                                                                                                                                | Parameters                                                                                                                                                                                                                                              | Returns                                                                                                                                                     |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FetchNativeBalances(ctx, addresses, atBlock, rpcClient, batchSize)`                | Retrieves native token balances (e.g., ETH) for multiple addresses.  The function first tries to use Multicall3 contract calls; if unavailable it sends batched `eth_getBalance` RPC calls.         | `ctx`: context; `addresses`: slice of addresses; `atBlock`: block number or `nil` for latest; `rpcClient`: implements `RPCClient`; `batchSize`: maximum addresses per batch.                                                                            | A map `map[common.Address]*big.Int` associating each address with its balance.  Errors indicate network issues or RPC failures.                             |
| `FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, rpcClient, batchSize)` | Retrieves ERC‑20 token balances for multiple addresses and tokens.  Uses Multicall3 contract calls when available or falls back to batched `eth_call` of `balanceOf` for each (address, token) pair. | `ctx`: context; `addresses`: slice of account addresses; `tokenAddresses`: slice of ERC‑20 contract addresses; `atBlock`: block number or `nil`; `rpcClient`: implements `RPCClient` and `BatchCaller`; `batchSize`: maximum number of calls per batch. | A nested map `map[address]map[token]*big.Int` where `balances[account][token]` is the token balance.  Errors indicate RPC failures or contract call errors. |

More specific functions are also available:

| Function                                                                                                          | Description                                                                                                                                                                                                                  |
| ----------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, multicallCaller, multicallAddress, batchSize)`        | Batches addresses into groups (`batchSize`) and uses Multicall3 to retrieve native balances; decodes results into big.Int balances.                                                                                          |
| `FetchErc20BalancesWithMulticall(ctx, accountAddresses, tokenAddresses, atBlock, multicallCaller, batchSize)`    | Uses Multicall3 to batch ERC20 balance calls for multiple accounts and tokens.                                                                                                       |
| `FetchNativeBalancesStandard(ctx, addresses, atBlock, batchCaller, batchSize)`                                    | Constructs `eth_getBalance` batch requests using the provided `BatchCaller`; decodes hex strings into big.Int balances.                                                                                                      |
| `FetchErc20BalancesStandard(ctx, addresses, tokenAddresses, atBlock, batchCaller, batchSize)`                     | Builds `eth_call` requests for each account/token pair using the ERC‑20 ABI and sends them in batches.                                                                                                                       |

**Multicall3 Deployments**

Multicall3 is deployed on 200+ chains and is the primary method for batching contract calls. The BalanceFetcher uses the following addresses:

| Chain | ChainID | Address |
| ----- | ------- | ------- |
| Ethereum Mainnet | 1 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Optimism Mainnet | 10 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Arbitrum Mainnet | 42161 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Base Mainnet | 8453 | 0xca11bde05977b3631167028862be2a173976ca11 |
| BSC Mainnet | 56 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Polygon Mainnet | 137 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Ethereum Sepolia | 11155111 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Arbitrum Sepolia | 421614 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Optimism Sepolia | 11155420 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Base Sepolia | 84532 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Status Network Sepolia | 1660990954 | 0xca11bde05977b3631167028862be2a173976ca11 |
| BSC Testnet | 97 | 0xca11bde05977b3631167028862be2a173976ca11 |

These are defined at `pkg/contracts/multicall3/deployments.go` and can be accessed via `multicall3.GetMulticall3Address(chainID)`.

**ERC-20 ABI Usage**

- Uses generated bindings in `pkg/contracts/erc20` or packs `balanceOf(address)` via the ABI for `eth_call` in standard mode.
- In standard mode, `balanceOf` is encoded with `abi.Pack("balanceOf", accountAddress)` and sent as `input` to the token contract `to` address.

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

### 3.3 Event Filter API (`pkg/eventfilter`)

The event filter package provides efficient filtering for Ethereum transfer events across ERC20, ERC721, and ERC1155 tokens.

#### 3.3.1 Configuration

| Type | Description | Values |
|------|-------------|---------|
| `TransferType` | Token type to filter | `TransferTypeERC20`, `TransferTypeERC721`, `TransferTypeERC1155` |
| `Direction` | Transfer direction | `Send`, `Receive`, `Both` |
| `TransferQueryConfig` | Main configuration struct | See below |

#### 3.3.2 TransferQueryConfig

```go
type TransferQueryConfig struct {
    FromBlock     *big.Int           // Start block number
    ToBlock       *big.Int           // End block number  
    Accounts      []common.Address   // Addresses to filter for
    TransferTypes []TransferType     // Token types to include
    Direction     Direction          // Transfer direction filter
}
```

#### 3.3.3 Core Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `FilterTransfers(client, config)` | Filter and parse transfer events | `client`: `FilterClient`, `config`: `TransferQueryConfig` | `[]eventlog.Event`, `error` |
| `config.ToFilterQueries()` | Generate optimized filter queries | `config`: `TransferQueryConfig` | `[]ethereum.FilterQuery` |

#### 3.3.4 FilterClient Interface

```go
type FilterClient interface {
    FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
}
```

#### 3.3.5 Query Optimization

The package minimizes API calls through intelligent query merging:

- **Single Transfer Types**: 1-2 queries (Send + Receive)
- **Mixed Transfer Types**: 2-3 queries maximum
- **Event Signature Merging**: Multiple event types in single query using OR operations
- **Topic Structure Optimization**: Merges compatible queries by omitting empty trailing topics

### 3.4 Event Log Parser API (`pkg/eventlog`)

The event log parser package provides automatic detection and parsing of Ethereum event logs.

#### 3.4.1 Core Types

```go
type Event struct {
    ContractKey ContractKey  // "erc20", "erc721", or "erc1155"
    ContractABI *abi.ABI     // Full contract ABI
    EventKey    EventKey     // Specific event type
    ABIEvent    *abi.Event   // ABI event definition
    Unpacked    any          // Type-safe parsed event data
}

type ContractKey string
type EventKey string
```

#### 3.4.2 Core Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `ParseLog(log)` | Parse a single log into events | `log`: `types.Log` | `[]Event` |

#### 3.4.3 Supported Events

| Contract | Event Types | Unpacked Type |
|----------|-------------|---------------|
| ERC20 | Transfer, Approval | `erc20.Erc20Transfer`, `erc20.Erc20Approval` |
| ERC721 | Transfer, Approval, ApprovalForAll | `erc721.Erc721Transfer`, `erc721.Erc721Approval`, `erc721.Erc721ApprovalForAll` |
| ERC1155 | TransferSingle, TransferBatch, ApprovalForAll, URI | `erc1155.Erc1155TransferSingle`, `erc1155.Erc1155TransferBatch`, `erc1155.Erc1155ApprovalForAll`, `erc1155.Erc1155URI` |

#### 3.4.4 Event Signatures

Uses standardized signatures from the `eventlog` package:
- **ERC20/ERC721 Transfer**: `0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef`
- **ERC1155 TransferSingle**: `0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62`
- **ERC1155 TransferBatch**: `0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb`

## 4. Example Applications

### 4.1 Multicall Usage Example

The `examples/multiclient3-usage` folder demonstrates how to use the multicall package for efficient batch contract calls:

```go
package main

import (
    "context"
    "fmt"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/status-im/go-wallet-sdk/pkg/multicall"
    "github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

func main() {
    ctx := context.Background()
    
    // Get Multicall3 address for Ethereum Mainnet
    multicallAddr, exists := multicall3.GetMulticall3Address(1)
    if !exists {
        panic("Multicall3 not available on Ethereum Mainnet")
    }
    
    // Create caller (you would use your RPC client here)
    // caller := multicall3.NewMulticall3Caller(multicallAddr, rpcClient)
    
    // Build calls for multiple accounts and tokens
    accounts := []common.Address{
        common.HexToAddress("0x1234..."),
        common.HexToAddress("0x5678..."),
    }
    
    tokens := []common.Address{
        common.HexToAddress("0xA0b86a33E6441b8C4C8C0C4C0C4C0C4C0C4C0C4C0"), // USDC
        common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), // DAI
    }
    
    // Build native balance calls
    nativeCalls := make([]multicall3.IMulticall3Call, 0, len(accounts))
    for _, account := range accounts {
        nativeCalls = append(nativeCalls, multicall.BuildNativeBalanceCall(account, multicallAddr))
    }
    
    // Build ERC20 balance calls
    tokenCalls := make([]multicall3.IMulticall3Call, 0, len(accounts)*len(tokens))
    for _, account := range accounts {
        for _, token := range tokens {
            tokenCalls = append(tokenCalls, multicall.BuildERC20BalanceCall(account, token))
        }
    }
    
    // Execute calls synchronously
    results := multicall.RunSync(ctx, [][]multicall3.IMulticall3Call{nativeCalls, tokenCalls}, nil, caller, 100)
    
    // Process native balance results
    for i, result := range results[0].Results {
        balance, err := multicall.ProcessNativeBalanceResult(result)
        if err != nil {
            fmt.Printf("Error processing native balance for account %d: %v\n", i, err)
            continue
        }
        fmt.Printf("Account %s native balance: %s wei\n", accounts[i].Hex(), balance.String())
    }
    
    // Process token balance results
    callIndex := 0
    for _, account := range accounts {
        for _, token := range tokens {
            result := results[1].Results[callIndex]
            balance, err := multicall.ProcessERC20BalanceResult(result)
            if err != nil {
                fmt.Printf("Error processing token balance: %v\n", err)
                callIndex++
                continue
            }
            fmt.Printf("Account %s token %s balance: %s\n", account.Hex(), token.Hex(), balance.String())
            callIndex++
        }
    }
}
```

### 4.2 Web‑Based Balance Fetcher

The `examples/balance-fetcher-web` folder contains a complete web application that demonstrates how to use the balance fetcher. Key aspects include:
- **Features** – The web UI allows users to specify custom chains (chain ID and RPC URL), enter multiple Ethereum addresses and an optional block number, then fetch balances across chains. It supports batch fetching for native tokens, automatic fallback to standard RPC, and displays balances in both ETH and wei. The example pre‑populates common chains such as Ethereum, Optimism, Arbitrum and Polygon.
- **Usage** – Running `go run .` in the example directory starts an HTTP server on `localhost:8080`. Users can configure chains, input addresses and click Fetch Balances. The backend sends a `POST /fetch` request containing a JSON payload with chains, addresses and block number.
- **Project Structure** – The example is organised into `main.go` (entry point), `types.go` (data structures), `rpc_client.go` (custom RPC client), `utils.go`, `templates.go` (HTML/JS templates), and `handlers.go` (HTTP handlers).
- **Security Considerations** – The example warns that it is for demonstration only. Production deployments should secure RPC endpoints, implement authentication, validate user input and add rate‑limiting.

### 4.3 Ethereum Client Example

The `examples/ethclient-usage` folder shows how to use the Ethereum client across multiple networks. It exercises a wide range of RPC methods and demonstrates multi‑endpoint support.

- **Features** – The example tests connectivity and functionality across multiple RPC endpoints, retrieves network and blockchain data, account balances and nonces, contract code, filters events, retrieves transaction details, checks network status, and estimates gas. It highlights the chain‑agnostic benefits of the custom **eth.go** methods, which make no assumptions about transaction types or chain‑specific fields.

- **Usage** – Users specify one or more RPC endpoints via the **ETH_RPC_ENDPOINTS** environment variable and run **go run main.go**. The program iterates through each endpoint, prints client and network information, queries blocks and transactions, and demonstrates various API calls. Example output shows block and transaction details, balances, gas prices and event logs.

- **Configuration** – The example includes defaults for Ethereum Mainnet, Optimism, Arbitrum and Sepolia but can be configured to use Infura, Alchemy or local nodes by setting `ETH_RPC_ENDPOINTS` ENV variable.

- **Code Structure** – The example is split into `main.go`, which loops over endpoints, and helper functions such as `testRPC()` that call various methods and handle errors.

### 4.4 Event Filter Example

The `examples/eventfilter-example` folder demonstrates how to use the event filter and event log parser packages to detect and display transfer events for specific accounts.

- **Features** – The example provides a command-line interface with flexible options for filtering transfer events. It supports multi-token filtering (ERC20, ERC721, and ERC1155), direction-based filtering (send, receive, or both), and comprehensive transfer details extraction. The example shows enhanced formatting with shortened addresses and scientific notation for large numbers, raw event metadata including event signatures and log properties, and debug information showing contract keys and unpacked types.

- **Usage** – Users can specify an account address, block range, and optional RPC endpoint. The example supports filtering by direction and displays detailed information about each transfer event found. Command-line options include `-account` (required), `-start` and `-end` block numbers (required), `-rpc` for custom endpoints, and `-direction` for filtering.

- **Output Format** – The example displays transfers grouped by token type with comprehensive details extracted from the `Unpacked` field. It shows block numbers, transaction hashes, addresses, amounts, token IDs, contract addresses, log indices, event signatures, and other metadata. Raw event data is also displayed for debugging purposes.

- **Integration** – The example demonstrates the seamless integration between the `eventfilter` and `eventlog` packages, showing how to filter events and parse them into type-safe Go structs for application use.

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

## 6. Limitations & Future Improvements

- 
