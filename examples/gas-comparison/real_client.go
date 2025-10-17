package main

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

type RealClient struct {
	rpcClient    *rpc.Client
	ethClient    *ethclient.Client
	httpClient   *http.Client
	infuraClient *infura.Client
}

func NewRealClient(network NetworkInfo) (*RealClient, error) {
	rpcClient, err := rpc.Dial(network.RPC)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}

	ethClient := ethclient.NewClient(rpcClient)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	infuraClient := infura.NewClient(httpClient)

	return &RealClient{
		rpcClient:    rpcClient,
		ethClient:    ethClient,
		infuraClient: infuraClient,
	}, nil
}

func (c *RealClient) Close() {
	c.rpcClient.Close()
}

func (c *RealClient) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
	return c.ethClient.FeeHistory(ctx, blockCount, lastBlock, rewardPercentiles)
}

func (c *RealClient) BlockNumber(ctx context.Context) (uint64, error) {
	return c.ethClient.BlockNumber(ctx)
}

func (c *RealClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return c.ethClient.SuggestGasPrice(ctx)
}

func (c *RealClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return c.ethClient.SuggestGasTipCap(ctx)
}

func (c *RealClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethclient.BlockWithFullTxs, error) {
	return c.ethClient.EthGetBlockByNumberWithFullTxs(ctx, number)
}

func (c *RealClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return c.ethClient.EstimateGas(ctx, msg)
}

func (c *RealClient) LineaEstimateGas(ctx context.Context, msg ethereum.CallMsg) (*ethclient.LineaEstimateGasResult, error) {
	return c.ethClient.LineaEstimateGas(ctx, msg)
}

func (c *RealClient) GetGasSuggestions(ctx context.Context, networkID int) (*infura.GasResponse, error) {
	return c.infuraClient.GetGasSuggestions(ctx, networkID)
}
