// Package ethclient provides an Ethereum JSON-RPC client with two method sets:
//
//   - Chain-agnostic methods matching the JSON-RPC spec (prefixed with Eth*, Net*, Web3*)
//   - A go-ethereum ethclient-compatible surface for easier migrations
//
// The chain-agnostic methods are intended to work across EVM-compatible chains
// (L1 and L2), while the compatible surface mirrors go-ethereum where possible.
package ethclient
