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

type DialContextFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// Instrument returns rt wrapped so that its traffic is recorded; a nil rt
// stands for http.DefaultTransport. An *http.Transport is cloned, so that its
// connections can be counted without touching the original.
func (r *Recorder) Instrument(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	// Only an *http.Transport is known to decompress on its own; taking that
	// over keeps the wire size of its responses visible.
	decompress := false
	if t, ok := rt.(*http.Transport); ok {
		t = t.Clone()
		dial := t.DialContext
		if dial == nil {
			dial = (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext
		}
		t.DialContext = r.WrapDialContext(dial)
		decompress = !t.DisableCompression
		rt = t
	}
	return &roundTripper{base: rt, rec: r, decompress: decompress}
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
}

func (c *countingConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if n > 0 {
		c.received.Add(uint64(n))
		if c.rec.Enabled() {
			c.rec.addConnBytes(c.host, 0, n)
		}
	}
	return n, err
}

func (c *countingConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if n > 0 {
		c.sent.Add(uint64(n))
		if c.rec.Enabled() {
			c.rec.addConnBytes(c.host, n, 0)
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
	conn           *countingConn
	sent, received uint64
}

func (m *connMark) gotConn(info httptrace.GotConnInfo) {
	if c := countingConnOf(info.Conn); c != nil {
		m.conn, m.sent, m.received = c, c.sent.Load(), c.received.Load()
	}
}

type roundTripper struct {
	base http.RoundTripper
	rec  *Recorder
	// decompress makes the round tripper ask for gzip and decode it itself,
	// as net/http would, so that both sizes of a response are known.
	decompress bool
}

func (t *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.rec.Enabled() {
		return t.base.RoundTrip(req)
	}
	start := t.rec.clock.Now()
	inspection := inspect(t.rec.inspector, req)
	attribution := t.rec.attribution
	caller := attribution.callerKey(req.Context())
	key := rawKey{
		host:   req.URL.Hostname(),
		method: req.Method,
		path:   attribution.recordedPath(caller, req.URL.Path) + inspection.Label,
		caller: caller,
	}
	x := exchange{key: key, requestBytes: requestSize(req), inspection: inspection}

	mark := &connMark{}
	ctx := httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{GotConn: mark.gotConn})
	decode := t.decompress && req.Header.Get("Accept-Encoding") == "" &&
		req.Header.Get("Range") == "" && req.Method != http.MethodHead
	if decode {
		req = req.Clone(ctx)
		req.Header.Set("Accept-Encoding", "gzip")
	} else {
		req = req.WithContext(ctx)
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		t.rec.addExchange(x)
		t.rec.finishExchange(key, t.rec.clock.Now().Sub(start), nil)
		return resp, err
	}

	x.statusCode = resp.StatusCode
	x.responseHeaderBytes = responseHeaderSize(resp)
	t.rec.addExchange(x)

	body := &countingBody{
		raw: resp.Body, rec: t.rec, key: key, start: start,
		mark: mark, exact: resp.ProtoMajor == 1,
		estimatedSent: x.requestBytes, counted: x.responseHeaderBytes,
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

// countingBody counts a response body as it comes off the connection and,
// when gzipped, as it is once decoded. When the body ends or is closed it
// takes the request's latency and, over HTTP/1.1, its exact bytes.
type countingBody struct {
	raw     io.ReadCloser
	rec     *Recorder
	key     rawKey
	start   time.Time
	gzipped bool

	mark          *connMark
	exact         bool
	estimatedSent uint64
	// counted is what was recorded as received so far: headers and wire body.
	counted uint64

	decoder io.Reader
	done    sync.Once
}

func (b *countingBody) Read(p []byte) (int, error) {
	if !b.gzipped {
		n, err := b.raw.Read(p)
		if n > 0 {
			b.counted += uint64(n)
			b.rec.addResponseBytes(b.key, n, n)
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
		b.rec.addResponseBytes(b.key, 0, n)
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
		var exact *exactBytes
		if b.exact && b.mark.conn != nil {
			sent := b.mark.conn.sent.Load() - b.mark.sent
			received := b.mark.conn.received.Load() - b.mark.received
			if sent <= b.estimatedSent+maxExactSlack && received <= b.counted+maxExactSlack {
				exact = &exactBytes{estimatedSent: b.estimatedSent, countedReceived: b.counted, sent: sent, received: received}
			}
		}
		b.rec.finishExchange(b.key, b.rec.clock.Now().Sub(b.start), exact)
	})
}

// wireCounter feeds the gzip decoder from the raw body, counting what it reads.
type wireCounter struct{ b *countingBody }

func (w wireCounter) Read(p []byte) (int, error) {
	n, err := w.b.raw.Read(p)
	if n > 0 {
		w.b.counted += uint64(n)
		w.b.rec.addResponseBytes(w.b.key, n, 0)
	}
	return n, err
}

// requestSize estimates the request as HTTP/1.1 puts it on the wire, counting
// in the headers net/http adds on its own.
func requestSize(req *http.Request) uint64 {
	size := len(req.Method) + len(" ") + len(req.URL.RequestURI()) + len(" HTTP/1.1\r\n")
	size += headerLineSize("Host", req.URL.Host)
	size += headersSize(req.Header)
	if req.Header.Get("User-Agent") == "" {
		size += headerLineSize("User-Agent", "Go-http-client/1.1")
	}
	if req.Header.Get("Accept-Encoding") == "" {
		size += headerLineSize("Accept-Encoding", "gzip")
	}
	if req.ContentLength > 0 {
		size += headerLineSize("Content-Length", strconv.FormatInt(req.ContentLength, 10))
		size += int(req.ContentLength)
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
