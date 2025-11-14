# mnemonic

Utilities for generating BIP39 mnemonic phrases and creating extended keys from them.

## Overview

Simple package for working with mnemonic seed phrases (BIP39) to generate deterministic wallets. Provides functions to create random mnemonics and derive extended keys from existing phrases.

## Key Features

- **Random Mnemonic Generation**: Generate cryptographically secure mnemonic phrases
- **Extended Key Creation**: Create BIP32 extended keys from mnemonic phrases
- **BIP39 Support**: Supports optional passphrase (BIP39 seed extension)
- **Multiple Lengths**: Supports 12, 15, 18, 21, and 24 word phrases

## Quick Start

```go
import "github.com/status-im/go-wallet-sdk/pkg/accounts/mnemonic"

// Generate a random 12-word mnemonic
phrase, err := mnemonic.CreateRandomMnemonic(12)
// Output: "abandon abandon abandon ..."

// Create an extended key from a mnemonic
extKey, err := mnemonic.CreateExtendedKeyFromMnemonic(phrase, "")
// With optional passphrase:
extKey, err := mnemonic.CreateExtendedKeyFromMnemonic(phrase, "my passphrase")
```

## Functions

- `CreateRandomMnemonic(length int)`: Generate random mnemonic (12, 15, 18, 21, or 24 words)
- `CreateRandomMnemonicWithDefaultLength()`: Generate 12-word mnemonic
- `CreateExtendedKeyFromMnemonic(phrase, passphrase string)`: Create extended key from mnemonic
- `LengthToEntropyStrength(length int)`: Convert word count to entropy strength

## Usage with extkeystore

```go
// Generate mnemonic and import into keystore
phrase, _ := mnemonic.CreateRandomMnemonic(12)
extKey, _ := mnemonic.CreateExtendedKeyFromMnemonic(phrase, "")
account, _ := keystore.ImportExtendedKey(extKey, "passphrase")
```

## Notes

- Mnemonic phrases must be 12, 15, 18, 21, or 24 words (multiples of 3)
- The optional passphrase follows BIP39 specification for seed extension
- Generated extended keys are BIP32 master keys suitable for HD wallet derivation

