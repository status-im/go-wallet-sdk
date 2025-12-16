// Package fetcher provides high-performance balance fetching for EVM-compatible chains.
//
// It supports:
//   - Native token balances for many accounts
//   - ERC-20 balances for many (account, token) pairs
//   - Fallback strategies (e.g., Multicall3 vs standard/batched RPC)
package fetcher
