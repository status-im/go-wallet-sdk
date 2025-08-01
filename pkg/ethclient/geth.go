package ethclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// Go-ethereum ethclient compatible methods

// BlockNumber returns the number of most recent block
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	return c.EthBlockNumber(ctx)
}

// ChainID returns the chain ID of the current network
func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	return c.EthChainId(ctx)
}

// SuggestGasPrice returns the current price per gas in wei
func (c *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return c.EthGasPrice(ctx)
}

// BalanceAt returns the balance of the account of given address at the given block number
func (c *Client) BalanceAt(ctx context.Context, address common.Address, blockNumber *big.Int) (*big.Int, error) {
	return c.EthGetBalance(ctx, address, blockNumber)
}

// BlockByHash returns information about a block by hash
func (c *Client) BlockByHash(ctx context.Context, hash common.Hash) (*BlockWithFullTxs, error) {
	return c.EthGetBlockByHashWithFullTxs(ctx, hash)
}

// BlockByNumber returns information about a block by block number
func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*BlockWithFullTxs, error) {
	return c.EthGetBlockByNumberWithFullTxs(ctx, number)
}

// CallContract executes a new message call immediately without creating a transaction on the block chain
func (c *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return c.EthCall(ctx, msg, blockNumber)
}

// EstimateGas generates and returns an estimate of how much gas is necessary to allow the transaction to complete
func (c *Client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return c.EthEstimateGas(ctx, msg)
}

// FeeHistory returns the fee history for the last n blocks
func (c *Client) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*FeeHistory, error) {
	return c.EthFeeHistory(ctx, blockCount, lastBlock, rewardPercentiles)
}

// PendingNonceAt returns the number of transactions sent from an address
func (c *Client) PendingNonceAt(ctx context.Context, address common.Address) (uint64, error) {
	return c.EthGetTransactionCount(ctx, address, nil)
}

// SendRawTransaction creates new message call transaction or a contract creation for signed transactions
func (c *Client) SendRawTransaction(ctx context.Context, encodedTx []byte) (common.Hash, error) {
	return c.EthSendRawTransaction(ctx, encodedTx)
}

// IsConnected checks if the client is connected to the node
func (c *Client) IsConnected(ctx context.Context) bool {
	_, err := c.EthBlockNumber(ctx)
	return err == nil
}

// CodeAt returns code at a given address
func (c *Client) CodeAt(ctx context.Context, address common.Address, blockNumber *big.Int) ([]byte, error) {
	return c.EthGetCode(ctx, address, blockNumber)
}

// StorageAt returns the value from a storage position at a given address
func (c *Client) StorageAt(ctx context.Context, address common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	return c.EthGetStorageAt(ctx, address, key, blockNumber)
}

// TransactionByHash returns the information about a transaction requested by transaction hash
func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (*Transaction, bool, error) {
	tx, err := c.EthGetTransactionByHash(ctx, hash)
	if err != nil {
		return nil, false, err
	}
	if tx == nil {
		return nil, false, ethereum.NotFound
	}
	return tx, tx.BlockNumber == nil, nil
}

// TransactionCount returns the number of transactions in a block from a block matching the given block hash
func (c *Client) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	return c.EthGetBlockTransactionCountByHash(ctx, blockHash)
}

// TransactionInBlock returns a single transaction at index in the given block
func (c *Client) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*Transaction, error) {
	return c.EthGetTransactionByBlockHashAndIndex(ctx, blockHash, index)
}

// TransactionReceipt returns the receipt of a transaction by transaction hash
func (c *Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*Receipt, error) {
	return c.EthGetTransactionReceipt(ctx, txHash)
}

// NonceAt returns the account nonce of the given account
func (c *Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return c.EthGetTransactionCount(ctx, account, blockNumber)
}

// NetworkID returns the network ID
func (c *Client) NetworkID(ctx context.Context) (*big.Int, error) {
	version, err := c.NetVersion(ctx)
	if err != nil {
		return nil, err
	}
	networkID := new(big.Int)
	networkID.SetString(version, 10)
	return networkID, nil
}

// PeerCount returns the number of peers currently connected to the client
func (c *Client) PeerCount(ctx context.Context) (uint64, error) {
	return c.NetPeerCount(ctx)
}

// Additional methods that are not in go-ethereum ethclient but are useful

// GetLatestBlock returns the latest block
func (c *Client) GetLatestBlock(ctx context.Context) (*BlockWithFullTxs, error) {
	return c.EthGetBlockByNumberWithFullTxs(ctx, nil)
}

// GetLatestBlockNumber returns the latest block number
func (c *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	return c.EthBlockNumber(ctx)
}

// GetBalance returns the balance of an address at the latest block
func (c *Client) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return c.EthGetBalance(ctx, address, nil)
}

// GetTransactionByHash returns a transaction by its hash
func (c *Client) GetTransactionByHash(ctx context.Context, hash common.Hash) (*Transaction, error) {
	return c.EthGetTransactionByHash(ctx, hash)
}

// GetTransactionReceipt returns a transaction receipt by transaction hash
func (c *Client) GetTransactionReceipt(ctx context.Context, hash common.Hash) (*Receipt, error) {
	return c.EthGetTransactionReceipt(ctx, hash)
}

// GetCode returns the contract code at an address
func (c *Client) GetCode(ctx context.Context, address common.Address) ([]byte, error) {
	return c.CodeAt(ctx, address, nil)
}

// GetNonce returns the transaction count (nonce) for an address
func (c *Client) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
	return c.EthGetTransactionCount(ctx, address, nil)
}

// GetGasPrice returns the current gas price
func (c *Client) GetGasPrice(ctx context.Context) (*big.Int, error) {
	return c.EthGasPrice(ctx)
}

// CreateFilter creates a new event filter
func (c *Client) CreateFilter(ctx context.Context, query ethereum.FilterQuery) (FilterID, error) {
	return c.EthNewFilter(ctx, query)
}

// GetFilterChanges gets the changes for a filter
func (c *Client) GetFilterChanges(ctx context.Context, filterID FilterID) ([]*Log, error) {
	return c.EthGetFilterChanges(ctx, filterID)
}

// UninstallFilter removes a filter
func (c *Client) UninstallFilter(ctx context.Context, filterID FilterID) (bool, error) {
	return c.EthUninstallFilter(ctx, filterID)
}

// GetChainID returns the chain ID of the network
func (c *Client) GetChainID(ctx context.Context) (*big.Int, error) {
	return c.EthChainId(ctx)
}

// GetNetworkID returns the network ID
func (c *Client) GetNetworkID(ctx context.Context) (string, error) {
	return c.NetVersion(ctx)
}

// GetClientVersion returns the client version
func (c *Client) GetClientVersion(ctx context.Context) (string, error) {
	return c.Web3ClientVersion(ctx)
}

// SuggestGasTipCap returns the current suggested gas tip cap for dynamic fee transactions
func (c *Client) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return c.EthMaxPriorityFeePerGas(ctx)
}

// SyncProgress returns the current sync progress
func (c *Client) SyncProgress(ctx context.Context) (interface{}, error) {
	return c.EthSyncing(ctx)
}

// FilterLogs returns all logs matching the given filter query
func (c *Client) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]*Log, error) {
	// This is equivalent to creating a filter and getting logs
	filterID, err := c.EthNewFilter(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = c.EthUninstallFilter(ctx, filterID)
	}()
	return c.EthGetFilterLogs(ctx, filterID)
}
