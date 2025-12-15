// Package extkeystore implements an extended Ethereum keystore that stores BIP32
// extended keys to enable HD wallet flows.
//
// It is derived from go-ethereum's keystore, modified to support storing and
// deriving extended keys while still supporting import/export workflows.
package extkeystore
