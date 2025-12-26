package builder

import (
	"fmt"
	"strings"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/common"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

const (
	NativeTokenListID = "native"

	EthereumNativeCrossChainID = "eth-native"
	EthereumNativeSymbol       = "ETH"
	EthereumNativeName         = "Ethereum"

	BinanceSmartChainNativeCrossChainID = "bsc-native"
	BinanceSmartChainNativeSymbol       = "BNB"
	BinanceSmartChainNativeName         = "BNB"
)

var (
	ErrEmptyRawTokenList = fmt.Errorf("raw token list data is empty")
	ErrParserIsNil       = fmt.Errorf("parser is nil")
)

// Builder builds token lists into a single list of unique tokens.
type Builder struct {
	chains           []uint64
	tokens           map[string]*types.Token
	tokenLists       map[string]*types.TokenList
	skippedTokenKeys map[string]bool // Set of token keys to skip (for fast lookup)
}

// New creates a new Builder instance.
func New(chains []uint64, skippedTokenKeys []string) *Builder {
	skippedKeysMap := make(map[string]bool)
	for _, key := range skippedTokenKeys {
		skippedKeysMap[strings.ToLower(key)] = true
	}

	return &Builder{
		chains:           chains,
		tokens:           make(map[string]*types.Token),
		tokenLists:       make(map[string]*types.TokenList),
		skippedTokenKeys: skippedKeysMap,
	}
}

// GetTokens returns the list of unique tokens of all added token lists.
func (b *Builder) GetTokens() map[string]*types.Token {
	return b.tokens
}

// GetTokenLists returns the list of added token lists.
func (b *Builder) GetTokenLists() map[string]*types.TokenList {
	return b.tokenLists
}

func getNativeToken(chainID uint64) *types.Token {
	crossChainID := EthereumNativeCrossChainID
	symbol := EthereumNativeSymbol
	name := EthereumNativeName
	logoURI := "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2/logo.png"
	if chainID == common.BSCMainnet || chainID == common.BSCTestnet {
		crossChainID = BinanceSmartChainNativeCrossChainID
		symbol = BinanceSmartChainNativeSymbol
		name = BinanceSmartChainNativeName
		logoURI = "https://assets.coingecko.com/coins/images/825/thumb/bnb-icon2_2x.png?1696501970"
	}
	return &types.Token{
		CrossChainID: crossChainID,
		ChainID:      chainID,
		Symbol:       symbol,
		Name:         name,
		Decimals:     18,
		LogoURI:      logoURI,
	}
}

// AddNativeTokenList adds the native tokens for all chains into a single token list.
func (b *Builder) AddNativeTokenList() error {
	nativeTokenList := &types.TokenList{
		ID:     NativeTokenListID,
		Name:   "Native tokens",
		Tokens: make([]*types.Token, 0),
	}

	for _, chainID := range b.chains {
		nativeToken := getNativeToken(chainID)
		nativeTokenList.Tokens = append(nativeTokenList.Tokens, nativeToken)
	}

	b.AddTokenList(NativeTokenListID, nativeTokenList)
	return nil
}

// AddTokenList adds a token list to the builder and adds the tokens to the list of unique tokens.
// Tokens with keys in the skippedTokenKeys list will be excluded.
func (b *Builder) AddTokenList(tokenListID string, tokenList *types.TokenList) {
	b.tokenLists[tokenListID] = tokenList
	for _, token := range tokenList.Tokens {
		tokenKey := token.Key()
		if b.skippedTokenKeys[tokenKey] {
			continue
		}
		if _, exists := b.tokens[tokenKey]; !exists {
			b.tokens[tokenKey] = token
		}
	}
}

// AddRawTokenList adds a raw token list to the builder using the provided parser and adds the tokens to the list of unique tokens.
func (b *Builder) AddRawTokenList(tokenListID string, raw []byte, sourceURL string, fetchedAt time.Time, parser parsers.TokenListParser) error {
	if len(raw) == 0 {
		return ErrEmptyRawTokenList
	}

	if parser == nil {
		return ErrParserIsNil
	}

	tokenList, err := parser.Parse(raw, b.chains)
	if err != nil {
		return err
	}
	tokenList.ID = tokenListID
	tokenList.Source = sourceURL
	tokenList.FetchedTimestamp = fetchedAt.Format(time.RFC3339)

	b.AddTokenList(tokenListID, tokenList)

	return nil
}
