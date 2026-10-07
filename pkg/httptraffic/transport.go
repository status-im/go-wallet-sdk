package httptraffic

import (
	"compress/gzip"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// DialContextFunc has the signature of net.Dialer.DialContext and
// http.Transport.DialContext.
type DialContextFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// Instrument returns rt wrapped so that its traffic is recorded; a nil rt
// stands for http.DefaultTransport. An *http.Transport is cloned, so that its
// connections can be counted without touching the original; it dials as it
// did, through its DialContext, its Dial or a plain net.Dialer.
//
// Connections an *http.Transport opens with DialTLSContext or DialTLS are not
// counted: net/http negotiates HTTP/2 on the *tls.Conn they return, which a
// counting wrapper would hide. Their requests are recorded with estimated
// bytes. Other round trippers are recorded with estimates only.
func (r *Recorder) Instrument(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	// Only an *http.Transport is known to decompress on its own; taking that
	// over keeps the wire size of its responses visible.
	decompress, countsConnections, dialsTLS := false, false, false
	if t, ok := rt.(*http.Transport); ok {
		t = t.Clone()
		dial := t.DialContext
		switch legacy := t.Dial; { //nolint:staticcheck // Dial is deprecated, still honoured by net/http
		case dial != nil:
		case legacy != nil:
			dial = func(_ context.Context, network, addr string) (net.Conn, error) { return legacy(network, addr) }
		default:
			dial = (&net.Dialer{}).DialContext
		}
		t.DialContext = r.WrapDialContext(dial)
		decompress = !t.DisableCompression
		countsConnections = true
		dialsTLS = t.DialTLSContext != nil || t.DialTLS != nil //nolint:staticcheck // DialTLS is deprecated, still honoured by net/http
		rt = t
	}
	return &roundTripper{base: rt, rec: r, decompress: decompress, countsConnections: countsConnections, dialsTLS: dialsTLS}
}

// WrapDialContext returns dial with every connection it opens counted.
func (r *Recorder) WrapDialContext(dial DialContextFunc) DialContextFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := dial(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		host, _, splitErr := net.SplitHostPort(addr)
		if splitErr != nil {
			host = addr
		}
		if isPrivateDestination(ctx) {
			host = privateHost
		}
		if r.Enabled() {
			r.addConnection(host)
		}
		return &countingConn{Conn: conn, rec: r, host: host}, nil
	}
}

// countingConn counts the bytes of one connection, for its host and for
// itself; its own counts are what makes the bytes of an HTTP/1.1 request exact.
type countingConn struct {
	net.Conn
	rec      *Recorder
	host     string
	sent     atomic.Uint64
	received atomic.Uint64
	// private is set once a request to a private destination got the
	// connection, which another request may have opened: its bytes are
	// counted without the host from then on.
	private atomic.Bool
}

func (c *countingConn) recordedHost() string {
	if c.private.Load() {
		return privateHost
	}
	return c.host
}

func (c *countingConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if n > 0 {
		c.received.Add(uint64(n))
		if c.rec.Enabled() {
			c.rec.addConnBytes(c.recordedHost(), 0, n)
		}
	}
	return n, err
}

func (c *countingConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if n > 0 {
		c.sent.Add(uint64(n))
		if c.rec.Enabled() {
			c.rec.addConnBytes(c.recordedHost(), n, 0)
		}
	}
	return n, err
}

// countingConnOf finds the countingConn under conn, looking through TLS.
func countingConnOf(conn net.Conn) *countingConn {
	for i := 0; i < 4 && conn != nil; i++ {
		if c, ok := conn.(*countingConn); ok {
			return c
		}
		inner, ok := conn.(interface{ NetConn() net.Conn })
		if !ok {
			return nil
		}
		conn = inner.NetConn()
	}
	return nil
}

// connMark remembers which connection a request got and its counts then.
type connMark struct {
	private        bool
	conn           *countingConn
	sent, received uint64
}

func (m *connMark) gotConn(info httptrace.GotConnInfo) {
	if c := countingConnOf(info.Conn); c != nil {
		if m.private {
			c.private.Store(true)
		}
		m.conn, m.sent, m.received = c, c.sent.Load(), c.received.Load()
	}
}

