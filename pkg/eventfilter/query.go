package eventfilter

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/eventlog"
)

type TransferType string

const (
	TransferTypeERC20   TransferType = "erc20"
	TransferTypeERC721  TransferType = "erc721"
	TransferTypeERC1155 TransferType = "erc1155"
)

type Direction string

const (
	Send    Direction = "send"
	Receive Direction = "receive"
	Both    Direction = "both"
)

type TransferQueryConfig struct {
	FromBlock     *big.Int
	ToBlock       *big.Int
	Accounts      []common.Address
	TransferTypes []TransferType
	Direction     Direction
}

func (c *TransferQueryConfig) ToFilterQueries() []ethereum.FilterQuery {
	// FilterQuery should match Transfer events of the given types, with
	// any of the given addresses in the from/to fields

	// Convert addresses to topic format (32-byte padded)
	var addressTopics []common.Hash
	for _, addr := range c.Accounts {
		addressTopics = append(addressTopics, common.BytesToHash(addr.Bytes()))
	}

	// Create optimized queries based on transfer types and direction
	switch c.Direction {
	case Send:
		return buildSendQueries(c.FromBlock, c.ToBlock, addressTopics, c.TransferTypes)
	case Receive:
		return buildReceiveQueries(c.FromBlock, c.ToBlock, addressTopics, c.TransferTypes)
	case Both:
		return buildBothQueries(c.FromBlock, c.ToBlock, addressTopics, c.TransferTypes)
	}
	return nil
}

func buildFilterQuery(fromBlock *big.Int, toBlock *big.Int, topics [][]common.Hash) ethereum.FilterQuery {
	return ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Topics:    topics,
	}
}

func unpackTransferTypes(transferTypes []TransferType) (hasERC20 bool, hasERC721 bool, hasERC1155 bool) {
	for _, transferType := range transferTypes {
		switch transferType {
		case TransferTypeERC20:
			hasERC20 = true
		case TransferTypeERC721:
			hasERC721 = true
		case TransferTypeERC1155:
			hasERC1155 = true
		}
	}
	return
}

func buildSendQueries(fromBlock *big.Int, toBlock *big.Int, addressTopics []common.Hash, transferTypes []TransferType) []ethereum.FilterQuery {
	var queries []ethereum.FilterQuery

	// Send direction: separate queries for each transfer type
	// - ERC20/ERC721: [eventSignature, address] (2 topics)
	// - ERC1155: [eventSignature, {}, address] (3 topics, omitting empty last topic)

	hasERC20, hasERC721, hasERC1155 := unpackTransferTypes(transferTypes)

	if hasERC20 || hasERC721 {
		queries = append(queries, buildFilterQuery(fromBlock, toBlock, [][]common.Hash{
			{eventlog.ERC20TransferID}, // Match Transfer event signature (same for ERC20 and ERC721)
			addressTopics,              // Match any of our addresses in 'from' field
		}))
	}

	if hasERC1155 {
		queries = append(queries, buildFilterQuery(fromBlock, toBlock, [][]common.Hash{
			{eventlog.ERC1155TransferSingleID, eventlog.ERC1155TransferBatchID}, // Match either TransferSingle OR TransferBatch
			{},            // Any operator
			addressTopics, // Match any of our addresses in 'from' field
		}))
	}

	return queries
}

func buildReceiveQueries(fromBlock *big.Int, toBlock *big.Int, addressTopics []common.Hash, transferTypes []TransferType) []ethereum.FilterQuery {
	var queries []ethereum.FilterQuery

	// Receive direction: separate queries for each transfer type
	// - ERC20/ERC721: [eventSignature, {}, address] (3 topics)
	// - ERC1155: [eventSignature, {}, {}, address] (4 topics)

	hasERC20, hasERC721, hasERC1155 := unpackTransferTypes(transferTypes)

	if hasERC20 || hasERC721 {
		queries = append(queries, buildFilterQuery(fromBlock, toBlock, [][]common.Hash{
			{eventlog.ERC20TransferID}, // Match Transfer event signature (same for ERC20 and ERC721)
			{},                         // Any 'from' address
			addressTopics,              // Match any of our addresses in 'to' field
		}))
	}

	if hasERC1155 {
		queries = append(queries, buildFilterQuery(fromBlock, toBlock, [][]common.Hash{
			{eventlog.ERC1155TransferSingleID, eventlog.ERC1155TransferBatchID}, // Match either TransferSingle OR TransferBatch
			{},            // Any operator
			{},            // Any 'from' address
			addressTopics, // Match any of our addresses in 'to' field
		}))
	}

	return queries
}

func buildBothQueries(fromBlock *big.Int, toBlock *big.Int, addressTopics []common.Hash, transferTypes []TransferType) []ethereum.FilterQuery {
	var queries []ethereum.FilterQuery

	// Both direction: optimized with merging where possible
	// - ERC20/ERC721 Send: [eventSignature, address] (2 topics)
	// - Merged ERC20/ERC721 Receive + ERC1155 Send: [eventSignature, {}, address] (3 topics)
	// - ERC1155 Receive: [eventSignature, {}, {}, address] (4 topics)

	hasERC20, hasERC721, hasERC1155 := unpackTransferTypes(transferTypes)

	// ERC20/ERC721 Send query
	if hasERC20 || hasERC721 {
		queries = append(queries, buildFilterQuery(fromBlock, toBlock, [][]common.Hash{
			{eventlog.ERC20TransferID}, // Match Transfer event signature (same for ERC20 and ERC721)
			addressTopics,              // Match any of our addresses in 'from' field
		}))
	}

	// Merged ERC20/ERC721 Receive + ERC1155 Send query (only if we have both)
	{
		var eventSignatures []common.Hash
		if hasERC20 || hasERC721 {
			eventSignatures = append(eventSignatures, eventlog.ERC20TransferID) // Transfer event signature (same for ERC20 and ERC721)
		}
		if hasERC1155 {
			eventSignatures = append(eventSignatures, eventlog.ERC1155TransferSingleID, eventlog.ERC1155TransferBatchID)
		}

		queries = append(queries, buildFilterQuery(fromBlock, toBlock, [][]common.Hash{
			eventSignatures, // Match any of the event signatures
			{},              // Any 'from' address (or operator for ERC1155)
			addressTopics,   // Match any of our addresses in 'to' field
		}))
	}

	// ERC1155 Receive query
	if hasERC1155 {
		queries = append(queries, buildFilterQuery(fromBlock, toBlock, [][]common.Hash{
			{eventlog.ERC1155TransferSingleID, eventlog.ERC1155TransferBatchID}, // Match either TransferSingle OR TransferBatch
			{},            // Any operator
			{},            // Any 'from' address
			addressTopics, // Match any of our addresses in 'to' field
		}))
	}

	return queries
}
