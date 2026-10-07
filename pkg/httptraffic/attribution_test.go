package httptraffic

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	appModule = "example.com/app/"
	libModule = "example.com/lib/"
)

var testAttribution = Attribution{
	Modules:     []string{appModule, libModule},
	Plumbing:    []string{appModule + "internal/rpc", libModule + "client."},
	EntryPoints: regexp.MustCompile(`\.\(\*API\)\.`),
	Sources: []SourceRule{
		{Path: "/coins/list", Source: "Token list"},
		{Function: "/balance", Source: "Balances"},
		{Function: "/market.", Source: "Prices"},
	},
	PrivateCallers: []string{appModule + "preview."},
	Unattributed:   "Unknown",
}

func TestAttribution_SkipsPlumbingEntryPointsAndOtherModules(t *testing.T) {
	a := &testAttribution
	for function, want := range map[string]bool{
		appModule + "balance.(*Controller).fetchChain.func1": true,
		libModule + "fetcher.FetchBalances.func2":            true,
		appModule + "market.(*Manager).FetchPrices":          true,
		appModule + "internal/rpc/chain.(*Client).Call":      false,
		libModule + "client.(*HTTPClient).Get":               false,
		appModule + "wallet.(*API).FetchPrices":              false,
		selfPackage + "(*roundTripper).RoundTrip":            false,
		"net/http.(*Client).Do":                              false,
	} {
		require.Equal(t, want, a.IsFeature(function), function)
	}
}

func TestAttribution_WithoutModulesWalksNoStack(t *testing.T) {
	require.Empty(t, (&Attribution{}).callerOf())
	require.Empty(t, (*Attribution)(nil).callerOf())
}

func TestShortFunction(t *testing.T) {
	require.Equal(t, "balance.(*Controller).fetchChain",
		shortFunction(appModule+"balance.(*Controller).fetchChain.func1.gowrap2"))
}

func TestAttribution_NamesSources(t *testing.T) {
	a := &testAttribution
	for _, c := range []struct{ function, path, want string }{
		{appModule + "balance.(*Controller).fetchChain", "/rpc#eth_call", "Balances"},
		{libModule + "balance/fetcher.FetchBalances.func1", "/rpc#eth_call", "Balances"},
		{appModule + "market.(*Manager).FetchPrices", "/v1/coins/list", "Token list"},
		{appModule + "market.(*Manager).FetchPrices", "/v1/simple/price", "Prices"},
		{appModule + "stickers.(*Service).fetch", "/stickers", "stickers"},
		{"", "/coins/list", "Token list"},
		{"", "/x", "Unknown"},
	} {
		require.Equal(t, c.want, a.SourceOf(c.function, c.path), c.function+" "+c.path)
	}
	require.Equal(t, "Other", (*Attribution)(nil).SourceOf("", "/x"), "without attribution nothing is named")
}

func TestAttribution_TagsNameTheirSourceOutright(t *testing.T) {
	require.Equal(t, "Balances", (*Attribution)(nil).classify(tagPrefix+"Balances", "/coins/list"))
}
