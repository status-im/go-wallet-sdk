package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

type RealClient struct {
	rpcClient    *rpc.Client
	ethClient    *ethclient.Client
	infuraClient *infura.Client
}

func NewRealClient(network NetworkInfo, infuraToken string) (*RealClient, error) {
	rpcClient, err := rpc.Dial(network.RPC)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}

	ethClient := ethclient.NewClient(rpcClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create eth client: %w", err)
	}

	infuraClient := infura.NewClient(infuraToken)

	return &RealClient{
		rpcClient:    rpcClient,
		ethClient:    ethClient,
		infuraClient: infuraClient,
	}, nil
}

func (c *RealClient) Close() {
	c.rpcClient.Close()
}

func (c *RealClient) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethclient.FeeHistory, error) {
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
	return c.ethClient.BlockByNumber(ctx, number)
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
