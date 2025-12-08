package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"encoding/json"
	"errors"
	"math/big"
	"runtime/cgo"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/status-im/extkeys"

	"github.com/status-im/go-wallet-sdk/pkg/accounts/extkeystore"
)

func castToExtKeyStore(h cgo.Handle) *extkeystore.KeyStore {
	if h == 0 {
		return nil
	}
	ks, ok := h.Value().(*extkeystore.KeyStore)
	if !ok {
		return nil
	}
	return ks
}

//export GoWSK_accounts_extkeystore_NewKeyStore
func GoWSK_accounts_extkeystore_NewKeyStore(keydir *C.char, scryptN, scryptP C.int, errOut **C.char) C.uintptr_t {
	if keydir == nil {
		handleError(errOut, errors.New("keydir is NULL"))
		return 0
	}
	keydirStr := C.GoString(keydir)
	ks := extkeystore.NewKeyStore(keydirStr, int(scryptN), int(scryptP))
	h := cgo.NewHandle(ks)
	return C.uintptr_t(h)
}

//export GoWSK_accounts_extkeystore_CloseKeyStore
func GoWSK_accounts_extkeystore_CloseKeyStore(handle C.uintptr_t) {
	h := cgo.Handle(handle)
	if h == 0 {
		return
	}
	h.Delete()
}

//export GoWSK_accounts_extkeystore_Accounts
func GoWSK_accounts_extkeystore_Accounts(handle C.uintptr_t, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	accs := ks.Accounts()
	accountsJSON := make([]accountJSON, len(accs))
	for i, acc := range accs {
		accountsJSON[i] = accountJSON{
			Address: acc.Address.Hex(),
			URL:     acc.URL.String(),
		}
	}
	jsonBytes, err := json.Marshal(accountsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(jsonBytes))
}

