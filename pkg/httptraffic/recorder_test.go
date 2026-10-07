package httptraffic

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func gzipped(t *testing.T, s string) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(s))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func newServer(t *testing.T) *httptest.Server {
	payload := strings.Repeat("price ", 1000)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gzip":
			w.Header().Set("Content-Encoding", "gzip")
			_, _ = w.Write(gzipped(t, payload))
		case "/not-modified":
			w.WriteHeader(http.StatusNotModified)
		case "/fail":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			_, _ = w.Write([]byte("ok"))
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func get(t *testing.T, client *http.Client, url string, header ...string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func endpointFor(t *testing.T, s Snapshot, path string) EndpointStats {
	for _, e := range s.Endpoints {
		if e.Path == path {
			return e
		}
	}
	t.Fatalf("no stats for endpoint %s", path)
	return EndpointStats{}
}

func TestRecorder_AttributesRequestsToEndpointsWithoutQuery(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	get(t, client, server.URL+"/prices?ids=a")
	get(t, client, server.URL+"/prices?ids="+strings.Repeat("token,", 500))

	e := endpointFor(t, rec.Snapshot(), "/prices")
	require.Equal(t, uint64(2), e.Requests)
	require.Equal(t, http.MethodGet, e.Method)
	require.Equal(t, "127.0.0.1", e.Host)
	require.Greater(t, e.MaxRequestBytes, uint64(3000), "a long query has to show up in the request size")
	require.Less(t, e.RequestBytes-e.MaxRequestBytes, uint64(500))
}

func TestRecorder_ClassifiesOutcomes(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	get(t, client, server.URL+"/not-modified", "If-None-Match", `"etag"`)
	get(t, client, server.URL+"/fail")

	s := rec.Snapshot()
	require.Equal(t, uint64(1), endpointFor(t, s, "/not-modified").NotModified)
	require.Equal(t, uint64(1), endpointFor(t, s, "/fail").Failed)
}

func TestRecorder_CountsTransportErrorsAsFailed(t *testing.T) {
	rec := NewRecorder()
	failing := &http.Transport{DialContext: func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("offline")
	}}
	client := &http.Client{Transport: rec.Instrument(failing)}

	_, err := client.Get("http://example.invalid/prices") //nolint:noctx
	require.Error(t, err)

	e := endpointFor(t, rec.Snapshot(), "/prices")
	require.Equal(t, uint64(1), e.Requests)
	require.Equal(t, uint64(1), e.Failed)
}

func TestRecorder_ResponseBytesAreWireBytesWhenClientDecodes(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}
	wireBody := uint64(len(gzipped(t, strings.Repeat("price ", 1000))))

	get(t, client, server.URL+"/gzip", "Accept-Encoding", "gzip")

	e := endpointFor(t, rec.Snapshot(), "/gzip")
	require.Equal(t, wireBody, e.DecodedBodyBytes, "a client decoding on its own leaves the body as it came")
	require.Greater(t, e.ResponseBytes, wireBody)
	require.Less(t, e.ResponseBytes, wireBody+300, "only the status line and headers come on top of the body")
}

func TestRecorder_DecodesGzipItselfAndCountsBothSizes(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}
	payload := strings.Repeat("price ", 1000)
	wireBody := uint64(len(gzipped(t, payload)))

	resp, err := client.Get(server.URL + "/gzip") //nolint:noctx
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, payload, string(body))
	require.True(t, resp.Uncompressed)
	require.Empty(t, resp.Header.Get("Content-Encoding"))

	e := endpointFor(t, rec.Snapshot(), "/gzip")
	require.Equal(t, uint64(len(payload)), e.DecodedBodyBytes)
	require.Greater(t, e.ResponseBytes, wireBody)
	require.Less(t, e.ResponseBytes, wireBody+300)
}

func TestRecorder_CountsStatusCodesAndLatency(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	get(t, client, server.URL+"/fail")
	get(t, client, server.URL+"/fail")
	get(t, client, server.URL+"/a")

	s := rec.Snapshot()
	require.Equal(t, map[string]uint64{"500": 2}, endpointFor(t, s, "/fail").StatusCodes)
	require.Equal(t, map[string]uint64{"200": 1}, endpointFor(t, s, "/a").StatusCodes)
	require.GreaterOrEqual(t, endpointFor(t, s, "/a").Latency.Max, endpointFor(t, s, "/a").Latency.P50)
}

func TestRecorder_SummarizesTotalsAndSources(t *testing.T) {
	rec := NewRecorder()
	addExchange(rec, exchange{key: rawKey{"h", http.MethodGet, "/a", tagPrefix + "Balances"}, requestBytes: 100, statusCode: 200})
	addExchange(rec, exchange{key: rawKey{"h", http.MethodGet, "/b", tagPrefix + "Balances"}, requestBytes: 50, statusCode: 500})
	addExchange(rec, exchange{key: rawKey{"h", http.MethodGet, "/c", tagPrefix + "Market: prices"}, requestBytes: 10, statusCode: 200})
	addExchange(rec, exchange{key: rawKey{"h", http.MethodGet, "/c", tagPrefix + "Market: prices"}, requestBytes: 10, statusCode: 304})

	s := rec.Snapshot()
	require.Equal(t, uint64(4), s.Totals.Requests)
	require.Equal(t, uint64(1), s.Totals.Failed)
	require.Equal(t, []SourceStats{
		{Source: "Balances", Requests: 2, Failed: 1, BytesSent: 150},
		{Source: "Market: prices", Requests: 2, NotModified: 1, BytesSent: 20},
	}, s.Sources)
}

