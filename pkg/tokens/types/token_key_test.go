package types

import (
	"fmt"
	"strings"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestTokenKeyFormat(t *testing.T) {
	for _, chainID := range []uint64{0, 1, 56, 8453, 59144, ^uint64(0)} {
		for _, addr := range []gethcommon.Address{
			{},
			gethcommon.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
			gethcommon.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff"),
		} {
			want := fmt.Sprintf("%d%s%s", chainID, tokenKeySeparator, strings.ToLower(addr.Hex()))
			require.Equal(t, want, TokenKey(chainID, addr))
		}
	}
}

func BenchmarkTokenKey(b *testing.B) {
	addr := gethcommon.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	b.ReportAllocs()
	for b.Loop() {
		_ = TokenKey(8453, addr)
	}
}
