package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
)

// CustomRPCClient implements the required interfaces for the balance fetcher
type CustomRPCClient struct {
	client    *gethrpc.Client
	ethclient *ethclient.Client
	url       string // Store URL for logging
}

func NewCustomRPCClient(url string) (*CustomRPCClient, error) {
	log.Printf("Creating RPC client for URL: %s", url)

	client, err := gethrpc.Dial(url)
	if err != nil {
		log.Printf("Failed to create RPC client for %s: %v", url, err)
		return nil, err
	}

	log.Printf("Successfully created RPC client for URL: %s", url)
	return &CustomRPCClient{
		client:    client,
		ethclient: ethclient.NewClient(client),
		url:       url,
	}, nil
}

func (c *CustomRPCClient) ChainID(ctx context.Context) (*big.Int, error) {
	log.Printf("RPC Call [%s]: eth_chainId", c.url)
	chainID, err := c.ethclient.ChainID(ctx)
	if err != nil {
		log.Printf("RPC Call [%s]: eth_chainId failed: %v", c.url, err)
	} else {
		log.Printf("RPC Call [%s]: eth_chainId returned: %s", c.url, chainID.String())
	}
	return chainID, err
}

func (c *CustomRPCClient) BatchCallContext(ctx context.Context, batch []gethrpc.BatchElem) error {
	log.Printf("RPC Batch Call [%s]: %d requests", c.url, len(batch))

	// Log each batch element
	for i, elem := range batch {
		log.Printf("RPC Batch Call [%s]: [%d] %s with params: %v", c.url, i, elem.Method, elem.Args)
	}

	err := c.client.BatchCallContext(ctx, batch)
	if err != nil {
		log.Printf("RPC Batch Call [%s]: failed: %v", c.url, err)
	} else {
		log.Printf("RPC Batch Call [%s]: completed successfully", c.url)
	}
	return err
}

// Implement bind.ContractBackend interface methods
func (c *CustomRPCClient) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	blockStr := "latest"
	if blockNumber != nil {
		blockStr = blockNumber.String()
	}
	log.Printf("RPC Call [%s]: eth_getCode contract=%s block=%s", c.url, contract.Hex(), blockStr)

	code, err := c.ethclient.CodeAt(ctx, contract, blockNumber)
	if err != nil {
		log.Printf("RPC Call [%s]: eth_getCode failed: %v", c.url, err)
	} else {
		log.Printf("RPC Call [%s]: eth_getCode returned code length: %d", c.url, len(code))
	}
	return code, err
}

func (c *CustomRPCClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	blockStr := "latest"
	if blockNumber != nil {
		blockStr = blockNumber.String()
	}
	log.Printf("RPC Call [%s]: eth_call to=%s block=%s", c.url, call.To.Hex(), blockStr)

	result, err := c.ethclient.CallContract(ctx, call, blockNumber)
	if err != nil {
		log.Printf("RPC Call [%s]: eth_call failed: %v", c.url, err)
	} else {
		log.Printf("RPC Call [%s]: eth_call returned result length: %d", c.url, len(result))
	}
	return result, err
}
