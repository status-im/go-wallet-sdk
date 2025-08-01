package ethclient

//go:generate mockgen -destination=mock/client.go . RPCClient

import (
	"context"
)

type RPCClient interface {
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
}

type Client struct {
	rpcClient RPCClient
}

// NewClient creates a new Ethereum client
func NewClient(rpcClient RPCClient) *Client {
	return &Client{
		rpcClient: rpcClient,
	}
}
