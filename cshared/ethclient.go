package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"runtime/cgo"

	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	sdkethclient "github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

func castToEthClient(h cgo.Handle) *sdkethclient.Client {
	if h == 0 {
		return nil
	}
	c, ok := h.Value().(*sdkethclient.Client)
	if !ok {
		return nil
	}
	return c
}

//export GoWSK_ethclient_NewClient
func GoWSK_ethclient_NewClient(rpcURL *C.char, errOut **C.char) C.uintptr_t {
	if rpcURL == nil {
		handleError(errOut, errors.New("rpcURL is NULL"))
		return 0
	}
	url := C.GoString(rpcURL)
	rpcClient, err := gethrpc.Dial(url)
	if err != nil {
		handleError(errOut, err)
		return 0
	}
	client := sdkethclient.NewClient(rpcClient)
	h := cgo.NewHandle(client)
	return C.uintptr_t(h)
}

//export GoWSK_ethclient_CloseClient
func GoWSK_ethclient_CloseClient(handle C.uintptr_t) {
	h := cgo.Handle(handle)
	if h == 0 {
		return
	}
	if client := castToEthClient(h); client != nil {
		client.Close()
	}
	h.Delete()
}

//export GoWSK_ethclient_ChainID
func GoWSK_ethclient_ChainID(handle C.uintptr_t, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	c := castToEthClient(h)
	if c == nil {
		handleError(errOut, errors.New("invalid client handle"))
		return nil
	}
	id, err := c.EthChainId(context.Background())
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(id.String())
}

//export GoWSK_ethclient_GetBalance
func GoWSK_ethclient_GetBalance(handle C.uintptr_t, address *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	c := castToEthClient(h)
	if c == nil {
		handleError(errOut, errors.New("invalid client handle"))
		return nil
	}
	if address == nil {
		handleError(errOut, errors.New("address is NULL"))
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	bal, err := c.EthGetBalance(context.Background(), addr, nil)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	return C.CString(bal.String())
}

type rpcRequestJSON struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

//export GoWSK_ethclient_RPCCall
func GoWSK_ethclient_RPCCall(handle C.uintptr_t, requestC *C.char, errOut **C.char) *C.char {
	h := cgo.Handle(handle)
	c := castToEthClient(h)
	if c == nil {
		handleError(errOut, errors.New("invalid client handle"))
		return nil
	}
	if requestC == nil {
		handleError(errOut, errors.New("request is NULL"))
		return nil
	}
	request := rpcRequestJSON{}
	err := json.Unmarshal([]byte(C.GoString(requestC)), &request)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	response := json.RawMessage{}
	err = c.RPCCall(context.Background(), &response, request.Method, request.Params...)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(string(response))
}
