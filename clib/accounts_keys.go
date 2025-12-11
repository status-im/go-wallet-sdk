package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/status-im/extkeys"

	"github.com/status-im/go-wallet-sdk/pkg/accounts/mnemonic"
)

//export GoWSK_accounts_keys_CreateExtKeyFromMnemonic
func GoWSK_accounts_keys_CreateExtKeyFromMnemonic(phrase, passphrase *C.char, errOut **C.char) *C.char {
	if phrase == nil {
		handleError(errOut, errors.New("phrase is NULL"))
		return nil
	}
	phraseStr := C.GoString(phrase)
	passphraseStr := ""
	if passphrase != nil {
		passphraseStr = C.GoString(passphrase)
	}
	extKey, err := mnemonic.CreateExtendedKeyFromMnemonic(phraseStr, passphraseStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	defer zeroExtendedKey(extKey)
	// Return the extended key as a human-readable base58-encoded string.
	return C.CString(extKey.String())
}

//export GoWSK_accounts_keys_DeriveExtKey
func GoWSK_accounts_keys_DeriveExtKey(extKeyStrC *C.char, pathStrC *C.char, errOut **C.char) *C.char {
	if extKeyStrC == nil {
		handleError(errOut, errors.New("extKey is NULL"))
		return nil
	}
	extKeyStr := C.GoString(extKeyStrC)
	extKey, err := extkeys.NewKeyFromString(extKeyStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	defer zeroExtendedKey(extKey)
	if pathStrC == nil {
		handleError(errOut, errors.New("path is NULL"))
		return nil
	}
	pathStr := C.GoString(pathStrC)
	path, err := accounts.ParseDerivationPath(pathStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	derivedKey, err := extKey.Derive(path)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	defer zeroExtendedKey(derivedKey)
	// Return the derived key as a human-readable base58-encoded string.
	return C.CString(derivedKey.String())
}

//export GoWSK_accounts_keys_ExtKeyToECDSA
func GoWSK_accounts_keys_ExtKeyToECDSA(extKeyStrC *C.char, errOut **C.char) *C.char {
	if extKeyStrC == nil {
		handleError(errOut, errors.New("extKey is NULL"))
		return nil
	}
	extKeyStr := C.GoString(extKeyStrC)
	extKey, err := extkeys.NewKeyFromString(extKeyStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	defer zeroExtendedKey(extKey)
	privateKeyECDSA := extKey.ToECDSA()
	defer zeroPrivateKey(privateKeyECDSA)
	// Return the private key as a hex-encoded string, always 32 bytes (zero-padded).
	privBytes := privateKeyECDSA.D.Bytes()
	padded := make([]byte, 32)
	copy(padded[32-len(privBytes):], privBytes)
	hexStr := hex.EncodeToString(padded)
	clear(privBytes)
	clear(padded)
	return C.CString(hexStr)
}

//export GoWSK_accounts_keys_ECDSAToPublicKey
func GoWSK_accounts_keys_ECDSAToPublicKey(privateKeyECDSAStrC *C.char, errOut **C.char) *C.char {
	if privateKeyECDSAStrC == nil {
		handleError(errOut, errors.New("privateKeyECDSA is NULL"))
		return nil
	}
	privateKeyECDSAStr := C.GoString(privateKeyECDSAStrC)
	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyECDSAStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	defer zeroPrivateKey(privateKeyECDSA)
	// Return the public key as a hex-encoded string.
	return C.CString(hex.EncodeToString(crypto.FromECDSAPub(&privateKeyECDSA.PublicKey)))
}

//export GoWSK_accounts_keys_PublicKeyToAddress
func GoWSK_accounts_keys_PublicKeyToAddress(publicKeyStrC *C.char, errOut **C.char) *C.char {
	if publicKeyStrC == nil {
		handleError(errOut, errors.New("publicKey is NULL"))
		return nil
	}
	publicKeyStr := C.GoString(publicKeyStrC)
	publicKeyBytes, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	address := crypto.PubkeyToAddress(*publicKey)
	// Return the address as a hex-encoded string.
	return C.CString(address.Hex())
}

// zeroExtendedKey zeroes an extended key in memory.
func zeroExtendedKey(k *extkeys.ExtendedKey) {
	clear(k.KeyData)
}

// zeroPrivateKey zeroes a private key in memory.
func zeroPrivateKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	clear(b)
}
