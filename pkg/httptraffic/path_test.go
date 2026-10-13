package httptraffic

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecordedPath_MasksSegmentsThatIdentifySomething(t *testing.T) {
	for path, want := range map[string]string{
		"/v3/0123456789abcdef0123456789abcdef":                          "/v3/{id}",
		"/nft/v3/AbCdEfGhIjKlMnOpQrStUvWxYz/getNFTsForOwner":            "/nft/v3/{id}/getNFTsForOwner",
		"/accounts/0xAbC0000000000000000000000000000000000001/balances": "/accounts/{id}/balances",
		"/collections/0xdead/tokens/123456789":                          "/collections/{id}/tokens/{id}",
		"/u/550e8400-e29b-41d4-a716-446655440000":                       "/u/{id}",
		"/ethereum/mainnet":                                             "/ethereum/mainnet",
		"/api/v3/coins/list":                                            "/api/v3/coins/list",
		"/coins/usd-coin/market_chart":                                  "/coins/usd-coin/market_chart",
		"/v1/chains/8453":                                               "/v1/chains/8453",
		"/":                                                             "/",
		"":                                                              "",
	} {
		require.Equal(t, want, (*Attribution)(nil).recordedPath("", path), path)
	}
}

func TestRecordedPath_OmitsPathsOfPrivateCallers(t *testing.T) {
	a := &testAttribution
	require.Equal(t, userPath, a.recordedPath(appModule+"preview.(*Unfurler).Unfurl", "/watch/my-holiday"))
	require.Equal(t, "/api/v3/coins/list", a.recordedPath(appModule+"market.(*Client).fetchTokens", "/api/v3/coins/list"))
	require.Equal(t, "/rpc", a.recordedPath(tagPrefix+"Balances", "/rpc"))
	require.Equal(t, "/v3/{id}", (*Attribution)(nil).recordedPath("", "/v3/0123456789abcdef0123456789abcdef"))
}

func TestEndpointPath_KeepsTheLabelOffPathsOfPrivateCallers(t *testing.T) {
	a := &testAttribution
	require.Equal(t, userPath, a.endpointPath(appModule+"preview.(*Unfurler).Unfurl", "/rpc", "#eth_call"))
	require.Equal(t, "/rpc#eth_call", a.endpointPath(tagPrefix+"Balances", "/rpc", "#eth_call"))
}

func TestRecorder_NeverRecordsAProviderTokenInAPath(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}
	const token = "0123456789abcdef0123456789abcdef"

	get(t, client, server.URL+"/v3/"+token)

	s := rec.Snapshot()
	require.NotNil(t, endpointFor(t, s, "/v3/{id}"))

	summary := strings.Join(s.Summary(5, 5, 5).Endpoints, "\n")
	require.NotContains(t, summary, token)
	require.Contains(t, summary, "/v3/{id}")
}

func TestRecordedPath_MasksLongLowercaseTokens(t *testing.T) {
	a := (*Attribution)(nil)
	require.Equal(t, "/v1/{id}", a.recordedPath("", "/v1/abcdefghijklmnopqrstuvwxyzabcdef"))
	require.Equal(t, "/v1/leaderboard/simpleprices", a.recordedPath("", "/v1/leaderboard/simpleprices"))
}
