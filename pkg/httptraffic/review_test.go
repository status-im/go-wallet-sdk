package httptraffic

import (
	"context"
	"net"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSnapshot_SubGivesZeroForACountThatExactBytesLowered(t *testing.T) {
	key := EndpointStats{Host: "h", Method: http.MethodPost, Path: "/rpc", Source: "s"}
	before, after := key, key
	before.Requests, before.RequestBytes, before.ResponseBytes = 1, 141, 10
	before.Background = Background{BytesSent: 141}
	after.Requests, after.RequestBytes, after.ResponseBytes = 1, 122, 300
	after.Background = Background{BytesSent: 122}
	prev := Snapshot{Endpoints: []EndpointStats{before}, Hosts: []HostStats{{Host: "h", BytesSent: 141}}}
	cur := Snapshot{Endpoints: []EndpointStats{after}, Hosts: []HostStats{{Host: "h", BytesSent: 122, BytesReceived: 300}}}

	d := cur.Sub(prev)

	require.Len(t, d.Endpoints, 1)
	require.Zero(t, d.Endpoints[0].RequestBytes)
	require.Zero(t, d.Endpoints[0].Background.BytesSent)
	require.Equal(t, uint64(290), d.Endpoints[0].ResponseBytes)
	require.Zero(t, d.Hosts[0].BytesSent)
	require.Zero(t, d.Totals.BytesSent)
}

func TestRecorder_RecordsNeitherHostNorPathOfAPrivateDestination(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	ctx := WithPrivateDestination(WithSource(context.Background(), "Link previews"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/watch/my-holiday", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	readAll(t, resp)

	s := rec.Snapshot()
	require.Len(t, s.Endpoints, 1)
	require.Equal(t, privateHost, s.Endpoints[0].Host)
	require.Equal(t, userPath, s.Endpoints[0].Path)
	require.Equal(t, "Link previews", s.Endpoints[0].Source)
	require.Len(t, s.Hosts, 1, "the connection is counted")
	require.Equal(t, privateHost, s.Hosts[0].Host)
}

func TestSnapshot_OrdersEqualEndpointsByEveryKeyField(t *testing.T) {
	endpoints := func(methods ...string) []rawFact {
		var facts []rawFact
		for _, method := range methods {
			facts = append(facts, rawFact{
				key:      rawKey{host: "h", method: method, path: "/p", caller: tagPrefix + "s"},
				endpoint: rawEndpoint{requests: 1, requestBytes: 10, statusCodes: map[string]uint64{}},
			})
		}
		return facts
	}
	methods := func(s Snapshot) []string {
		var out []string
		for _, e := range s.Endpoints {
			out = append(out, e.Method)
		}
		return out
	}

	a := analyze(facts{endpoints: endpoints(http.MethodPost, http.MethodGet)}, nil)
	b := analyze(facts{endpoints: endpoints(http.MethodGet, http.MethodPost)}, nil)

	require.Equal(t, methods(a), methods(b))
	require.True(t, sort.StringsAreSorted(methods(a)))
}

func TestRecorder_LowersTheLargestRequestToItsExactBytes(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	// An empty User-Agent keeps the header off the wire; the estimate counts it.
	get(t, client, server.URL+"/a", "User-Agent", "")

	e := endpointFor(t, rec.Snapshot(), "/a")
	require.Equal(t, uint64(1), e.ExactRequests)
	require.Equal(t, e.RequestBytes, e.MaxRequestBytes)
}

func TestRecorder_ReportKeepsWhatWasRecordedBeforeAPause(t *testing.T) {
	rec := NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reported := make(chan Snapshot, 16)
	go rec.Report(ctx, 200*time.Millisecond, 1, func(s Snapshot) { reported <- s })
	// The first sample tells that the baseline is taken.
	require.Eventually(t, func() bool { return len(rec.Snapshot().Series) > 0 }, 5*time.Second, 5*time.Millisecond)

	addExchange(rec, exchange{key: rawKey{host: "h", method: http.MethodGet, path: "/before"}, requestBytes: 10, statusCode: 200})
	rec.SetEnabled(false)
	time.Sleep(250 * time.Millisecond) // a sample passes while recording is off
	rec.SetEnabled(true)

	select {
	case s := <-reported:
		require.Len(t, s.Endpoints, 1)
		require.Equal(t, "/before", s.Endpoints[0].Path)
	case <-time.After(2 * time.Second):
		t.Fatal("what was recorded before the pause was never reported")
	}
}

func TestInstrument_KeepsTheDeprecatedDial(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	dialed := 0
	transport := &http.Transport{Dial: func(network, addr string) (net.Conn, error) { //nolint:staticcheck // the deprecated field is the point
		dialed++
		return net.Dial(network, addr)
	}}
	client := &http.Client{Transport: rec.Instrument(transport)}

	get(t, client, server.URL+"/a")

	require.Equal(t, 1, dialed)
	require.Len(t, rec.Snapshot().Hosts, 1, "and its connections are counted")
}

func TestAttribution_NilIsNoFeature(t *testing.T) {
	require.False(t, (*Attribution)(nil).IsFeature(appModule+"market.(*Client).fetch"))
}

func TestRecorder_KeepsThePrivateHostOffAReusedConnection(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}
	get(t, client, server.URL+"/a")
	require.Equal(t, uint64(1), rec.Snapshot().Hosts[0].Connections)
	rec.Reset()

	req, err := http.NewRequestWithContext(WithPrivateDestination(context.Background()), http.MethodGet, server.URL+"/watch", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	readAll(t, resp)

	s := rec.Snapshot()
	require.Len(t, s.Hosts, 1)
	require.Equal(t, privateHost, s.Hosts[0].Host)
	require.Zero(t, s.Hosts[0].Connections, "the connection was reused")
	require.Positive(t, s.Hosts[0].BytesSent)
}

func TestSnapshot_SubKeepsAnEndpointWhoseRequestBytesAloneGrew(t *testing.T) {
	before := EndpointStats{Host: "h", Method: http.MethodPost, Path: "/upload", Source: "s", Requests: 1, ResponseBytes: 88}
	after := before
	after.RequestBytes = 100_000

	d := Snapshot{Endpoints: []EndpointStats{after}}.Sub(Snapshot{Endpoints: []EndpointStats{before}})

	require.Len(t, d.Endpoints, 1)
	require.Equal(t, uint64(100_000), d.Endpoints[0].RequestBytes)
}

func TestSnapshot_NamesTheSameCallerForEqualCounts(t *testing.T) {
	fact := func(caller string) rawFact {
		return rawFact{
			key:      rawKey{host: "h", method: http.MethodGet, path: "/coins/list", caller: appModule + caller},
			endpoint: rawEndpoint{requests: 1, statusCodes: map[string]uint64{}},
		}
	}
	a := analyze(facts{endpoints: []rawFact{fact("market.(*Client).b"), fact("market.(*Client).a")}}, &testAttribution)
	b := analyze(facts{endpoints: []rawFact{fact("market.(*Client).a"), fact("market.(*Client).b")}}, &testAttribution)

	require.Len(t, a.Endpoints, 1)
	require.Equal(t, "market.(*Client).a", a.Endpoints[0].Caller)
	require.Equal(t, a.Endpoints[0].Caller, b.Endpoints[0].Caller)
}

func TestEndpointPath_EscapesControlCharacters(t *testing.T) {
	a := (*Attribution)(nil)
	require.Equal(t, "/a%0Ab", a.endpointPath("", "/a\nb", ""))
	require.Equal(t, "/rpc#eth%0Dcall", a.endpointPath("", "/rpc", "#eth\rcall"))
	require.Equal(t, "/v1/prices", a.endpointPath("", "/v1/prices", ""))
}
