package eventlog_test

import (
	_ "embed"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc1155"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc721"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/eventlog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/erc20_approval_tx_receipt.json
var erc20ApprovalTxReceiptJSON string

//go:embed testdata/erc20_transfer_tx_receipt.json
var erc20TransferTxReceiptJSON string

//go:embed testdata/erc721_approval_for_all_tx_receipt.json
var erc721ApprovalForAllTxReceiptJSON string

//go:embed testdata/erc721_approval_tx_receipt.json
var erc721ApprovalTxReceiptJSON string

//go:embed testdata/erc721_transfer_tx_receipt.json
var erc721TransferTxReceiptJSON string

//go:embed testdata/erc1155_approval_for_all_tx_receipt.json
var erc1155ApprovalForAllTxReceiptJSON string

//go:embed testdata/erc1155_transfer_batch_tx_receipt.json
var erc1155TransferBatchTxReceiptJSON string

//go:embed testdata/erc1155_transfer_single_tx_receipt.json
var erc1155TransferSingleTxReceiptJSON string

//go:embed testdata/erc1155_uri_tx_receipt.json
var erc1155UriTxReceiptJSON string

// Helper function to load and parse a transaction receipt from JSON file
func loadTransactionReceipt(receiptJSON string) (*ethclient.Receipt, error) {
	var receipt ethclient.Receipt
	err := json.Unmarshal([]byte(receiptJSON), &receipt)
	if err != nil {
		return nil, err
	}

	return &receipt, nil
}

func TestParseLog_ERC20Approval(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc20ApprovalTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 1)

	log := *receipt.Logs[0]
	events := eventlog.ParseLog(log)

	require.Len(t, events, 1)
	event := events[0]

	// Verify event structure
	assert.Equal(t, eventlog.ERC20, event.ContractKey)
	assert.Equal(t, eventlog.ERC20Approval, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	approval, ok := event.Unpacked.(erc20.Erc20Approval)
	require.True(t, ok, "Expected erc20.Erc20Approval, got %T", event.Unpacked)

	// Verify approval details from the JSON data
	expectedOwner := common.HexToAddress("0x6a3c63d9ac2bbef74cb62fa10a579e43725ba474")
	expectedSpender := common.HexToAddress("0x9cd8a5b91ee80fdee6c0e2832d75a56555b9f37f")
	expectedValue := new(big.Int)
	expectedValue.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16) // max uint256

	assert.Equal(t, expectedOwner, approval.Owner)
	assert.Equal(t, expectedSpender, approval.Spender)
	assert.Equal(t, expectedValue, approval.Value)
	assert.Equal(t, log, approval.Raw)
}

func TestParseLog_ERC20Transfer(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc20TransferTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 1)

	log := *receipt.Logs[0]
	events := eventlog.ParseLog(log)

	require.Len(t, events, 1)
	event := events[0]

	// Verify event structure
	assert.Equal(t, eventlog.ERC20, event.ContractKey)
	assert.Equal(t, eventlog.ERC20Transfer, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	transfer, ok := event.Unpacked.(erc20.Erc20Transfer)
	require.True(t, ok, "Expected erc20.Erc20Transfer, got %T", event.Unpacked)

	// Verify transfer details from the JSON data
	expectedFrom := common.HexToAddress("0x47f21fccc72f6de655827740b9dc9277c89350a7")
	expectedTo := common.HexToAddress("0x4945ce2d1b5bd904cac839b7fdabafd19cab982b")
	expectedValue := big.NewInt(21850000) // 0x14d6790

	assert.Equal(t, expectedFrom, transfer.From)
	assert.Equal(t, expectedTo, transfer.To)
	assert.Equal(t, expectedValue, transfer.Value)
	assert.Equal(t, log, transfer.Raw)
}