//export GoWSK_accounts_extkeystore_NewAccount
func GoWSK_accounts_extkeystore_NewAccount(handle C.uintptr_t, passphrase *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	account, err := ks.NewAccount(passphraseStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	accJSON := accountJSON{
		Address: account.Address.Hex(),
		URL:     account.URL.String(),
	}
	jsonBytes, err := json.Marshal(accJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(jsonBytes))
}

//export GoWSK_accounts_extkeystore_Import
func GoWSK_accounts_extkeystore_Import(handle C.uintptr_t, keyJSON *C.char, passphrase, newPassphrase *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if keyJSON == nil {
		handleError(errOut, errors.New("keyJSON is NULL"))
		return nil
	}
	keyJSONBytes := []byte(C.GoString(keyJSON))
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	newPassphraseStr := ""
	if newPassphrase != nil {
		newPassphraseStr = C.GoString(newPassphrase)
	}
	account, err := ks.Import(keyJSONBytes, passphraseStr, newPassphraseStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	accJSON := accountJSON{
		Address: account.Address.Hex(),
		URL:     account.URL.String(),
	}
	jsonBytes, err := json.Marshal(accJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(jsonBytes))
}

//export GoWSK_accounts_extkeystore_ImportExtendedKey
func GoWSK_accounts_extkeystore_ImportExtendedKey(handle C.uintptr_t, extKeyStr *C.char, passphrase *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if extKeyStr == nil {
		handleError(errOut, errors.New("extKeyStr is NULL"))
		return nil
	}
	extKey, err := extkeys.NewKeyFromString(C.GoString(extKeyStr))
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	defer zeroExtendedKey(extKey)
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	account, err := ks.ImportExtendedKey(extKey, passphraseStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	accJSON := accountJSON{
		Address: account.Address.Hex(),
		URL:     account.URL.String(),
	}
	jsonBytes, err := json.Marshal(accJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(jsonBytes))
}

//export GoWSK_accounts_extkeystore_ExportExt
func GoWSK_accounts_extkeystore_ExportExt(handle C.uintptr_t, address *C.char, passphrase, newPassphrase *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	newPassphraseStr := ""
	if newPassphrase != nil {
		newPassphraseStr = C.GoString(newPassphrase)
	}
	keyJSON, err := ks.ExportExt(account, passphraseStr, newPassphraseStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(keyJSON))
}

//export GoWSK_accounts_extkeystore_ExportPriv
func GoWSK_accounts_extkeystore_ExportPriv(handle C.uintptr_t, address *C.char, passphrase, newPassphrase *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	newPassphraseStr := ""
	if newPassphrase != nil {
		newPassphraseStr = C.GoString(newPassphrase)
	}
	keyJSON, err := ks.ExportPriv(account, passphraseStr, newPassphraseStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(keyJSON))
}

//export GoWSK_accounts_extkeystore_Delete
func GoWSK_accounts_extkeystore_Delete(handle C.uintptr_t, address *C.char, passphrase *C.char, errOut **C.char) {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	err := ks.Delete(account, passphraseStr)
	if err != nil {
		handleError(errOut, err)
	}
}

//export GoWSK_accounts_extkeystore_HasAddress
func GoWSK_accounts_extkeystore_HasAddress(handle C.uintptr_t, address *C.char, errOut **C.char) C.int {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return 0
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return 0
	}
	addr := common.HexToAddress(C.GoString(address))
	has := ks.HasAddress(addr)
	if has {
		return 1
	}
	return 0
}

//export GoWSK_accounts_extkeystore_Unlock
func GoWSK_accounts_extkeystore_Unlock(handle C.uintptr_t, address *C.char, passphrase *C.char, errOut **C.char) {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	err := ks.Unlock(account, passphraseStr)
	if err != nil {
		handleError(errOut, err)
	}
}

//export GoWSK_accounts_extkeystore_Lock
func GoWSK_accounts_extkeystore_Lock(handle C.uintptr_t, address *C.char, errOut **C.char) {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return
	}
	addr := common.HexToAddress(C.GoString(address))
	err := ks.Lock(addr)
	if err != nil {
		handleError(errOut, err)
	}
}

//export GoWSK_accounts_extkeystore_TimedUnlock
func GoWSK_accounts_extkeystore_TimedUnlock(handle C.uintptr_t, address *C.char, passphrase *C.char, timeoutSeconds C.ulong, errOut **C.char) {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	timeout := time.Duration(timeoutSeconds) * time.Second
	err := ks.TimedUnlock(account, passphraseStr, timeout)
	if err != nil {
		handleError(errOut, err)
	}
}

//export GoWSK_accounts_extkeystore_Update
func GoWSK_accounts_extkeystore_Update(handle C.uintptr_t, address *C.char, passphrase, newPassphrase *C.char, errOut **C.char) {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	newPassphraseStr := ""
	if newPassphrase != nil {
		newPassphraseStr = C.GoString(newPassphrase)
	}
	err := ks.Update(account, passphraseStr, newPassphraseStr)
	if err != nil {
		handleError(errOut, err)
	}
}

//export GoWSK_accounts_extkeystore_SignHash
func GoWSK_accounts_extkeystore_SignHash(handle C.uintptr_t, address *C.char, hashHex *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	if hashHex == nil {
		handleError(errOut, errors.New("hashHex is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	hashBytes, err := hexutil.Decode(C.GoString(hashHex))
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	signature, err := ks.SignHash(account, hashBytes)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(hexutil.Encode(signature))
}

//export GoWSK_accounts_extkeystore_SignHashWithPassphrase
func GoWSK_accounts_extkeystore_SignHashWithPassphrase(handle C.uintptr_t, address *C.char, passphrase *C.char, hashHex *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	if hashHex == nil {
		handleError(errOut, errors.New("hashHex is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	hashBytes, err := hexutil.Decode(C.GoString(hashHex))
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	signature, err := ks.SignHashWithPassphrase(account, passphraseStr, hashBytes)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(hexutil.Encode(signature))
}

//export GoWSK_accounts_extkeystore_SignTx
func GoWSK_accounts_extkeystore_SignTx(handle C.uintptr_t, address *C.char, txJSON *C.char, chainIDHex *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	if txJSON == nil {
		handleError(errOut, errors.New("txJSON is NULL"))
		return nil
	}
	if chainIDHex == nil {
		handleError(errOut, errors.New("chainIDHex is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}

	// Parse chainID (supports both decimal and hex with 0x prefix)
	chainIDStr := C.GoString(chainIDHex)
	chainID, ok := new(big.Int).SetString(chainIDStr, 0)
	if !ok {
		handleError(errOut, errors.New("invalid chainID format"))
		return nil
	}

	// Unmarshal transaction
	var tx types.Transaction
	err := json.Unmarshal([]byte(C.GoString(txJSON)), &tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	// Sign transaction
	signedTx, err := ks.SignTx(account, &tx, chainID)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	// Marshal signed transaction back to JSON
	signedTxJSON, err := json.Marshal(signedTx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(signedTxJSON))
}

//export GoWSK_accounts_extkeystore_SignTxWithPassphrase
func GoWSK_accounts_extkeystore_SignTxWithPassphrase(handle C.uintptr_t, address *C.char, passphrase *C.char, txJSON *C.char, chainIDHex *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	if txJSON == nil {
		handleError(errOut, errors.New("txJSON is NULL"))
		return nil
	}
	if chainIDHex == nil {
		handleError(errOut, errors.New("chainIDHex is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}

	// Parse chainID (supports both decimal and hex with 0x prefix)
	chainIDStr := C.GoString(chainIDHex)
	chainID, ok := new(big.Int).SetString(chainIDStr, 0)
	if !ok {
		handleError(errOut, errors.New("invalid chainID format"))
		return nil
	}

	// Unmarshal transaction
	var tx types.Transaction
	err := json.Unmarshal([]byte(C.GoString(txJSON)), &tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	// Sign transaction
	signedTx, err := ks.SignTxWithPassphrase(account, passphraseStr, &tx, chainID)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	// Marshal signed transaction back to JSON
	signedTxJSON, err := json.Marshal(signedTx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(signedTxJSON))
}

//export GoWSK_accounts_extkeystore_Derive
func GoWSK_accounts_extkeystore_Derive(handle C.uintptr_t, address *C.char, derivationPath *C.char, pin C.int, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	if derivationPath == nil {
		handleError(errOut, errors.New("derivationPath is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	pathStr := C.GoString(derivationPath)
	path, err := accounts.ParseDerivationPath(pathStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	pinBool := pin != 0
	derivedAccount, err := ks.Derive(account, path, pinBool)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	accJSON := accountJSON{
		Address: derivedAccount.Address.Hex(),
		URL:     derivedAccount.URL.String(),
	}
	jsonBytes, err := json.Marshal(accJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(jsonBytes))
}

//export GoWSK_accounts_extkeystore_DeriveWithPassphrase
func GoWSK_accounts_extkeystore_DeriveWithPassphrase(handle C.uintptr_t, address *C.char, derivationPath *C.char, pin C.int, passphrase, newPassphrase *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	if derivationPath == nil {
		handleError(errOut, errors.New("derivationPath is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	pathStr := C.GoString(derivationPath)
	path, err := accounts.ParseDerivationPath(pathStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	pinBool := pin != 0
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	newPassphraseStr := ""
	if newPassphrase != nil {
		newPassphraseStr = C.GoString(newPassphrase)
	}
	derivedAccount, err := ks.DeriveWithPassphrase(account, path, pinBool, passphraseStr, newPassphraseStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	accJSON := accountJSON{
		Address: derivedAccount.Address.Hex(),
		URL:     derivedAccount.URL.String(),
	}
	jsonBytes, err := json.Marshal(accJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(jsonBytes))
}

//export GoWSK_accounts_extkeystore_Find
func GoWSK_accounts_extkeystore_Find(handle C.uintptr_t, address *C.char, url *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	ks := castToExtKeyStore(h)
	if ks == nil {
		handleError(errOut, errors.New("invalid keystore handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	account := accounts.Account{Address: addr}
	if url != nil {
		urlStr := C.GoString(url)
		// Parse URL in format "scheme://path"
		parts := strings.Split(urlStr, "://")
		if len(parts) == 2 && parts[0] != "" {
			account.URL = accounts.URL{
				Scheme: parts[0],
				Path:   parts[1],
			}
		} else if len(parts) == 1 {
			// If no scheme, treat as path only
			account.URL = accounts.URL{
				Path: parts[0],
			}
		}
	}
	foundAccount, err := ks.Find(account)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	accJSON := accountJSON{
		Address: foundAccount.Address.Hex(),
		URL:     foundAccount.URL.String(),
	}
	jsonBytes, err := json.Marshal(accJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(string(jsonBytes))
}
