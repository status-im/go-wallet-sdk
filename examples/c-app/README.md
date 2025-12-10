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
- Exported functions:
  - GoWSK_ethclient_NewClient(const char* rpcURL, char** errOut) -> unsigned long long
  - GoWSK_ethclient_CloseClient(unsigned long long handle)
  - GoWSK_ethclient_ChainID(unsigned long long handle, char** errOut) -> char*
  - GoWSK_ethclient_GetBalance(unsigned long long handle, const char* address, char** errOut) -> char*
  - GoWSK_ethclient_RPCCall(unsigned long long handle, const char* method, const char* params, char** errOut) -> char*
  - GoWSK_FreeCString(char* s)
