# C example using the Go Wallet SDK shared library

Build steps:
- From repo root: make build-c-lib
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
  - GoWSK_NewClient(const char* rpcURL, char** errOut) -> unsigned long long
  - GoWSK_CloseClient(unsigned long long handle)
  - GoWSK_ChainID(unsigned long long handle, char** errOut) -> char*
  - GoWSK_GetBalance(unsigned long long handle, const char* address, char** errOut) -> char*
  - GoWSK_FreeCString(char* s)
