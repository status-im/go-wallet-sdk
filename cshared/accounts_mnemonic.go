package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"errors"

	"github.com/status-im/go-wallet-sdk/pkg/accounts/mnemonic"
)

//export GoWSK_accounts_mnemonic_CreateRandomMnemonic
func GoWSK_accounts_mnemonic_CreateRandomMnemonic(length C.int, errOut **C.char) *C.char {
	if length <= 0 {
		handleError(errOut, errors.New("length must be positive"))
		return nil
	}
	phrase, err := mnemonic.CreateRandomMnemonic(int(length))
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(phrase)
}

//export GoWSK_accounts_mnemonic_CreateRandomMnemonicWithDefaultLength
func GoWSK_accounts_mnemonic_CreateRandomMnemonicWithDefaultLength(errOut **C.char) *C.char {
	phrase, err := mnemonic.CreateRandomMnemonicWithDefaultLength()
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(phrase)
}

//export GoWSK_accounts_mnemonic_LengthToEntropyStrength
func GoWSK_accounts_mnemonic_LengthToEntropyStrength(length C.int, errOut **C.char) C.uint32_t {
	if length <= 0 {
		handleError(errOut, errors.New("length must be positive"))
		return 0
	}
	entropyStrength, err := mnemonic.LengthToEntropyStrength(int(length))
	if err != nil {
		handleError(errOut, err)
		return 0
	}
	return C.uint32_t(entropyStrength)
}