func TestParseLog_ERC721ApprovalForAll(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc721ApprovalForAllTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 1)

	log := *receipt.Logs[0]
	events := eventlog.ParseLog(log)

	var event *eventlog.Event
	for i := range events {
		if events[i].ContractKey == eventlog.ERC721 {
			event = &events[i]
			break
		}
	}
	require.NotNil(t, event, "Expected ERC721ApprovalForAll event not found")

	// Verify event structure
	assert.Equal(t, eventlog.ERC721, event.ContractKey)
	assert.Equal(t, eventlog.ERC721ApprovalForAll, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	approvalForAll, ok := event.Unpacked.(erc721.Erc721ApprovalForAll)
	require.True(t, ok, "Expected erc721.Erc721ApprovalForAll, got %T", event.Unpacked)

	// Verify approval for all details from the JSON data
	expectedOwner := common.HexToAddress("0x9e10002b5a242362fcdd689f9709dfe6a08f05f3")
	expectedOperator := common.HexToAddress("0xa21917cd7a91d322ee88dc06cb11a6495f91c906")
	expectedApproved := true

	assert.Equal(t, expectedOwner, approvalForAll.Owner)
	assert.Equal(t, expectedOperator, approvalForAll.Operator)
	assert.Equal(t, expectedApproved, approvalForAll.Approved)
	assert.Equal(t, log, approvalForAll.Raw)
}

func TestParseLog_ERC721Approval(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc721ApprovalTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 1)

	log := *receipt.Logs[0]
	events := eventlog.ParseLog(log)

	require.Len(t, events, 1)
	event := events[0]

	// Verify event structure
	assert.Equal(t, eventlog.ERC721, event.ContractKey)
	assert.Equal(t, eventlog.ERC721Approval, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	approval, ok := event.Unpacked.(erc721.Erc721Approval)
	require.True(t, ok, "Expected erc721.Erc721Approval, got %T", event.Unpacked)

	// Verify approval details from the JSON data
	expectedOwner := common.HexToAddress("0x1f00db89777c0f5e6d8e74014df9970467da69d5")
	expectedApproved := common.HexToAddress("0xe9827e11ed5e27ed7dea736d13e49cbaba0a4555")
	expectedTokenId := big.NewInt(6898) // 0x1af2

	assert.Equal(t, expectedOwner, approval.Owner)
	assert.Equal(t, expectedApproved, approval.Approved)
	assert.Equal(t, expectedTokenId, approval.TokenId)
	assert.Equal(t, log, approval.Raw)
}

func TestParseLog_ERC721Transfer(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc721TransferTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 1)

	log := *receipt.Logs[0]
	events := eventlog.ParseLog(log)

	require.Len(t, events, 1)
	event := events[0]

	// Verify event structure
	assert.Equal(t, eventlog.ERC721, event.ContractKey)
	assert.Equal(t, eventlog.ERC721Transfer, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	transfer, ok := event.Unpacked.(erc721.Erc721Transfer)
	require.True(t, ok, "Expected erc721.Erc721Transfer, got %T", event.Unpacked)

	// Verify transfer details from the JSON data
	expectedFrom := common.HexToAddress("0x8c9a94648c27816aef7ec8b9b1dbb15d4fb9745a")
	expectedTo := common.HexToAddress("0x331d934aa5f0daed685fd465db2cffeaf082da27")
	expectedTokenId := big.NewInt(5090) // 0x13e2

	assert.Equal(t, expectedFrom, transfer.From)
	assert.Equal(t, expectedTo, transfer.To)
	assert.Equal(t, expectedTokenId, transfer.TokenId)
	assert.Equal(t, log, transfer.Raw)
}

