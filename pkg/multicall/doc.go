// Package multicall provides utilities to batch many contract calls via Multicall3.
//
// Typical use cases include querying balances and other read-only contract state
// efficiently by packing many calls into fewer JSON-RPC requests.
package multicall
