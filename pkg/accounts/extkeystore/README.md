# extkeystore

Extended keystore for Ethereum accounts with BIP32 hierarchical deterministic (HD) wallet support.

## Overview

An enhanced keystore that stores **BIP32 extended keys** instead of just private keys, enabling derivation of child accounts from parent keys. Based on go-ethereum's keystore with modifications to support HD wallets.

## Key Features

- **HD Wallet Support**: Store extended keys (BIP32) for hierarchical account derivation
- **Encrypted Storage**: Keys stored as encrypted JSON files following Web3 Secret Storage specification
- **Child Account Derivation**: Derive child accounts from parent keys using BIP44 derivation paths
- **Import/Export**: Import extended keys or standard private keys; export in both formats
- **Account Management**: Create, unlock, lock, sign, and delete accounts

## Quick Start

```go
import "github.com/status-im/go-wallet-sdk/pkg/accounts/extkeystore"

// Create a new keystore
ks := extkeystore.NewKeyStore("/path/to/keystore", 
    extkeystore.LightScryptN, extkeystore.LightScryptP)

// Import an extended key
account, err := ks.ImportExtendedKey(extKey, "passphrase")

// Derive a child account
childAccount, err := ks.DeriveWithPassphrase(account, 
    accounts.DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000, 0, 0},
    true, "passphrase", "newPassphrase")

// Sign a transaction
signature, err := ks.SignHash(account, hash)
```

## Main Types

- `KeyStore`: Main keystore instance managing accounts
- `Key`: Encrypted key structure containing extended key and address

## Constants

- `LightScryptN`, `LightScryptP`: Fast scrypt parameters for development
- `StandardScryptN`, `StandardScryptP`: Standard scrypt parameters for production

## Notes

This package is derived from [go-ethereum's keystore](https://github.com/ethereum/go-ethereum/tree/master/accounts/keystore), modified to store extended keys instead of private keys.
