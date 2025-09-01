package tokenlists

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"go.uber.org/zap"
)

const (
	tokenKeySeparator = "-"
)

// Token represents a token with cross-chain identification.
type Token struct {
	CrossChainID string             `json:"crossChainId"`
	ChainID      uint64             `json:"chainId"`
	Address      gethcommon.Address `json:"address"`
	Decimals     uint               `json:"decimals"`
	Name         string             `json:"name"`
	Symbol       string             `json:"symbol"`
	LogoURI      string             `json:"logoUri"`

	CustomToken bool `json:"custom"`
}

// TokenKey creates a key from provided chainID and address.
func TokenKey(chainID uint64, addr gethcommon.Address) string {
	return fmt.Sprintf("%d%s%s", chainID, tokenKeySeparator, strings.ToLower(addr.Hex()))
}

// ChainAndAddressFromTokenKey extracts chainID and address from a token key.
func ChainAndAddressFromTokenKey(tokenKey string) (uint64, gethcommon.Address, bool) {
	split := strings.Split(tokenKey, tokenKeySeparator)
	if len(split) != 2 {
		return 0, gethcommon.Address{}, false
	}
	chainID, err := strconv.ParseUint(split[0], 10, 64)
	if err != nil {
		return 0, gethcommon.Address{}, false
	}
	address := gethcommon.HexToAddress(split[1])
	return chainID, address, true
}

func (t *Token) Key() string {
	return TokenKey(t.ChainID, t.Address)
}

func (t *Token) IsNative() bool {
	if (t.Address != gethcommon.Address{}) {
		return false
	}

	if t.ChainID == common.BSCMainnet ||
		t.ChainID == common.BSCTestnet {
		return strings.EqualFold(t.Symbol, BinanceSmartChainNativeSymbol)
	}
	return strings.EqualFold(t.Symbol, EthereumNativeSymbol)
}

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

func (r *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", r.Major, r.Minor, r.Patch)
}

// StandardTokenList represents the TokenLists standard format.
type StandardTokenList struct {
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	Version   struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
		Patch int `json:"patch"`
	} `json:"version"`
	Tags     map[string]interface{} `json:"tags"`
	LogoURI  string                 `json:"logoURI"`
	Keywords []string               `json:"keywords"`
	Tokens   []struct {
		ChainID  uint64 `json:"chainId"`
		Address  string `json:"address"`
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Decimals uint   `json:"decimals"`
		LogoURI  string `json:"logoURI"`
	} `json:"tokens"`
}

// TokenList represents a token list.
type TokenList struct {
	Name             string                 `json:"name"`
	Timestamp        string                 `json:"timestamp"`        // time when the list was last updated
	FetchedTimestamp string                 `json:"fetchedTimestamp"` // time when the list was fetched
	Source           string                 `json:"source"`
	Version          Version                `json:"version"`
	Tags             map[string]interface{} `json:"tags"`
	LogoURI          string                 `json:"logoURI"`
	Keywords         []string               `json:"keywords"`
	Tokens           []*Token               `json:"tokens"`
}

// tokenList represents a token list in the remote list of token lists.
type tokenList struct {
	ID        string `json:"id"`
	SourceURL string `json:"sourceUrl"`
	Schema    string `json:"schema"`
}

// fetchedTokenList represents a fetched token list.
type fetchedTokenList struct {
	tokenList
	Etag     string
	Fetched  time.Time
	JsonData []byte
}

// remoteListOfTokenLists represents the remote list of token lists.
type remoteListOfTokenLists struct {
	Timestamp  string      `json:"timestamp"`
	Version    Version     `json:"version"`
	TokenLists []tokenList `json:"tokenLists"`
}

// TokensList is the public interface for managing token lists.
type TokensList interface {
	Start(ctx context.Context, notifyCh chan struct{}) error
	Stop() error

	LastRefreshTime() (time.Time, error)
	RefreshNow(ctx context.Context) error

	PrivacyModeUpdated(ctx context.Context) error

	UniqueTokens() []*Token
	GetTokenByChainAddress(chainID uint64, addr gethcommon.Address) (*Token, bool)
	GetTokensByChain(chainID uint64) []*Token

	TokenLists() []*TokenList
	TokenList(id string) (*TokenList, bool)
}

// Parser interface for parsing different token list formats.
type Parser interface {
	Parse(raw []byte, sourceURL string, fetchedAt time.Time, supportedChains []uint64) (*TokenList, error)
}

// PrivacyGuard interface for checking privacy mode.
type PrivacyGuard interface {
	IsPrivacyOn() (bool, error)
}

// LastTokenListsUpdateTimeStore interface for storing and retrieving the last token lists update time.
type LastTokenListsUpdateTimeStore interface {
	Get() (time.Time, error)
	Set(time.Time) error
}

type Content struct {
	SourceURL string
	Etag      string
	Data      []byte
	Fetched   time.Time
}

// ContentStore interface for storing and retrieving fetched content.
type ContentStore interface {
	GetEtag(id string) (string, error)
	Get(id string) (Content, error)
	Set(id string, content Content) error
	GetAll() (map[string]Content, error)
}

// CustomTokenStore interface for storing and retrieving custom tokens.
type CustomTokenStore interface {
	GetAll() ([]*Token, error)
}

// Config holds the configuration for TokensList.
type Config struct {
	MainList     []byte
	MainListID   string
	InitialLists map[string][]byte
	Parsers      map[string]Parser

	Chains                []uint64
	CoinGeckoChainsMapper map[string]uint64

	RemoteListOfTokenListsURL string
	AutoRefreshInterval       time.Duration
	AutoRefreshCheckInterval  time.Duration // must be <= AutoRefreshInterval

	logger                        *zap.Logger
	PrivacyGuard                  PrivacyGuard
	LastTokenListsUpdateTimeStore LastTokenListsUpdateTimeStore
	ContentStore                  ContentStore
	CustomTokenStore              CustomTokenStore
}

// state represents the internal state of TokensList.
type state struct {
	tokens     map[string]*Token     // key: "chainID-address"
	tokenLists map[string]*TokenList // key: "tokenListID"
}
