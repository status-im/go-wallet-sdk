# Accounts Example

This example demonstrates the capabilities of the `pkg/accounts` packages through a web-based interface.

## What it demonstrates

- **Generate Random Seed Phrase**: Create mnemonic phrases with 12, 15, 18, 21, or 24 words
- **Create Keystore Account**: Import an account from a mnemonic phrase
- **Derive Child Accounts**: Derive child accounts from a parent account using BIP44 derivation paths
- **Import/Export JSON Keys**: Import and export encrypted JSON key files
- **Export Private Key**: Export keys in standard keystore format (compatible with `keystore.KeyStore`)
- **Sign Messages**: Sign messages using Ethereum message signing
- **Account Management**: View all accounts, unlock/lock accounts, and see account details

## Run

1. Navigate to the example directory:
   ```bash
   cd examples/accounts
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the server:
   ```bash
   go run .
   ```

4. Open your browser and navigate to:
   ```
   http://localhost:8081
   ```

## Usage

### Generate Random Seed Phrase

1. Select the desired word count (12, 15, 18, 21, or 24)
2. Click "Generate Mnemonic"
3. The generated mnemonic will be displayed and automatically filled into the "Create Account" form

### Create Account from Seed Phrase

1. Enter a mnemonic phrase (or use the generated one)
2. Optionally provide a passphrase for encryption
3. Click "Create Account"
4. The account will be added to your keystore and appear in the accounts list

### Derive Child Account

1. Enter the parent account address
2. Enter a derivation path (e.g., `m/44'/60'/0'/0/0`)
3. Provide the parent account's passphrase
4. Optionally check "Pin derived account" to save it to the keystore
5. Click "Derive Account"

### Import/Export Keys

**Import:**
1. Paste a JSON key file into the text area
2. Enter the current passphrase
3. Optionally provide a new passphrase
4. Click "Import Key"

**Export:**
1. Enter the account address
2. Enter the current passphrase
3. Optionally provide a new passphrase for the exported key
4. Click "Export Key"
5. The JSON key will be displayed in a text area

### Export Private Key (Standard Keystore)

1. Enter the account address
2. Enter the current passphrase
3. Enter a new passphrase for the exported key
4. Click "Export Private Key"
5. The exported key will be in standard keystore format, compatible with the original `keystore.KeyStore` implementation

### Sign Messages

1. Enter the account address
2. Enter the message to sign
3. If the account is locked, provide the passphrase
4. Click "Sign Message"
5. The signature and message hash will be displayed

### Unlock/Lock Account

**Unlock:**
1. Enter the account address
2. Enter the passphrase
3. Set a timeout in seconds (0 for indefinite)
4. Click "Unlock"

**Lock:**
1. Enter the account address
2. Click "Lock"

### View Accounts

- All accounts in the keystore are displayed in the right panel
- Click on an account to view detailed information
- The selected account's address will be automatically filled into relevant forms

## Derivation Path Format

Derivation paths follow the BIP44 standard:
- Format: `m/44'/60'/0'/0/0` or `44'/60'/0'/0/0`
- The `m/` prefix is optional
- Hardened indices are indicated with `'` or `h` (e.g., `44'` or `44h`)
- Example paths:
  - `m/44'/60'/0'/0/0` - First account, first change, first address
  - `m/44'/60'/0'/0/1` - First account, first change, second address
  - `m/44'/60'/1'/0/0` - Second account, first change, first address

## Notes

- The keystore is stored in a temporary directory and will be cleared when the server stops
- Accounts must be unlocked before signing messages (unless providing a passphrase)
- Derived accounts are only saved to the keystore if "Pin derived account" is checked
- All operations use light scrypt parameters for faster performance in this example
