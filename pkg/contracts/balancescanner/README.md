# BalanceScanner

`BalanceScanner` is a smart contract which allows fetching multiple balances in a single call. Copied over from https://github.com/MyCryptoHQ/eth-scan.

Go bindings (`balancescanner.go`) autogenerated with
```
solc --abi BalanceScanner.sol -o build
abigen --abi build/BalanceScanner.abi --pkg balancescanner --out balancescanner.go
```

`addresses.go` provides contract address and deployment block number for different chains. 