type roundTripper struct {
	base http.RoundTripper
	rec  *Recorder
	// decompress makes the round tripper ask for gzip and decode it itself,
	// as net/http would, so that both sizes of a response are known.
	decompress bool
	// countsConnections tells that the connections base dials are counted, and
	// dialsTLS that it dials those of https requests itself, uncounted.
	countsConnections bool
	dialsTLS          bool
}

// countsConnectionOf tells whether the connection req travels on is counted,
// so that it sent nothing if it failed without getting one.
func (t *roundTripper) countsConnectionOf(req *http.Request) bool {
	return t.countsConnections && !(t.dialsTLS && req.URL.Scheme == "https")
}

func (t *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.rec.Enabled() {
		return t.base.RoundTrip(req)
	}
	start := t.rec.clock.Now()
	inspection := inspect(t.rec.inspector, req)
	attribution := t.rec.attribution
	caller := attribution.callerKey(req.Context())
	host, path := req.URL.Hostname(), attribution.endpointPath(caller, req.URL.Path, inspection.Label)
	reqCtx := req.Context()
	private := isPrivateDestination(reqCtx) || attribution.IsPrivate(callerFunction(caller))
	if private {
		// The mark travels with the context to the dialer, which names the host.
		host, path, reqCtx = privateHost, userPath, WithPrivateDestination(reqCtx)
	}
	x := &exchange{
		key: rawKey{
			host:   host,
			method: req.Method,
			path:   path,
			caller: caller,
		},
		generation:   t.rec.generation.Load(),
		enablings:    t.rec.enablings.Load(),
		background:   t.rec.background.Load(),
		requestBytes: requestSize(req),
		inspection:   inspection,
	}

	mark := &connMark{private: private}
	ctx := httptrace.WithClientTrace(reqCtx, &httptrace.ClientTrace{GotConn: mark.gotConn})
	decode := t.decompress && acceptsGzipByDefault(req)
	if decode {
		req = req.Clone(ctx)
		req.Header.Set("Accept-Encoding", "gzip")
	} else {
		req = req.WithContext(ctx)
	}
	// A body of unknown length is counted as it is read; req is a copy.
	var streamed *countingRequestBody
	if req.Body != nil && req.Body != http.NoBody && req.ContentLength <= 0 {
		streamed = &countingRequestBody{ReadCloser: req.Body}
		req.Body = streamed
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		sentNothing := mark.conn == nil && t.countsConnectionOf(req)
		if sentNothing {
			x.requestBytes = 0
		}
		t.rec.addExchange(x)
		if !sentNothing {
			t.rec.addRequestBytes(x, streamed.count())
		}
		t.rec.finishExchange(x, t.rec.clock.Now().Sub(start), nil)
		return resp, err
	}

	x.statusCode = resp.StatusCode
	x.responseHeaderBytes = responseHeaderSize(resp)
	t.rec.addExchange(x)

	body := &countingBody{
		raw: resp.Body, rec: t.rec, x: x, start: start, streamed: streamed,
		mark: mark, exact: resp.ProtoMajor == 1, counted: x.responseHeaderBytes,
	}
	// Over HTTP/1.1 the request is on the wire once its response headers are
	// back, and the connection carries nothing else until its body is read.
	if mark.conn != nil {
		body.sent = mark.conn.sent.Load() - mark.sent
		body.received = mark.conn.received.Load() - mark.received
	}
	if decode && strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		body.gzipped = true
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
		resp.ContentLength = -1
		resp.Uncompressed = true
	}
	if resp.Body == nil || resp.Body == http.NoBody {
		body.finish()
		return resp, nil
	}
	resp.Body = body
	return resp, nil
}

// acceptsGzipByDefault tells whether net/http asks for gzip on req by itself.
func acceptsGzipByDefault(req *http.Request) bool {
	return req.Header.Get("Accept-Encoding") == "" && req.Header.Get("Range") == "" && req.Method != http.MethodHead
}

// countingRequestBody counts what a request body of unknown length carries.
type countingRequestBody struct {
	io.ReadCloser
	n atomic.Uint64
}

func (b *countingRequestBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.n.Add(uint64(n))
	return n, err
}

func (b *countingRequestBody) count() uint64 {
	if b == nil {
		return 0
	}
	return b.n.Load()
}

