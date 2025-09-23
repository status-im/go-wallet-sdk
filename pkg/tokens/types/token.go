package types

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

const (
	tokenKeySeparator = "-"
)

var (
	ErrTokenChainNotAllowed   = errors.New("token chain not allowed")
	ErrInvalidAddressLength   = errors.New("invalid address length")
	ErrNoSymbol               = errors.New("token has no symbol")
	ErrDecimalsExceedsMaximum = errors.New("decimals exceeds maximum")
	ErrInvalidLogoURI         = errors.New("invalid logo URI")
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

// No token except the native one for the chain should have an empty address.
func (t *Token) IsNative() bool {
	return t.Address == gethcommon.Address{}
}

func isChainAllowed(chainID uint64, allowedChains []uint64) bool {
	for _, allowed := range allowedChains {
		if allowed == chainID {
			return true
		}
	}
	return false
}

func isValidLogoURI(logoURI string) bool {
	if logoURI == "" {
		return true
	}

	_, err := url.Parse(logoURI)
	return err == nil
}

func (t *Token) Validate(allowedChains []uint64) error {
	if len(allowedChains) > 0 && !isChainAllowed(t.ChainID, allowedChains) {
		return ErrTokenChainNotAllowed
	}

	if len(t.Address) != gethcommon.AddressLength {
		return ErrInvalidAddressLength
	}

	if t.Symbol == "" {
		return ErrNoSymbol
	}

	// even theoretically the limit is 256, in practice we should not let users use more than 18
	if t.Decimals > 18 {
		return ErrDecimalsExceedsMaximum
	}

	if !isValidLogoURI(t.LogoURI) {
		return ErrInvalidLogoURI
	}

	return nil
}
