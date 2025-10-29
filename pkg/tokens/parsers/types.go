package parsers

//go:generate mockgen -destination=mock/parser.go . TokenListParser,ListOfTokenListsParser

import (
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

// TokenListParser interface for parsing different token list formats.
type TokenListParser interface {
	// Parse parses raw bytes (e.g. JSON) as a token list and converts to TokenList objects.
	// ID, Source, FetchedTimestamp are set by the caller.
	Parse(raw []byte, supportedChains []uint64) (*types.TokenList, error)
}

// ListOfTokenListsParser interface for parsing the list of token lists.s
type ListOfTokenListsParser interface {
	// Parse parses raw bytes (e.g. JSON) as a list of token lists and converts to ListOfTokenLists objects.
	Parse(raw []byte) (*types.ListOfTokenLists, error)
}
