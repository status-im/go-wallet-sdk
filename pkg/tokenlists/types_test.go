package tokenlists

import (
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"github.com/stretchr/testify/assert"
)

func TestTokenKey(t *testing.T) {
	token := &Token{
		ChainID: 1,
		Address: gethcommon.HexToAddress("0x123"),
	}
	assert.Equal(t, "1-0x0000000000000000000000000000000000000123", token.Key())
}

func TestTokenIsNative(t *testing.T) {
	token := &Token{}
	assert.False(t, token.IsNative())

	token = &Token{
		ChainID: 1,
	}
	assert.False(t, token.IsNative())

	token = &Token{
		ChainID: 1,
		Address: gethcommon.HexToAddress("0x123"),
	}
	assert.False(t, token.IsNative())

	token = &Token{
		ChainID: 1,
		Address: gethcommon.Address{},
		Symbol:  "ETH",
	}
	assert.True(t, token.IsNative())

	token = &Token{
		ChainID: 1,
		Address: gethcommon.HexToAddress("0x123"),
		Symbol:  "ETH",
	}
	assert.False(t, token.IsNative())

	token = &Token{
		ChainID: common.BSCMainnet,
		Address: gethcommon.HexToAddress("0x123"),
		Symbol:  "ETH",
	}
	assert.False(t, token.IsNative())

	token = &Token{
		ChainID: common.BSCTestnet,
		Address: gethcommon.HexToAddress("0x123"),
		Symbol:  "ETH",
	}
	assert.False(t, token.IsNative())

	token = &Token{
		ChainID: common.BSCMainnet,
		Address: gethcommon.Address{},
		Symbol:  "BNB",
	}
	assert.True(t, token.IsNative())

	token = &Token{
		ChainID: common.BSCTestnet,
		Address: gethcommon.Address{},
		Symbol:  "BNB",
	}
	assert.True(t, token.IsNative())

	token = &Token{
		ChainID: common.EthereumMainnet,
		Address: gethcommon.HexToAddress("0x123"),
		Symbol:  "BNB",
	}
	assert.False(t, token.IsNative())

	token = &Token{
		ChainID: common.EthereumMainnet,
		Address: gethcommon.Address{},
		Symbol:  "BNB",
	}
	assert.False(t, token.IsNative())

	token = &Token{
		ChainID: common.BSCTestnet,
		Address: gethcommon.Address{},
		Symbol:  "ETH",
	}
	assert.False(t, token.IsNative())
}
