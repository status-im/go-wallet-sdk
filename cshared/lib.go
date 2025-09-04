package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"sync"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	sdkethclient "github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

var (
	clientsMutex sync.RWMutex
	nextHandle   uint64 = 1
	clients             = map[uint64]*sdkethclient.Client{}
)

func storeClient(c *sdkethclient.Client) uint64 {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	h := nextHandle
	nextHandle++
	clients[h] = c
	return h
}

func getClient(handle uint64) *sdkethclient.Client {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	return clients[handle]
}

func deleteClient(handle uint64) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	delete(clients, handle)
}

//export GoWSK_NewClient
func GoWSK_NewClient(rpcURL *C.char, errOut **C.char) C.ulonglong {
	if rpcURL == nil {
		if errOut != nil {
			*errOut = C.CString("rpcURL is NULL")
		}
		return 0
	}
	url := C.GoString(rpcURL)
	rpcClient, err := gethrpc.Dial(url)
	if err != nil {
		if errOut != nil {
			*errOut = C.CString(err.Error())
		}
		return 0
	}
	client := sdkethclient.NewClient(rpcClient)
	handle := storeClient(client)
	return C.ulonglong(handle)
}

//export GoWSK_CloseClient
func GoWSK_CloseClient(handle C.ulonglong) {
	h := uint64(handle)
	c := getClient(h)
	if c != nil {
		c.Close()
		deleteClient(h)
	}
}

//export GoWSK_ChainID
func GoWSK_ChainID(handle C.ulonglong, errOut **C.char) *C.char {
	c := getClient(uint64(handle))
	if c == nil {
		if errOut != nil {
			*errOut = C.CString("invalid client handle")
		}
		return nil
	}
	id, err := c.EthChainId(context.Background())
	if err != nil {
		if errOut != nil {
			*errOut = C.CString(err.Error())
		}
		return nil
	}
	return C.CString(id.String())
}

//export GoWSK_GetBalance
func GoWSK_GetBalance(handle C.ulonglong, address *C.char, errOut **C.char) *C.char {
	c := getClient(uint64(handle))
	if c == nil {
		if errOut != nil {
			*errOut = C.CString("invalid client handle")
		}
		return nil
	}
	if address == nil {
		if errOut != nil {
			*errOut = C.CString("address is NULL")
		}
		return nil
	}
	addr := common.HexToAddress(C.GoString(address))
	bal, err := c.EthGetBalance(context.Background(), addr, nil)
	if err != nil {
		if errOut != nil {
			*errOut = C.CString(err.Error())
		}
		return nil
	}
	return C.CString(bal.String())
}

// frees C strings returned by GoWSK functions to prevent memory leaks.
//
//export GoWSK_FreeCString
func GoWSK_FreeCString(s *C.char) {
	if s != nil {
		C.free(unsafe.Pointer(s))
	}
}

func main() {}
