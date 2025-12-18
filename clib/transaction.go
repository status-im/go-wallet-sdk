package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/core/types"
)

// JSONToRLP converts a transaction from JSON format to RLP-encoded binary (hex string)
// Returns the RLP-encoded transaction as a hex string (with 0x prefix)
//
//export GoWSK_transaction_JSONToRLP
func GoWSK_transaction_JSONToRLP(txJSON *C.char, errOut **C.char) *C.char {
	if txJSON == nil {
		handleError(errOut, errors.New("txJSON is NULL"))
		return nil
	}

	// Unmarshal transaction from JSON
	var tx types.Transaction
	if err := json.Unmarshal([]byte(C.GoString(txJSON)), &tx); err != nil {
		handleError(errOut, err)
		return nil
	}

	// Encode to binary using MarshalBinary
	binaryBytes, err := tx.MarshalBinary()
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	// Return as hex string with 0x prefix
	return C.CString("0x" + hex.EncodeToString(binaryBytes))
}

// RLPToJSON converts a transaction from RLP-encoded binary (hex string) to JSON format
// The RLP hex string can have or not have 0x prefix
// Returns the transaction in JSON format (with raw field included)
//
//export GoWSK_transaction_RLPToJSON
func GoWSK_transaction_RLPToJSON(rlpHex *C.char, errOut **C.char) *C.char {
	if rlpHex == nil {
		handleError(errOut, errors.New("rlpHex is NULL"))
		return nil
	}

	// Parse hex string (remove 0x prefix if present)
	hexStr := C.GoString(rlpHex)
	if len(hexStr) >= 2 && hexStr[0:2] == "0x" {
		hexStr = hexStr[2:]
	}

	// Decode hex to bytes
	binaryBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	// Decode binary to transaction using UnmarshalBinary
	var tx types.Transaction
	if err := tx.UnmarshalBinary(binaryBytes); err != nil {
		handleError(errOut, err)
		return nil
	}

	// Marshal to JSON
	txJSON, err := tx.MarshalJSON()
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	// Parse JSON to add raw binary bytes
	var txMap map[string]interface{}
	if err := json.Unmarshal(txJSON, &txMap); err != nil {
		handleError(errOut, err)
		return nil
	}

	// Add raw binary bytes
	txMap["raw"] = "0x" + hex.EncodeToString(binaryBytes)

	// Marshal back to JSON
	resultJSON, err := json.Marshal(txMap)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(string(resultJSON))
}

// GetHash returns the transaction hash from a JSON-encoded transaction
// Returns the hash as a hex string (with 0x prefix)
//
//export GoWSK_transaction_GetHash
func GoWSK_transaction_GetHash(txJSON *C.char, errOut **C.char) *C.char {
	if txJSON == nil {
		handleError(errOut, errors.New("txJSON is NULL"))
		return nil
	}

	// Unmarshal transaction from JSON
	var tx types.Transaction
	if err := json.Unmarshal([]byte(C.GoString(txJSON)), &tx); err != nil {
		handleError(errOut, err)
		return nil
	}

	// Get hash
	hash := tx.Hash()
	return C.CString(hash.Hex())
}