func TestParseLog_ERC1155ApprovalForAll(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc1155ApprovalForAllTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 1)

	log := *receipt.Logs[0]
	events := eventlog.ParseLog(log)

	var event *eventlog.Event
	for i := range events {
		if events[i].ContractKey == eventlog.ERC1155 {
			event = &events[i]
			break
		}
	}
	require.NotNil(t, event, "Expected ERC1155ApprovalForAll event not found")

	// Verify event structure
	assert.Equal(t, eventlog.ERC1155, event.ContractKey)
	assert.Equal(t, eventlog.ERC1155ApprovalForAll, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	approvalForAll, ok := event.Unpacked.(erc1155.Erc1155ApprovalForAll)
	require.True(t, ok, "Expected erc1155.Erc1155ApprovalForAll, got %T", event.Unpacked)

	// Verify approval for all details from the JSON data
	expectedAccount := common.HexToAddress("0xf98a39456903cfde206603bdd1e1d4227dbe1243")
	expectedOperator := common.HexToAddress("0x994f997a8c1a7deb15cb33bbbbd17839a1c95e58")
	expectedApproved := false

	assert.Equal(t, expectedAccount, approvalForAll.Account)
	assert.Equal(t, expectedOperator, approvalForAll.Operator)
	assert.Equal(t, expectedApproved, approvalForAll.Approved)
	assert.Equal(t, log, approvalForAll.Raw)
}

func TestParseLog_ERC1155TransferBatch(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc1155TransferBatchTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 1)

	log := *receipt.Logs[0]
	events := eventlog.ParseLog(log)

	require.Len(t, events, 1)
	event := events[0]

	// Verify event structure
	assert.Equal(t, eventlog.ERC1155, event.ContractKey)
	assert.Equal(t, eventlog.ERC1155TransferBatch, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	transferBatch, ok := event.Unpacked.(erc1155.Erc1155TransferBatch)
	require.True(t, ok, "Expected erc1155.Erc1155TransferBatch, got %T", event.Unpacked)

	// Verify transfer batch details from the JSON data
	expectedOperator := common.HexToAddress("0xf161ff39e19f605b2115afaeccbb3a112bbe4004")
	expectedFrom := common.HexToAddress("0xf161ff39e19f605b2115afaeccbb3a112bbe4004")
	expectedTo := common.HexToAddress("0xd4012980ef607f79b839095781a31cb2595461cf")
	expectedIds := []*big.Int{
		big.NewInt(157), // 0x9d
		big.NewInt(3),   // 0x3
		big.NewInt(31),  // 0x1f
		big.NewInt(69),  // 0x45
	}
	expectedValues := []*big.Int{
		big.NewInt(1),
		big.NewInt(1),
		big.NewInt(1),
		big.NewInt(1),
	}

	assert.Equal(t, expectedOperator, transferBatch.Operator)
	assert.Equal(t, expectedFrom, transferBatch.From)
	assert.Equal(t, expectedTo, transferBatch.To)
	assert.Equal(t, len(expectedIds), len(transferBatch.Ids))
	for i, id := range expectedIds {
		assert.Equal(t, id, transferBatch.Ids[i])
	}
	assert.Equal(t, len(expectedValues), len(transferBatch.Values))
	for i, value := range expectedValues {
		assert.Equal(t, value, transferBatch.Values[i])
	}
	assert.Equal(t, log, transferBatch.Raw)
}

func TestParseLog_ERC1155TransferSingle(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc1155TransferSingleTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 1)

	log := *receipt.Logs[0]
	events := eventlog.ParseLog(log)

	require.Len(t, events, 1)
	event := events[0]

	// Verify event structure
	assert.Equal(t, eventlog.ERC1155, event.ContractKey)
	assert.Equal(t, eventlog.ERC1155TransferSingle, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	transfer, ok := event.Unpacked.(erc1155.Erc1155TransferSingle)
	require.True(t, ok, "Expected erc1155.Erc1155TransferSingle, got %T", event.Unpacked)

	// Verify transfer details from the JSON data
	expectedOperator := common.HexToAddress("0xfb59acd8d35a43d59b9c455eb0e759e4959309c8")
	expectedFrom := common.HexToAddress("0xfb59acd8d35a43d59b9c455eb0e759e4959309c8")
	expectedTo := common.HexToAddress("0x4eb859fc83977267fbb6ae1066a51fb4c9086c28")
	expectedId := big.NewInt(1)    // 0x1
	expectedValue := big.NewInt(3) // 0x3

	assert.Equal(t, expectedOperator, transfer.Operator)
	assert.Equal(t, expectedFrom, transfer.From)
	assert.Equal(t, expectedTo, transfer.To)
	assert.Equal(t, expectedId, transfer.Id)
	assert.Equal(t, expectedValue, transfer.Value)
	assert.Equal(t, log, transfer.Raw)
}

