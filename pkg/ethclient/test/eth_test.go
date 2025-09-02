package ethclient_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	mock_ethclient "github.com/status-im/go-wallet-sdk/pkg/ethclient/mock"
)

//go:embed block_with_details.json
var blockWithDetailsJSON string

//go:embed block_without_details.json
var blockWithoutDetailsJSON string

//go:embed transaction.json
var txJSON string

//go:embed receipt.json
var receiptJSON string

//go:embed block_receipts.json
var blockReceiptsJSON string

//go:embed code.json
var codeJSON string

//go:embed proof.json
var proofJSON string

func TestGethMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	// Test EthBlockNumber
	blockNumberJSON := `"0x1b4"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_blockNumber").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(blockNumberJSON), result)
		})

	blockNumber, err := client.EthBlockNumber(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(436), blockNumber)

	// Test EthChainId
	chainIdJSON := `"0x1"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_chainId").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(chainIdJSON), result)
		})

	chainId, err := client.EthChainId(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1), chainId)

	// Test EthGasPrice
	gasPriceJSON := `"0x4a817c800"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_gasPrice").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(gasPriceJSON), result)
		})

	gasPrice, err := client.EthGasPrice(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(20000000000), gasPrice)

	// Test EthHashrate
	hashrateJSON := `"0x0"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_hashrate").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(hashrateJSON), result)
		})

	hashrate, err := client.EthHashrate(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), hashrate)

	// Test EthMining
	miningJSON := `false`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_mining").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(miningJSON), result)
		})

	mining, err := client.EthMining(context.Background())
	assert.NoError(t, err)
	assert.False(t, mining)

	// Test EthProtocolVersion
	protocolVersionJSON := `"0x3f"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_protocolVersion").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(protocolVersionJSON), result)
		})

	protocolVersion, err := client.EthProtocolVersion(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "0x3f", protocolVersion)

	// Test EthSyncing
	syncingJSON := `false`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_syncing").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(syncingJSON), result)
		})

	syncing, err := client.EthSyncing(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, syncing)

	// Test EthCoinbase
	coinbaseJSON := `"0x3535353535353535353535353535353535353535"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_coinbase").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(coinbaseJSON), result)
		})

	coinbase, err := client.EthCoinbase(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, common.HexToAddress("0x3535353535353535353535353535353535353535"), coinbase)
}

func TestBlockMethodsWithDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	blockHash := common.HexToHash("0xa917fcc721a5465a484e9be17cda0cc5493933dd3bc70c9adbee192cb419c9d7")
	blockNumber := big.NewInt(12911679)

	// Test EthGetBlockByHash
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getBlockByHash", blockHash, true).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(blockWithDetailsJSON), result)
		})

	block, err := client.EthGetBlockByHashWithFullTxs(context.Background(), blockHash)
	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, big.NewInt(12911679), block.Number)
	assert.Equal(t, 204, len(block.Transactions))

	// Test EthGetBlockByNumber
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getBlockByNumber", "0xc5043f", true).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(blockWithDetailsJSON), result)
		})

	block, err = client.EthGetBlockByNumberWithFullTxs(context.Background(), blockNumber)
	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, big.NewInt(12911679), block.Number)
	assert.Equal(t, 204, len(block.Transactions))

	// Test EthGetBlockTransactionCountByHash
	txCountJSON := `"0x1"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getBlockTransactionCountByHash", blockHash).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(txCountJSON), result)
		})

	txCount, err := client.EthGetBlockTransactionCountByHash(context.Background(), blockHash)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), txCount)

	// Test EthGetBlockTransactionCountByNumber
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getBlockTransactionCountByNumber", "0xc5043f").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(txCountJSON), result)
		})

	txCount, err = client.EthGetBlockTransactionCountByNumber(context.Background(), blockNumber)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), txCount)

	// Test EthGetUncleByBlockHashAndIndex
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getUncleByBlockHashAndIndex", blockHash, hexutil.Uint(0)).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(blockWithDetailsJSON), result)
		})

	uncle, err := client.EthGetUncleByBlockHashAndIndex(context.Background(), blockHash, 0)
	assert.NoError(t, err)
	assert.NotNil(t, uncle)

	// Test EthGetUncleByBlockNumberAndIndex
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getUncleByBlockNumberAndIndex", "0xc5043f", hexutil.Uint(0)).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(blockWithDetailsJSON), result)
		})

	uncle, err = client.EthGetUncleByBlockNumberAndIndex(context.Background(), blockNumber, 0)
	assert.NoError(t, err)
	assert.NotNil(t, uncle)

	// Test EthGetUncleCountByBlockHash
	uncleCountJSON := `"0x0"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getUncleCountByBlockHash", blockHash).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(uncleCountJSON), result)
		})

	uncleCount, err := client.EthGetUncleCountByBlockHash(context.Background(), blockHash)
	assert.NoError(t, err)
	assert.Equal(t, hexutil.MustDecodeBig("0x0"), uncleCount)

	// Test EthGetUncleCountByBlockNumber
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getUncleCountByBlockNumber", "0xc5043f").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(uncleCountJSON), result)
		})

	uncleCount, err = client.EthGetUncleCountByBlockNumber(context.Background(), blockNumber)
	assert.NoError(t, err)
	assert.Equal(t, hexutil.MustDecodeBig("0x0"), uncleCount)
}

func TestBlockMethodsWithoutDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	blockHash := common.HexToHash("0xa917fcc721a5465a484e9be17cda0cc5493933dd3bc70c9adbee192cb419c9d7")
	blockNumber := big.NewInt(12911679)

	// Test EthGetBlockByHash
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getBlockByHash", blockHash, false).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(blockWithoutDetailsJSON), result)
		})

	block, err := client.EthGetBlockByHashWithTxHashes(context.Background(), blockHash)
	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, big.NewInt(12911679), block.Number)
	assert.Equal(t, 204, len(block.TransactionHashes))

	// Test EthGetBlockByNumber
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getBlockByNumber", "0xc5043f", false).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(blockWithoutDetailsJSON), result)
		})

	block, err = client.EthGetBlockByNumberWithTxHashes(context.Background(), blockNumber)
	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, big.NewInt(12911679), block.Number)
	assert.Equal(t, 204, len(block.TransactionHashes))
}

func TestTransactionMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	txHash := common.HexToHash("0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b")
	blockHash := common.HexToHash("0xe670ec64341771606e55d6b4ca35a1a6b75ee3d28e5ec2e4b0f1b2b3c4d5e6f7")
	blockNumber := big.NewInt(436)
	address := common.HexToAddress("0xa7d9ddbe1f17865597fbd27ec712455208b6b76d")

	// Test EthGetTransactionByHash
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getTransactionByHash", txHash).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(txJSON), result)
		})

	tx, err := client.EthGetTransactionByHash(context.Background(), txHash)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, txHash, tx.Hash)

	// Test EthGetTransactionByBlockHashAndIndex
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getTransactionByBlockHashAndIndex", blockHash, hexutil.Uint(0)).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(txJSON), result)
		})

	tx, err = client.EthGetTransactionByBlockHashAndIndex(context.Background(), blockHash, 0)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Test EthGetTransactionByBlockNumberAndIndex
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getTransactionByBlockNumberAndIndex", "0x1b4", hexutil.Uint(0)).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(txJSON), result)
		})

	tx, err = client.EthGetTransactionByBlockNumberAndIndex(context.Background(), blockNumber, 0)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Test EthGetTransactionCount
	txCountJSON := `"0x2a"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getTransactionCount", address, "latest").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(txCountJSON), result)
		})

	txCount, err := client.EthGetTransactionCount(context.Background(), address, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(42), txCount)

	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getTransactionReceipt", txHash).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(receiptJSON), result)
		})

	receipt, err := client.EthGetTransactionReceipt(context.Background(), txHash)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, uint64(1), receipt.Status)
}

func TestStateMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	address := common.HexToAddress("0xa7d9ddbe1f17865597fbd27ec712455208b6b76d")
	storageKey := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")

	// Test EthGetBalance
	balanceJSON := `"0xde0b6b3a7640000"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getBalance", address, "latest").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(balanceJSON), result)
		})

	balance, err := client.EthGetBalance(context.Background(), address, nil)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1000000000000000000), balance)

	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getCode", address, "latest").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(codeJSON), result)
		})

	code, err := client.EthGetCode(context.Background(), address, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Test EthGetStorageAt
	storageJSON := `"0x0000000000000000000000000000000000000000000000000000000000000000"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getStorageAt", address, storageKey, "latest").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(storageJSON), result)
		})

	storage, err := client.EthGetStorageAt(context.Background(), address, storageKey, nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, storage)

	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getProof", address, []common.Hash{}, "latest").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(proofJSON), result)
		})

	proof, err := client.EthGetProof(context.Background(), address, []common.Hash{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.Equal(t, address, proof.Address)
	assert.Equal(t, big.NewInt(1000000000000000000), proof.Balance)
	assert.Equal(t, uint64(42), proof.Nonce)
}

func TestMiningMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	// Test EthMining
	miningJSON := `false`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_mining").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(miningJSON), result)
		})

	mining, err := client.EthMining(context.Background())
	assert.NoError(t, err)
	assert.False(t, mining)

	// Test EthHashrate
	hashrateJSON := `"0x0"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_hashrate").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(hashrateJSON), result)
		})

	hashrate, err := client.EthHashrate(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), hashrate)

	// Test EthGetWork
	workJSON := `["0x123", "0x456", "0x789", "0xabc"]`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getWork").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(workJSON), result)
		})

	work, err := client.EthGetWork(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, work)

	// Test EthSubmitHashrate
	submitHashrateJSON := `true`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_submitHashrate", hexutil.Uint64(1000000), common.HexToHash("0x123")).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(submitHashrateJSON), result)
		})

	success, err := client.EthSubmitHashrate(context.Background(), 1000000, common.HexToHash("0x123"))
	assert.NoError(t, err)
	assert.True(t, success)

	// Test EthMaxPriorityFeePerGas
	maxPriorityFeeJSON := `"0x59682f00"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_maxPriorityFeePerGas").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(maxPriorityFeeJSON), result)
		})

	maxPriorityFee, err := client.EthMaxPriorityFeePerGas(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, hexutil.MustDecodeBig("0x59682f00"), maxPriorityFee)
}

func TestTransactionSubmission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	// Test EthSendRawTransaction
	rawTx := []byte{0xd4, 0x6e, 0x8d, 0xd6, 0x7c, 0x5d, 0x32, 0xbe, 0x8d, 0x46, 0xe8, 0xdd, 0x67, 0xc5, 0xd3, 0x2b, 0xe8, 0x05, 0x8b, 0xb8, 0xeb, 0x97, 0x08, 0x70, 0xf0, 0x72, 0x44, 0x56, 0x75, 0x05, 0x8b, 0xb8, 0xeb, 0x97, 0x08, 0x70, 0xf0, 0x72, 0x44, 0x56, 0x75}
	txHashJSON := `"0xe670ec64341771606e55d6b4ca35a1a6b75ee3d28e5ec2e4b0f1b2b3c4d5e6f7"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_sendRawTransaction", hexutil.Bytes(rawTx)).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(txHashJSON), result)
		})

	txHash, err := client.EthSendRawTransaction(context.Background(), rawTx)
	assert.NoError(t, err)
	assert.Equal(t, common.HexToHash("0xe670ec64341771606e55d6b4ca35a1a6b75ee3d28e5ec2e4b0f1b2b3c4d5e6f7"), txHash)
}

func TestSigningMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	address := common.HexToAddress("0xa7d9ddbe1f17865597fbd27ec712455208b6b76d")
	data := []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f} // "hello"

	// Test EthSign
	signatureJSON := `"0x2ac1db7bdf61354f4a6b9b7c3c5c2cef33019bac3249e2c0a2192766d1721c4ba69724e8f69de52f0125ad8b3c5c2cef33019bac3249e2c0a2192766d1721c1b"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_sign", address, hexutil.Bytes(data)).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(signatureJSON), result)
		})

	signature, err := client.EthSign(context.Background(), address, data)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x2a, 0xc1, 0xdb, 0x7b, 0xdf, 0x61, 0x35, 0x4f, 0x4a, 0x6b, 0x9b, 0x7c, 0x3c, 0x5c, 0x2c, 0xef, 0x33, 0x01, 0x9b, 0xac, 0x32, 0x49, 0xe2, 0xc0, 0xa2, 0x19, 0x27, 0x66, 0xd1, 0x72, 0x1c, 0x4b, 0xa6, 0x97, 0x24, 0xe8, 0xf6, 0x9d, 0xe5, 0x2f, 0x01, 0x25, 0xad, 0x8b, 0x3c, 0x5c, 0x2c, 0xef, 0x33, 0x01, 0x9b, 0xac, 0x32, 0x49, 0xe2, 0xc0, 0xa2, 0x19, 0x27, 0x66, 0xd1, 0x72, 0x1c, 0x1b}, signature)

	// Test EthSignTransaction
	to := common.HexToAddress("0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb")
	tx := &ethclient.Transaction{
		From:  address,
		To:    &to,
		Gas:   50000,
		Value: big.NewInt(1000000000000000000),
		Input: data,
	}
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_signTransaction", tx).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(signatureJSON), result)
		})

	signedTx, err := client.EthSignTransaction(context.Background(), tx)
	assert.NoError(t, err)
	assert.NotEmpty(t, signedTx)
}

func TestGasEstimation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	// Test EthEstimateGas
	callMsg := ethereum.CallMsg{
		From:  common.HexToAddress("0xa7d9ddbe1f17865597fbd27ec712455208b6b76d"),
		To:    &common.Address{0xf0, 0x2c, 0x1c, 0x8e, 0x61, 0x14, 0xb1, 0xdb, 0xe8, 0x93, 0x7a, 0x39, 0x26, 0x0b, 0x5b, 0x0a, 0x37, 0x44, 0x32, 0xbb},
		Value: big.NewInt(1000000000000000000),
		Data:  []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}, // "hello"
	}

	gasEstimateJSON := `"0x5208"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_estimateGas", gomock.Any()).
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(gasEstimateJSON), result)
		})

	gasEstimate, err := client.EthEstimateGas(context.Background(), callMsg)
	assert.NoError(t, err)
	assert.Equal(t, uint64(21000), gasEstimate)
}

func TestCallMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	// Test EthCall
	callMsg := ethereum.CallMsg{
		From:  common.HexToAddress("0xa7d9ddbe1f17865597fbd27ec712455208b6b76d"),
		To:    &common.Address{0xf0, 0x2c, 0x1c, 0x8e, 0x61, 0x14, 0xb1, 0xdb, 0xe8, 0x93, 0x7a, 0x39, 0x26, 0x0b, 0x5b, 0x0a, 0x37, 0x44, 0x32, 0xbb},
		Value: big.NewInt(0),
		Data:  []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}, // "hello"
	}

	callResultJSON := `"0x0000000000000000000000000000000000000000000000000000000000000001"`
	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_call", gomock.Any(), "latest").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(callResultJSON), result)
		})

	callResult, err := client.EthCall(context.Background(), callMsg, nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, callResult)
}

func TestBlockReceipts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPC := mock_ethclient.NewMockRPCClient(ctrl)

	client := ethclient.NewClient(mockRPC)

	blockNumber := big.NewInt(436)

	mockRPC.EXPECT().
		CallContext(gomock.Any(), gomock.Any(), "eth_getBlockReceipts", "0x1b4").
		DoAndReturn(func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return json.Unmarshal([]byte(blockReceiptsJSON), result)
		})

	receipts, err := client.EthGetBlockReceipts(context.Background(), blockNumber)
	assert.NoError(t, err)
	assert.Len(t, receipts, 1)
	assert.Equal(t, uint64(1), receipts[0].Status)
}
