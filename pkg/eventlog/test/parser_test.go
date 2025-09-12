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

//go:embed erc20_transfer_tx_receipt.json
var erc20TransferTxReceiptJSON string

//go:embed erc721_transfer_tx_receipt.json
var erc721TransferTxReceiptJSON string

//go:embed erc1155_transfer_tx_receipt.json
var erc1155TransferTxReceiptJSON string

// Helper function to load and parse a transaction receipt from JSON file
func loadTransactionReceipt(receiptJSON string) (*ethclient.Receipt, error) {
	var receipt ethclient.Receipt
	err := json.Unmarshal([]byte(receiptJSON), &receipt)
	if err != nil {
		return nil, err
	}

	return &receipt, nil
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

func TestParseLog_ERC1155TransferSingle(t *testing.T) {
	receipt, err := loadTransactionReceipt(erc1155TransferTxReceiptJSON)
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