func TestParseLog_ERC1155URI(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc1155UriTxReceiptJSON)
	require.NoError(t, err)
	require.Len(t, receipt.Logs, 3)

	log := *receipt.Logs[2]
	events := eventlog.ParseLog(log)

	require.Len(t, events, 1)
	event := events[0]

	// Verify event structure
	assert.Equal(t, eventlog.ERC1155, event.ContractKey)
	assert.Equal(t, eventlog.ERC1155URI, event.EventKey)
	assert.NotNil(t, event.ContractABI)
	assert.NotNil(t, event.ABIEvent)
	assert.NotNil(t, event.Unpacked)

	// Verify unpacked data
	uri, ok := event.Unpacked.(erc1155.Erc1155URI)
	require.True(t, ok, "Expected erc1155.Erc1155URI, got %T", event.Unpacked)

	// Verify URI details from the JSON data
	expectedId := big.NewInt(10045) // 0x273d
	expectedValue := "/ipfs/QmZk5Vb8RDnMHgEMcQYHtRhYRxXgTGZG1u2aBJ4ysqyhWT"

	assert.Equal(t, expectedId, uri.Id)
	assert.Equal(t, expectedValue, uri.Value)
	assert.Equal(t, log, uri.Raw)
}

func TestParseLog_UnknownEvent(t *testing.T) {
	// Create a log with an unknown event signature
	log := types.Log{
		Address:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Topics:      []common.Hash{common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")},
		Data:        []byte{},
		BlockNumber: 12345,
		TxHash:      common.HexToHash("0xabcdef"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0x123456"),
		Index:       0,
		Removed:     false,
	}

	events := eventlog.ParseLog(log)
	assert.Len(t, events, 0, "Expected no events for unknown event signature")
}

func TestParseLog_MalformedData(t *testing.T) {
	// Create a log with ERC20 Transfer signature but malformed data
	log := types.Log{
		Address: common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"), // USDT contract
		Topics: []common.Hash{
			common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),        // Transfer signature
			common.BytesToHash(common.HexToAddress("0x47f21fccc72f6de655827740b9dc9277c89350a7").Bytes()), // from
			common.BytesToHash(common.HexToAddress("0x4945ce2d1b5bd904cac839b7fdabafd19cab982b").Bytes()), // to
		},
		Data:        []byte{0x00, 0x01, 0x02, 0x03}, // Malformed data
		BlockNumber: 12345,
		TxHash:      common.HexToHash("0xabcdef"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0x123456"),
		Index:       0,
		Removed:     false,
	}

	events := eventlog.ParseLog(log)
	assert.Len(t, events, 0, "Expected no events for malformed data")
}

func TestParseLog_EventSignatureMismatch(t *testing.T) {
	// Test with completely unknown event signature
	log := types.Log{
		Address: common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"), // USDT contract
		Topics: []common.Hash{
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),        // Unknown signature
			common.BytesToHash(common.HexToAddress("0x47f21fccc72f6de655827740b9dc9277c89350a7").Bytes()), // from
			common.BytesToHash(common.HexToAddress("0x4945ce2d1b5bd904cac839b7fdabafd19cab982b").Bytes()), // to
		},
		Data:        []byte{},
		BlockNumber: 12345,
		TxHash:      common.HexToHash("0xabcdef"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0x123456"),
		Index:       0,
		Removed:     false,
	}

	events := eventlog.ParseLog(log)
	assert.Len(t, events, 0, "Expected no events for unknown signature")
}
