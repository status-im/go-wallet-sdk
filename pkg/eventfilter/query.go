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

	var queries []ethereum.FilterQuery

	// Event signatures for transfer events (using constants from eventlog package)
	erc20_721TransferSig := eventlog.ERC20TransferID // Same signature for ERC20 and ERC721
	erc1155TransferSingleSig := eventlog.ERC1155TransferSingleID
	erc1155TransferBatchSig := eventlog.ERC1155TransferBatchID

	// Convert addresses to topic format (32-byte padded)
	var addressTopics []common.Hash
	for _, addr := range c.Accounts {
		addressTopics = append(addressTopics, common.BytesToHash(addr.Bytes()))
	}

	// Group transfer types by their event signatures for optimization
	hasERC20 := false
	hasERC721 := false
	hasERC1155 := false

	for _, transferType := range c.TransferTypes {
		switch transferType {
		case TransferTypeERC20:
			hasERC20 = true
		case TransferTypeERC721:
			hasERC721 = true
		case TransferTypeERC1155:
			hasERC1155 = true
		}
	}

	// Create optimized queries based on transfer types and direction
	switch c.Direction {
	case Send:
		// Send direction: separate queries for each transfer type
		// - ERC20/ERC721: [eventSignature, address] (2 topics)
		// - ERC1155: [eventSignature, {}, address] (3 topics, omitting empty last topic)

		if hasERC20 || hasERC721 {
			queries = append(queries, ethereum.FilterQuery{
				FromBlock: c.FromBlock,
				ToBlock:   c.ToBlock,
				Topics: [][]common.Hash{
					{erc20_721TransferSig}, // Match Transfer event signature
					addressTopics,          // Match any of our addresses in 'from' field
				},
			})
		}

		if hasERC1155 {
			queries = append(queries, ethereum.FilterQuery{
				FromBlock: c.FromBlock,
				ToBlock:   c.ToBlock,
				Topics: [][]common.Hash{
					{erc1155TransferSingleSig, erc1155TransferBatchSig}, // Match either TransferSingle OR TransferBatch
					{},            // Any operator
					addressTopics, // Match any of our addresses in 'from' field
				},
			})
		}

	case Receive:
		// Receive direction: separate queries for each transfer type
		// - ERC20/ERC721: [eventSignature, {}, address] (3 topics)
		// - ERC1155: [eventSignature, {}, {}, address] (4 topics)

		if hasERC20 || hasERC721 {
			queries = append(queries, ethereum.FilterQuery{
				FromBlock: c.FromBlock,
				ToBlock:   c.ToBlock,
				Topics: [][]common.Hash{
					{erc20_721TransferSig}, // Match Transfer event signature
					{},                     // Any 'from' address
					addressTopics,          // Match any of our addresses in 'to' field
				},
			})
		}

		if hasERC1155 {
			queries = append(queries, ethereum.FilterQuery{
				FromBlock: c.FromBlock,
				ToBlock:   c.ToBlock,
				Topics: [][]common.Hash{
					{erc1155TransferSingleSig, erc1155TransferBatchSig}, // Match either TransferSingle OR TransferBatch
					{},            // Any operator
					{},            // Any 'from' address
					addressTopics, // Match any of our addresses in 'to' field
				},
			})
		}

	case Both:
		// Both directions: optimized with merging where possible
		// - ERC20/ERC721 Send: [eventSignature, address] (2 topics)
		// - Merged ERC20/ERC721 Receive + ERC1155 Send: [eventSignature, {}, address] (3 topics)
		// - ERC1155 Receive: [eventSignature, {}, {}, address] (4 topics)

		// ERC20/ERC721 Send query
		if hasERC20 || hasERC721 {
			queries = append(queries, ethereum.FilterQuery{
				FromBlock: c.FromBlock,
				ToBlock:   c.ToBlock,
				Topics: [][]common.Hash{
					{erc20_721TransferSig}, // Match Transfer event signature
					addressTopics,          // Match any of our addresses in 'from' field
				},
			})
		}

		// Merged ERC20/ERC721 Receive + ERC1155 Send query (only if we have both)
		{
			var eventSignatures []common.Hash
			if hasERC20 || hasERC721 {
				eventSignatures = append(eventSignatures, erc20_721TransferSig)
			}
			if hasERC1155 {
				eventSignatures = append(eventSignatures, erc1155TransferSingleSig, erc1155TransferBatchSig)
			}

			queries = append(queries, ethereum.FilterQuery{
				FromBlock: c.FromBlock,
				ToBlock:   c.ToBlock,
				Topics: [][]common.Hash{
					eventSignatures, // Match any of the event signatures
					{},              // Any 'from' address (or operator for ERC1155)
					addressTopics,   // Match any of our addresses in 'to' field
				},
			})
		}

		// ERC1155 Receive query
		if hasERC1155 {
			queries = append(queries, ethereum.FilterQuery{
				FromBlock: c.FromBlock,
				ToBlock:   c.ToBlock,
				Topics: [][]common.Hash{
					{erc1155TransferSingleSig, erc1155TransferBatchSig}, // Match either TransferSingle OR TransferBatch
					{},            // Any operator
					{},            // Any 'from' address
					addressTopics, // Match any of our addresses in 'to' field
				},
			})
		}
	}

	return queries
}
