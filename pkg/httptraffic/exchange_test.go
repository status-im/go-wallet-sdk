package httptraffic

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newHeldServer answers with its headers at once and with its body only once
// release is closed, so that tests can act in the middle of an exchange.
func newHeldServer(t *testing.T) (*httptest.Server, chan struct{}) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-release
		_, _ = w.Write([]byte(strings.Repeat("body ", 1000)))
	}))
	t.Cleanup(server.Close)
	return server, release
}

// startHeld sends a request to a held server and returns its response, whose
// body the server has not sent yet.
func startHeld(t *testing.T, client *http.Client, url string) *http.Response {
	resp, err := client.Get(url) //nolint:noctx
	require.NoError(t, err)
	return resp
}

func readAll(t *testing.T, resp *http.Response) {
	_, err := io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func TestRecorder_DropsWhatAnExchangeReadsAfterAReset(t *testing.T) {
	server, release := newHeldServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	resp := startHeld(t, client, server.URL+"/held")
	rec.Reset()
	close(release)
	readAll(t, resp)

	require.Empty(t, rec.Snapshot().Endpoints, "the exchange started before the reset")
}

func TestRecorder_StopsCountingAnExchangeOnceDisabled(t *testing.T) {
	server, release := newHeldServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	resp := startHeld(t, client, server.URL+"/held")
	rec.SetEnabled(false)
	close(release)
	readAll(t, resp)
	rec.SetEnabled(true)

	e := endpointFor(t, rec.Snapshot(), "/held")
	require.Equal(t, uint64(1), e.Requests)
	require.Zero(t, e.DecodedBodyBytes, "the body was read while recording was off")
	require.Zero(t, e.Latency.Max)
}

func TestRecorder_DropsAnExchangeThatWaitedForItsHeadersWhileDisabled(t *testing.T) {
	arrived, release := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(arrived)
		<-release
		_, _ = w.Write([]byte(strings.Repeat("body ", 1000)))
	}))
	t.Cleanup(server.Close)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	responses := make(chan *http.Response, 1)
	go func() {
		resp, err := client.Get(server.URL + "/held") //nolint:noctx
		if err != nil {
			resp = nil
		}
		responses <- resp
	}()
	<-arrived
	rec.SetEnabled(false)
	close(release)
	resp := <-responses
	require.NotNil(t, resp)
	rec.SetEnabled(true)
	readAll(t, resp)

	require.Empty(t, rec.Snapshot().Endpoints, "its request was never recorded")
}

func TestRecorder_KeepsTheHeadersOfAResponseClosedUnread(t *testing.T) {
	server, release := newHeldServer(t)
	defer close(release)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	resp := startHeld(t, client, server.URL+"/held")
	require.NoError(t, resp.Body.Close())

	e := endpointFor(t, rec.Snapshot(), "/held")
	require.Equal(t, uint64(1), e.ExactRequests)
	require.Positive(t, e.ResponseBytes, "the connection carried the response headers")
}

func TestRecorder_KeepsAnExchangeOnTheSideOfTheBackgroundItStartedOn(t *testing.T) {
	server, release := newHeldServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	resp := startHeld(t, client, server.URL+"/held")
	rec.SetBackground(true)
	close(release)
	readAll(t, resp)

	e := endpointFor(t, rec.Snapshot(), "/held")
	require.Positive(t, e.ResponseBytes)
	require.Equal(t, Background{}, e.Background, "the request started in the foreground")
}

func TestRecorder_CountsNoBytesForARequestThatGotNoConnection(t *testing.T) {
	rec := NewRecorder()
	failing := &http.Transport{DialContext: func(context.Context, string, string) (net.Conn, error) {
		return nil, fmt.Errorf("no route")
	}}
	client := &http.Client{Transport: rec.Instrument(failing)}

	_, err := client.Get("http://rpc.example/a") //nolint:noctx,bodyclose
	require.Error(t, err)

	e := endpointFor(t, rec.Snapshot(), "/a")
	require.Equal(t, uint64(1), e.Failed)
	require.Zero(t, e.RequestBytes, "nothing reached a connection")
}

func TestRecorder_CountsNoBytesForAPlainRequestThatGotNoConnectionBesideATLSDialer(t *testing.T) {
	rec := NewRecorder()
	failing := &http.Transport{
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			return nil, fmt.Errorf("no route")
		},
		DialTLSContext: func(context.Context, string, string) (net.Conn, error) {
			return nil, fmt.Errorf("unused")
		},
	}
	client := &http.Client{Transport: rec.Instrument(failing)}

	_, err := client.Get("http://rpc.example/a") //nolint:noctx,bodyclose
	require.Error(t, err)

	require.Zero(t, endpointFor(t, rec.Snapshot(), "/a").RequestBytes, "plain connections are still counted")
}

