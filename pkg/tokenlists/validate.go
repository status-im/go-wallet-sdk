package tokenlists

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xeipuuv/gojsonschema"
)

func validateConfig(config *Config) error {
	if config == nil {
		return ErrConfigNotProvided
	}
	if config.logger == nil {
		return ErrLoggerNotProvided
	}
	if config.MainList == nil {
		return ErrMainListNotProvided
	}
	if config.MainListID == "" {
		return ErrMainListIDNotProvided
	}
	_, existsInProvidedParsers := config.Parsers[StatusListID]
	_, existsInDefaultParsers := DefaultParsers[config.MainListID]
	if !existsInProvidedParsers && !existsInDefaultParsers {
		return ErrMainListParserNotFound
	}
	for listID := range config.InitialLists {
		if listID == config.MainListID {
			return ErrMainListIDCannotBeUsedAsInitialListID
		}
		_, existsInProvidedParsers := config.Parsers[listID]
		_, existsInDefaultParsers := DefaultParsers[listID]
		if !existsInProvidedParsers && !existsInDefaultParsers {
			return fmt.Errorf("%w listID: %s", ErrInitialListParserNotFound, listID)
		}
	}

	if len(config.Chains) == 0 {
		return ErrChainsNotProvided
	}
	if config.AutoRefreshCheckInterval > config.AutoRefreshInterval {
		return ErrAutoRefreshCheckIntervalGreaterThanInterval
	}
	if config.PrivacyGuard == nil {
		return ErrPrivacyGuardNotProvided
	}
	if config.LastTokenListsUpdateTimeStore == nil {
		return ErrLastTokenListsUpdateTimeStoreNotProvided
	}
	if config.ContentStore == nil {
		return ErrContentStoreNotProvided
	}
	return nil
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
	if logoURI == "" ||
		strings.HasPrefix(logoURI, "data:") ||
		strings.HasPrefix(logoURI, "ipfs://") ||
		strings.HasPrefix(logoURI, "http://") ||
		strings.HasPrefix(logoURI, "https://") {
		return true
	}

	return false
}

func validateToken(token *Token, allowedChains []uint64) error {
	if token == nil {
		return ErrTokenNotProvided
	}

	if !isChainAllowed(token.ChainID, allowedChains) {
		return fmt.Errorf("%w  chainID: %d", ErrChainNotAllowed, token.ChainID)
	}

	if len(token.Address) != common.AddressLength {
		return fmt.Errorf("%w address length: %d", ErrInvalidAddressLength, len(token.Address))
	}

	if token.Symbol == "" {
		return ErrSymbolCannotBeEmpty
	}

	// even theoretically the limit is 256, in practice we should not let users use more than 18
	if token.Decimals > 18 {
		return fmt.Errorf("%w decimals: %d", ErrDecimalsExceedsMaximum, token.Decimals)
	}

	if !isValidLogoURI(token.LogoURI) {
		return fmt.Errorf("%w logoURI: %s", ErrInvalidLogoURI, token.LogoURI)
	}

	return nil
}

func validateJsonAgainstSchema(jsonData string, schemaLoader gojsonschema.JSONLoader) error {
	docLoader := gojsonschema.NewStringLoader(jsonData)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		return ErrTokenListDoesNotMatchSchema
	}

	return nil
}