func TestRecorder_CountsConnectionBytesPerHost(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	get(t, client, server.URL+"/a")
	get(t, client, server.URL+"/b")

	s := rec.Snapshot()
	require.Len(t, s.Hosts, 1)
	h := s.Hosts[0]
	require.Equal(t, "127.0.0.1", h.Host)
	require.Equal(t, uint64(1), h.Connections, "keep-alive has to reuse the connection")

	var requestBytes, responseBytes uint64
	for _, e := range s.Endpoints {
		requestBytes += e.RequestBytes
		responseBytes += e.ResponseBytes
	}
	require.Equal(t, requestBytes, h.BytesSent, "over plain HTTP/1.1 the estimate is exact")
	require.Equal(t, responseBytes, h.BytesReceived)
}

func TestRecorder_InstrumentLeavesTheOriginalTransportAlone(t *testing.T) {
	original := &http.Transport{}
	_ = NewRecorder().Instrument(original)
	require.Nil(t, original.DialContext)
}

func TestSnapshot_SubKeepsOnlyTheTrafficInBetween(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	get(t, client, server.URL+"/a")
	before := rec.Snapshot()
	get(t, client, server.URL+"/b")
	get(t, client, server.URL+"/b")

	d := rec.Snapshot().Sub(before)
	require.Len(t, d.Endpoints, 1)
	require.Equal(t, "/b", d.Endpoints[0].Path)
	require.Equal(t, uint64(2), d.Endpoints[0].Requests)
	require.Len(t, d.Hosts, 1)
	require.Zero(t, d.Hosts[0].Connections)
	require.Positive(t, d.Hosts[0].BytesSent)
}

func TestRecorder_PoolsEndpointsBeyondTheCap(t *testing.T) {
	rec := NewRecorder()
	for i := 0; i < maxEndpoints+10; i++ {
		addExchange(rec, exchange{key: rawKey{host: "h", method: http.MethodGet, path: fmt.Sprintf("/coins/%d", i)}, requestBytes: 10, statusCode: 200})
	}

	s := rec.Snapshot()
	require.Len(t, s.Endpoints, maxEndpoints+1)
	require.Equal(t, uint64(10), endpointFor(t, s, overflowPath).Requests)
}

func TestRecorder_PoolsEndpointsOfEveryNewHostBeyondTheCap(t *testing.T) {
	rec := NewRecorder()
	for i := 0; i < maxEndpoints+1000; i++ {
		addExchange(rec, exchange{key: rawKey{host: fmt.Sprintf("h%d", i), method: http.MethodGet, path: "/p", caller: "c"}, requestBytes: 10, statusCode: 200})
	}

	s := rec.Snapshot()
	require.Len(t, s.Endpoints, maxEndpoints+1)
	require.Equal(t, uint64(1000), endpointFor(t, s, overflowPath).Requests)
}

func TestRecorder_PoolsHostsBeyondTheCap(t *testing.T) {
	rec := NewRecorder()
	for i := 0; i < maxHosts+50; i++ {
		host := fmt.Sprintf("h%d", i)
		rec.addConnection(host)
		rec.addConnBytes(host, 10, 20)
	}

	s := rec.Snapshot()
	require.Len(t, s.Hosts, maxHosts+1)
	var other *HostStats
	for i := range s.Hosts {
		if s.Hosts[i].Host == overflowHost {
			other = &s.Hosts[i]
		}
	}
	require.NotNil(t, other)
	require.Equal(t, uint64(50), other.Connections)
	require.Equal(t, uint64(500), other.BytesSent)
}

func TestRecorder_Reset(t *testing.T) {
	rec := NewRecorder()
	addExchange(rec, exchange{key: rawKey{host: "h", method: http.MethodGet, path: "/p"}, requestBytes: 10, statusCode: 200})
	before := rec.Snapshot().Since

	rec.Reset()

	s := rec.Snapshot()
	require.Empty(t, s.Endpoints)
	require.True(t, s.Since.After(before) || s.Since.Equal(before))
}

func TestRecorder_KeepsTheLatestIntervals(t *testing.T) {
	rec := NewRecorder()
	generation := rec.Snapshot().generation
	for i := 0; i < maxSeries+5; i++ {
		rec.addInterval(generation, IntervalTotals{Requests: uint64(i)})
	}

	series := rec.Snapshot().Series
	require.Len(t, series, maxSeries)
	require.Equal(t, uint64(5), series[0].Requests, "the oldest intervals go first")
	require.Equal(t, uint64(maxSeries+4), series[maxSeries-1].Requests)

	rec.addInterval(generation-1, IntervalTotals{Requests: 999})
	require.Len(t, rec.Snapshot().Series, maxSeries, "an interval from before a reset is dropped")

	rec.Reset()
	require.Empty(t, rec.Snapshot().Series)
}

func TestRecorder_ReportSamplesIntoTheSeries(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		rec.Report(ctx, 20*time.Millisecond, 1000, func(Snapshot) {})
	}()

	get(t, client, server.URL+"/a")
	require.Eventually(t, func() bool {
		for _, s := range rec.Snapshot().Series {
			if s.Requests == 1 && s.BytesSent > 0 {
				return true
			}
		}
		return false
	}, 2*time.Second, 10*time.Millisecond)

	cancel()
	<-done
}

// addExchange records x as an exchange that started now.
func addExchange(rec *Recorder, x exchange) {
	x.generation = rec.generation.Load()
	x.enablings = rec.enablings.Load()
	rec.addExchange(&x)
}
