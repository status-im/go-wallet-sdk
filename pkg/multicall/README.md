# Multicall Package

Efficiently batch multiple Ethereum contract calls into single transactions using Multicall3.

## Quick Start

```go
import "github.com/status-im/go-wallet-sdk/pkg/multicall"

// Build calls
calls := []multicall3.IMulticall3Call{
    multicall.BuildNativeBalanceCall(account, multicall3Addr),
    multicall.BuildERC20BalanceCall(account, tokenAddr),
    multicall.BuildERC721BalanceCall(account, nftAddr),
}

// Execute synchronously
results := multicall.RunSync(ctx, [][]multicall3.IMulticall3Call{calls}, blockNumber, caller, batchSize)

// Process results
balance, err := multicall.ProcessNativeBalanceResult(results[0].Results[0])
```

## Features

- **Batch Execution**: Combine multiple calls into single transactions
- **Token Support**: Native ETH, ERC20, ERC721, ERC1155 balance queries
- **Chunked Processing**: Automatic batching for large call sets
- **Error Handling**: Graceful failure handling with detailed error reporting
- **Async Support**: Both synchronous and asynchronous execution modes

## API

### Call Builders
- `BuildNativeBalanceCall()` - Get ETH balance
- `BuildERC20BalanceCall()` - Get ERC20 token balance  
- `BuildERC721BalanceCall()` - Get ERC721 NFT balance
- `BuildERC1155BalanceCall()` - Get ERC1155 token balance

### Execution
- `RunSync()` - Execute calls synchronously, returns results
- `ProcessJobRunners()` - Used to execute calls asynchronously, get results via channels

### Result Processing
- `ProcessNativeBalanceResult()` - Parse ETH balance from result
- `ProcessERC20BalanceResult()` - Parse ERC20 balance from result
- `ProcessERC721BalanceResult()` - Parse ERC721 balance from result
- `ProcessERC1155BalanceResult()` - Parse ERC1155 balance from result

## Example

```go
// Multiple token balances in one call
nativeCalls := []multicall3.IMulticall3Call{
    multicall.BuildNativeBalanceCall(account1, multicall3Addr),
    multicall.BuildNativeBalanceCall(account2, multicall3Addr),
    multicall.BuildNativeBalanceCall(account3, multicall3Addr),
}
tokenCalls := []multicall3.IMulticall3Call{
    multicall.BuildERC20BalanceCall(account1, usdcAddr),
    multicall.BuildERC20BalanceCall(account1, daiAddr),
    multicall.BuildERC20BalanceCall(account2, usdcAddr),
    multicall.BuildERC20BalanceCall(account2, daiAddr),
    multicall.BuildERC20BalanceCall(account3, usdcAddr),
    multicall.BuildERC20BalanceCall(account3, daiAddr),
}

results := multicall.RunSync(ctx, [][]multicall3.IMulticall3Call{nativeCalls, tokenCalls}, blockNum, caller, 100)

// Process native balances
for _, result := results[0] {
  nativeBalance, err := multicall.ProcessNativeBalanceResult(result)
  // Do something
}

// Process token balances
for _, result := results[1] {
  tokenBalance, err := multicall.ProcessERC20BalanceResult(result)
  // Do something
}

```