// countingBody counts a response body as it comes off the connection and,
// when gzipped, as it is once decoded. When the body ends or is closed it
// takes the request's latency and, over HTTP/1.1, its exact bytes.
type countingBody struct {
	raw      io.ReadCloser
	rec      *Recorder
	x        *exchange
	start    time.Time
	gzipped  bool
	streamed *countingRequestBody

	mark  *connMark
	exact bool
	// sent is what the connection carried for the request, taken when its
	// response headers arrived; received is what it carried for the response,
	// taken then and right after each read of the body, before net/http can
	// hand the connection to another request.
	sent, received uint64
	// counted is what was recorded as received so far: headers and wire body.
	counted uint64

	decoder io.Reader
	done    sync.Once
}

// readRaw reads the wire body, counting it and what the connection carried.
func (b *countingBody) readRaw(p []byte) (int, error) {
	n, err := b.raw.Read(p)
	if b.mark.conn != nil {
		b.received = b.mark.conn.received.Load() - b.mark.received
	}
	if n > 0 {
		b.counted += uint64(n)
	}
	return n, err
}

func (b *countingBody) Read(p []byte) (int, error) {
	if !b.gzipped {
		n, err := b.readRaw(p)
		if n > 0 {
			b.rec.addResponseBytes(b.x, n, n)
		}
		if err != nil {
			b.finish()
		}
		return n, err
	}
	if b.decoder == nil {
		zr, err := gzip.NewReader(wireCounter{b})
		if err != nil {
			b.finish()
			return 0, err
		}
		b.decoder = zr
	}
	n, err := b.decoder.Read(p)
	if n > 0 {
		b.rec.addResponseBytes(b.x, 0, n)
	}
	if err != nil {
		b.finish()
	}
	return n, err
}

func (b *countingBody) Close() error {
	b.finish()
	return b.raw.Close()
}

// maxExactSlack is how far the exact bytes may exceed the estimate before they
// are taken for another request's, which could share a connection only after
// this one released it.
const maxExactSlack = 64 << 10

func (b *countingBody) finish() {
	b.done.Do(func() {
		b.rec.addRequestBytes(b.x, b.streamed.count())
		var exact *exactBytes
		if b.exact && b.mark.conn != nil && b.streamed == nil {
			estimatedSent := b.x.requestBytes
			if b.sent <= estimatedSent+maxExactSlack && b.received <= b.counted+maxExactSlack {
				exact = &exactBytes{estimatedSent: estimatedSent, countedReceived: b.counted, sent: b.sent, received: b.received}
			}
		}
		b.rec.finishExchange(b.x, b.rec.clock.Now().Sub(b.start), exact)
	})
}

// wireCounter feeds the gzip decoder from the raw body, counting what it reads.
type wireCounter struct{ b *countingBody }

func (w wireCounter) Read(p []byte) (int, error) {
	n, err := w.b.readRaw(p)
	if n > 0 {
		w.b.rec.addResponseBytes(w.b.x, n, 0)
	}
	return n, err
}

// requestSize estimates the request as HTTP/1.1 puts it on the wire, counting
// in the headers net/http adds on its own. A body of unknown length is added
// as it is read, without the chunk framing HTTP/1.1 sends it in.
func requestSize(req *http.Request) uint64 {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	size := len(req.Method) + len(" ") + len(req.URL.RequestURI()) + len(" HTTP/1.1\r\n")
	size += headerLineSize("Host", host)
	size += headersSize(req.Header)
	if req.Header.Get("User-Agent") == "" {
		size += headerLineSize("User-Agent", "Go-http-client/1.1")
	}
	if acceptsGzipByDefault(req) {
		size += headerLineSize("Accept-Encoding", "gzip")
	}
	switch {
	case req.ContentLength > 0:
		size += headerLineSize("Content-Length", strconv.FormatInt(req.ContentLength, 10))
		size += int(req.ContentLength)
	case req.Body != nil && req.Body != http.NoBody:
		size += headerLineSize("Transfer-Encoding", "chunked")
	}
	return uint64(size + len("\r\n"))
}

func responseHeaderSize(resp *http.Response) uint64 {
	size := len(resp.Proto) + len(" ") + len(resp.Status) + len("\r\n")
	size += headersSize(resp.Header)
	return uint64(size + len("\r\n"))
}

func headersSize(h http.Header) int {
	size := 0
	for k, vs := range h {
		for _, v := range vs {
			size += headerLineSize(k, v)
		}
	}
	return size
}

func headerLineSize(key, value string) int {
	return len(key) + len(": ") + len(value) + len("\r\n")
}
