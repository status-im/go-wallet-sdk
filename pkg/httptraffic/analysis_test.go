package httptraffic

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecorder_TaggedRequestsTakeTheirSourceFromTheTag(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	req, err := http.NewRequestWithContext(WithSource(context.Background(), "Balances"), http.MethodGet, server.URL+"/a", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	get(t, client, server.URL+"/a")

	s := rec.Snapshot()
	var sources []string
	for _, e := range s.Endpoints {
		sources = append(sources, e.Source)
	}
	require.ElementsMatch(t, []string{"Balances", "Other"}, sources,
		"the tagged and the untagged request land in sources of their own")
	for _, e := range s.Endpoints {
		if e.Source == "Balances" {
			require.Empty(t, e.Caller, "a tagged request names no function: the stack was not walked")
		}
	}
}

func TestAnalyze_MergesTheCallersOfOneSource(t *testing.T) {
	f := facts{endpoints: []rawFact{
		{key: rawKey{"h", http.MethodPost, "/rpc", "example.com/app/balance.(*Controller).fetch"},
			endpoint: rawEndpoint{requests: 3, requestBytes: 300, statusCodes: map[string]uint64{"200": 3}}, latencies: []int64{10, 30}},
		{key: rawKey{"h", http.MethodPost, "/rpc", "example.com/lib/balance/fetcher.FetchBalances.func1"},
			endpoint: rawEndpoint{requests: 5, requestBytes: 500, statusCodes: map[string]uint64{"200": 4, "429": 1}}, latencies: []int64{20}},
		{key: rawKey{"h", http.MethodPost, "/rpc", tagPrefix + "Market: prices"},
			endpoint: rawEndpoint{requests: 1, requestBytes: 10, statusCodes: map[string]uint64{"200": 1}}},
	}}

	s := analyze(f, &Attribution{Sources: []SourceRule{{Function: "/balance", Source: "Balances"}}})
	require.Len(t, s.Endpoints, 2)
	balances := s.Endpoints[0]
	require.Equal(t, "Balances", balances.Source)
	require.Equal(t, uint64(8), balances.Requests)
	require.Equal(t, uint64(800), balances.RequestBytes)
	require.Equal(t, map[string]uint64{"200": 7, "429": 1}, balances.StatusCodes)
	require.Equal(t, "fetcher.FetchBalances", balances.Caller, "the busiest caller names the endpoint")
	require.Equal(t, LatencyStats{P50: 20, P95: 20, Max: 30}, balances.Latency)
}

func TestRecorder_ExactBytesOverHTTP1(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	get(t, client, server.URL+"/gzip")
	get(t, client, server.URL+"/a")

	s := rec.Snapshot()
	var sent, received, exact uint64
	for _, e := range s.Endpoints {
		sent += e.RequestBytes
		received += e.ResponseBytes
		exact += e.ExactRequests
	}
	require.Equal(t, uint64(2), exact)
	require.Equal(t, s.Totals.BytesSent, sent, "every byte the connection sent is some request's")
	require.Equal(t, s.Totals.BytesReceived, received)
}

func TestInsights_NameTheHeaviestSourceAndWhy(t *testing.T) {
	s := Snapshot{
		Sources: []SourceStats{
			{Source: "Balances", BytesSent: 900, BytesReceived: 50},
			{Source: "Market: token list", BytesSent: 5, BytesReceived: 45},
		},
		Endpoints: []EndpointStats{
			{Source: "Balances", Requests: 3, RequestBytes: 900, Bundles: 3, BundledCalls: 7500},
			{Source: "Market: token list", Requests: 1, ResponseBytes: 45},
		},
	}
	require.Equal(t, Insights{
		TopSource: "Balances", TopShare: 0.95, TopUpload: true,
		BytesPerRequest: 300, CallsPerBundle: 2500,
		NextSource: "Market: token list", NextBytes: 50,
	}, insightsOf(s))
	require.Equal(t, Insights{}, insightsOf(Snapshot{}))
}

func TestSnapshot_FilterKeepsTheEndpointsAskedFor(t *testing.T) {
	s := Snapshot{Endpoints: []EndpointStats{
		{Host: "a", Source: "Balances"}, {Host: "b", Source: "Balances"}, {Host: "b", Source: "Activity"},
	}}
	require.Len(t, s.Filter(Query{Endpoints: "all"}).Endpoints, 3)
	require.Empty(t, s.Filter(Query{}).Endpoints)
	require.Len(t, s.Filter(Query{Endpoints: "source", Key: "Balances"}).Endpoints, 2)
	require.Len(t, s.Filter(Query{Endpoints: "host", Key: "b"}).Endpoints, 2)
}

func TestRecorder_UsesTheInjectedClock(t *testing.T) {
	clock := newFakeClock()
	rec := NewRecorder(WithClock(clock))
	clock.Advance(time.Hour)
	s := rec.Snapshot()
	require.Equal(t, time.Hour, s.Until.Sub(s.Since))
}