func TestRecorder_TellsResetsApartWithACoarseClock(t *testing.T) {
	clock := newFakeClock()
	rec := NewRecorder(WithClock(clock))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rec.Report(ctx, 5*time.Millisecond, 1000, func(Snapshot) {})

	for i := 0; i < 5; i++ {
		addExchange(rec, exchange{key: rawKey{host: "h", method: http.MethodGet, path: "/p"}, requestBytes: 10, statusCode: 200})
	}
	require.Eventually(t, func() bool { return len(rec.Snapshot().Series) > 0 }, 2*time.Second, 5*time.Millisecond)
	rec.Reset()
	addExchange(rec, exchange{key: rawKey{host: "h", method: http.MethodGet, path: "/p"}, requestBytes: 10, statusCode: 200})

	require.Eventually(t, func() bool {
		series := rec.Snapshot().Series
		return len(series) > 0 && series[0].Requests == 1
	}, 2*time.Second, 5*time.Millisecond, "the clock never moved, yet the reset is a reset")
	for _, interval := range rec.Snapshot().Series {
		require.LessOrEqual(t, interval.Requests, uint64(1), "no counter went below zero")
	}
}

func TestHostTable_StaysWithinItsCapUnderConcurrentNewHosts(t *testing.T) {
	rec := NewRecorder()
	var wg sync.WaitGroup
	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				rec.addConnBytes(fmt.Sprintf("h%d-%d", g, i), 1, 0)
			}
		}(g)
	}
	wg.Wait()

	require.LessOrEqual(t, len(rec.Snapshot().Hosts), maxHosts+1)
}

func TestRecorder_BoundsTheCallersOfOverflowingEndpoints(t *testing.T) {
	rec := NewRecorder()
	for i := 0; i < maxEndpoints+500; i++ {
		addExchange(rec, exchange{key: rawKey{host: "h", method: http.MethodGet, path: fmt.Sprintf("/p%d", i), caller: fmt.Sprintf("%sc%d", tagPrefix, i)}, requestBytes: 1, statusCode: 200})
	}

	require.LessOrEqual(t, len(rec.Snapshot().Endpoints), maxEndpoints+maxOverflowCallers+1)
}

func TestRecorder_KeepsACallerInItsOverflowBucketOnceCallersAreCapped(t *testing.T) {
	rec := NewRecorder()
	overflow := func(caller string) {
		addExchange(rec, exchange{key: rawKey{host: "h", method: http.MethodGet, path: "/over/" + caller, caller: caller}, statusCode: 200})
	}
	for i := 0; i < maxEndpoints; i++ {
		addExchange(rec, exchange{key: rawKey{host: "h", method: http.MethodGet, path: fmt.Sprintf("/p%d", i)}, statusCode: 200})
	}
	overflow("first")
	for i := 0; i < maxOverflowCallers; i++ {
		overflow(fmt.Sprintf("c%d", i))
	}
	overflow("first")

	first := rec.endpoints[rawKey{host: overflowHost, path: overflowPath, caller: "first"}]
	require.Equal(t, uint64(2), first.requests)
}

func TestRecorder_CountsABodyOfUnknownLength(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}
	body := strings.Repeat("x", 100_000)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/upload", io.MultiReader(strings.NewReader(body)))
	require.NoError(t, err)
	require.Zero(t, req.ContentLength)
	resp, err := client.Do(req)
	require.NoError(t, err)
	readAll(t, resp)

	require.GreaterOrEqual(t, endpointFor(t, rec.Snapshot(), "/upload").RequestBytes, uint64(len(body)))
}

func TestRequestSize_CountsTheHeadersNetHTTPSends(t *testing.T) {
	plain, err := http.NewRequest(http.MethodGet, "http://a.example/p", nil)
	require.NoError(t, err)
	overridden := plain.Clone(context.Background())
	overridden.Host = "a-much-longer-virtual-host.example"
	require.Equal(t, uint64(len(overridden.Host)-len("a.example")), requestSize(overridden)-requestSize(plain),
		"the Host header carries the override")

	ranged := plain.Clone(context.Background())
	ranged.Header.Set("Range", "bytes=0-1")
	head := plain.Clone(context.Background())
	head.Method = http.MethodHead
	gzipLine := uint64(len("Accept-Encoding: gzip\r\n"))
	require.Equal(t, requestSize(plain)+uint64(len("Range: bytes=0-1\r\n"))-gzipLine, requestSize(ranged),
		"net/http asks for no gzip on a range request")
	require.Equal(t, requestSize(plain)+1-gzipLine, requestSize(head), "nor on HEAD")
}

func TestInstrument_LeavesTLSDialingAlone(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()
	pool := server.Client().Transport.(*http.Transport).TLSClientConfig
	rec := NewRecorder()
	transport := &http.Transport{DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&tls.Dialer{Config: pool}).DialContext(ctx, network, addr)
	}}
	client := &http.Client{Transport: rec.Instrument(transport)}

	resp, err := client.Get(server.URL + "/tls") //nolint:noctx
	require.NoError(t, err)
	readAll(t, resp)

	s := rec.Snapshot()
	require.Equal(t, uint64(1), endpointFor(t, s, "/tls").Requests, "requests are still recorded")
	require.Empty(t, s.Hosts, "connections opened by DialTLSContext are not counted")
}

func TestSnapshot_SummaryTakesNegativeLimitsAsNone(t *testing.T) {
	s := Snapshot{Sources: []SourceStats{{Source: "a"}}}
	require.Equal(t, Summary{Sources: []string{}, Hosts: []string{}, Endpoints: []string{}}, s.Summary(-1, -1, -1))
}
