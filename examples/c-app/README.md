# C example using the Go Wallet SDK shared library

Build steps:
- From repo root: make shared-library
- Then: cd examples/c-app && make build

Run the example:
```bash
make
cd bin/
./c-app
```

Notes:
- The build generates build/libgowalletsdk.(so|dylib) and header build/libgowalletsdk.h at the repo root.
- On macOS, the example copies the dylib next to the executable and sets rpath for convenience.

Exported functions:

**Memory Management:**
- `GoWSK_FreeCString(char* s)` - Frees C strings returned by GoWSK functions

**Ethereum Client:**
- `GoWSK_ethclient_NewClient(const char* rpcURL, char** errOut) -> uintptr_t`
- `GoWSK_ethclient_CloseClient(uintptr_t handle)`
- `GoWSK_ethclient_ChainID(uintptr_t handle, char** errOut) -> char*`
- `GoWSK_ethclient_GetBalance(uintptr_t handle, const char* address, char** errOut) -> char*`
- `GoWSK_ethclient_RPCCall(uintptr_t handle, const char* method, const char* params, char** errOut) -> char*`

**Multi-Standard Balance Fetcher:**
- `GoWSK_balance_multistandardfetcher_FetchBalances(uintptr_t ethClientHandle, uintptr_t chainID, uintptr_t batchSize, char* fetchConfigJSON, uintptr_t* cancelHandleOut, char** errOut) -> char*`
- `GoWSK_balance_multistandardfetcher_CancelFetchBalances(uintptr_t cancelHandle)`
- `GoWSK_balance_multistandardfetcher_FreeCancelHandle(uintptr_t cancelHandle)`

**Mnemonic Utilities:**
- `GoWSK_accounts_mnemonic_CreateRandomMnemonic(int length, char** errOut) -> char*`
- `GoWSK_accounts_mnemonic_CreateRandomMnemonicWithDefaultLength(char** errOut) -> char*`
- `GoWSK_accounts_mnemonic_LengthToEntropyStrength(int length, char** errOut) -> uint32_t`

**Key Derivation and Conversion:**
- `GoWSK_accounts_keys_CreateExtKeyFromMnemonic(char* phrase, char* passphrase, char** errOut) -> char*`
- `GoWSK_accounts_keys_DeriveExtKey(char* extKeyStr, char* pathStr, char** errOut) -> char*`
- `GoWSK_accounts_keys_ExtKeyToECDSA(char* extKeyStr, char** errOut) -> char*`
- `GoWSK_accounts_keys_ECDSAToPublicKey(char* privateKeyHex, char** errOut) -> char*`
- `GoWSK_accounts_keys_PublicKeyToAddress(char* publicKeyHex, char** errOut) -> char*`

**Extended Keystore:**
- `GoWSK_accounts_extkeystore_NewKeyStore(char* keydir, int scryptN, int scryptP, char** errOut) -> uintptr_t`
- `GoWSK_accounts_extkeystore_CloseKeyStore(uintptr_t handle)`
- `GoWSK_accounts_extkeystore_Accounts(uintptr_t handle, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_NewAccount(uintptr_t handle, char* passphrase, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_Import(uintptr_t handle, char* keyJSON, char* oldPassphrase, char* newPassphrase, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_ImportExtendedKey(uintptr_t handle, char* extKeyStr, char* passphrase, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_ExportExt(uintptr_t handle, char* address, char* passphrase, char* newPassphrase, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_ExportPriv(uintptr_t handle, char* address, char* passphrase, char* newPassphrase, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_Delete(uintptr_t handle, char* address, char* passphrase, char** errOut)`
- `GoWSK_accounts_extkeystore_HasAddress(uintptr_t handle, char* address, char** errOut) -> int`
- `GoWSK_accounts_extkeystore_Unlock(uintptr_t handle, char* address, char* passphrase, char** errOut)`
- `GoWSK_accounts_extkeystore_Lock(uintptr_t handle, char* address, char** errOut)`
- `GoWSK_accounts_extkeystore_TimedUnlock(uintptr_t handle, char* address, char* passphrase, unsigned long timeout, char** errOut)`
- `GoWSK_accounts_extkeystore_Update(uintptr_t handle, char* address, char* oldPassphrase, char* newPassphrase, char** errOut)`
- `GoWSK_accounts_extkeystore_SignHash(uintptr_t handle, char* address, char* hash, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_SignHashWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* hash, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_SignTx(uintptr_t handle, char* address, char* txJSON, char* chainIDHex, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_SignTxWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* txJSON, char* chainIDHex, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_DeriveWithPassphrase(uintptr_t handle, char* address, char* pathStr, int pin, char* passphrase, char* newPassphrase, char** errOut) -> char*`
- `GoWSK_accounts_extkeystore_Find(uintptr_t handle, char* address, char* url, char** errOut) -> char*`

**Standard Keystore:**
- `GoWSK_accounts_keystore_NewKeyStore(char* keydir, int scryptN, int scryptP, char** errOut) -> uintptr_t`
- `GoWSK_accounts_keystore_CloseKeyStore(uintptr_t handle)`
- `GoWSK_accounts_keystore_Accounts(uintptr_t handle, char** errOut) -> char*`
- `GoWSK_accounts_keystore_NewAccount(uintptr_t handle, char* passphrase, char** errOut) -> char*`
- `GoWSK_accounts_keystore_Import(uintptr_t handle, char* keyJSON, char* passphrase, char* newPassphrase, char** errOut) -> char*`
- `GoWSK_accounts_keystore_Export(uintptr_t handle, char* address, char* passphrase, char* newPassphrase, char** errOut) -> char*`
- `GoWSK_accounts_keystore_Delete(uintptr_t handle, char* address, char* passphrase, char** errOut)`
- `GoWSK_accounts_keystore_HasAddress(uintptr_t handle, char* address, char** errOut) -> int`
- `GoWSK_accounts_keystore_Unlock(uintptr_t handle, char* address, char* passphrase, char** errOut)`
- `GoWSK_accounts_keystore_Lock(uintptr_t handle, char* address, char** errOut)`
- `GoWSK_accounts_keystore_TimedUnlock(uintptr_t handle, char* address, char* passphrase, unsigned long timeout, char** errOut)`
- `GoWSK_accounts_keystore_Update(uintptr_t handle, char* address, char* oldPassphrase, char* newPassphrase, char** errOut)`
- `GoWSK_accounts_keystore_SignHash(uintptr_t handle, char* address, char* hash, char** errOut) -> char*`
- `GoWSK_accounts_keystore_SignHashWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* hash, char** errOut) -> char*`
- `GoWSK_accounts_keystore_ImportECDSA(uintptr_t handle, char* privateKeyHex, char* passphrase, char** errOut) -> char*`
- `GoWSK_accounts_keystore_SignTx(uintptr_t handle, char* address, char* txJSON, char* chainIDHex, char** errOut) -> char*`
- `GoWSK_accounts_keystore_SignTxWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* txJSON, char* chainIDHex, char** errOut) -> char*`
- `GoWSK_accounts_keystore_Find(uintptr_t handle, char* address, char* url, char** errOut) -> char*`

For detailed function documentation, see [docs/specs.md](../../docs/specs.md#53-building-the-c-shared-library).
