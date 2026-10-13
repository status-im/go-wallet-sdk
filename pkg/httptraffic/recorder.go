package httptraffic

import (
	"sync"
	"sync/atomic"
	"time"
)

// maxEndpoints bounds the raw endpoint table and maxHosts the host table:
// clients that follow links reach any number of hosts in a long session.
// Requests beyond the endpoint cap are pooled per caller under overflowHost
// and overflowPath, for at most maxOverflowCallers callers and then all
// together; hosts beyond the host cap are pooled under overflowHost.
const (
	maxEndpoints       = 500
	maxOverflowCallers = 64
	maxHosts           = 200
	overflowHost       = "<other hosts>"
	privateHost        = "<private>"
	overflowPath       = "<other>"
)

// statusTransportError is the StatusCodes key of requests that got no answer.
const statusTransportError = "error"

// maxSeries bounds the sampling intervals a recorder keeps.
const maxSeries = 60

// Clock tells the time; tests use a fake one.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// Option configures a Recorder.
type Option func(*Recorder)

// WithClock makes the recorder tell the time by c.
func WithClock(c Clock) Option {
	return func(r *Recorder) { r.clock = c }
}

// WithAttribution makes the recorder name sources by a.
func WithAttribution(a Attribution) Option {
	return func(r *Recorder) { r.attribution = &a }
}

// WithInspector makes the recorder read request bodies with in.
func WithInspector(in Inspector) Option {
	return func(r *Recorder) { r.inspector = in }
}

// rawKey identifies a raw endpoint: who made the requests is part of it, so
// that classification can happen at analysis time.
type rawKey struct {
	host, method, path string
	// caller is a source tag (tagPrefix + source) or a calling function.
	caller string
}

type rawEndpoint struct {
	requests         uint64
	notModified      uint64
	failed           uint64
	exactRequests    uint64
	statusCodes      map[string]uint64
	requestBytes     uint64
	maxRequestBytes  uint64
	responseBytes    uint64
	decodedBodyBytes uint64
	calls            uint64
	bundles          uint64
	bundledCalls     uint64
	maxBundledCalls  uint64
	background       Background
	latencies        latencyRing
}

// hostTable maps a host to its counters. It is swapped whole on Reset, and
// read without a lock on every socket read and write.
type hostTable struct {
	m     sync.Map // string -> *hostCounters
	count atomic.Int64
}

func (t *hostTable) get(host string) *hostCounters {
	if v, ok := t.m.Load(host); ok {
		return v.(*hostCounters)
	}
	// A slot is reserved before the host is stored, so that concurrent first
	// connections to distinct hosts cannot overshoot the cap together.
	for {
		n := t.count.Load()
		if n >= maxHosts {
			v, _ := t.m.LoadOrStore(overflowHost, &hostCounters{})
			return v.(*hostCounters)
		}
		if t.count.CompareAndSwap(n, n+1) {
			break
		}
	}
	v, loaded := t.m.LoadOrStore(host, &hostCounters{})
	if loaded {
		t.count.Add(-1)
	}
	return v.(*hostCounters)
}

type hostCounters struct {
	connections        atomic.Uint64
	bytesSent          atomic.Uint64
	bytesReceived      atomic.Uint64
	backgroundSent     atomic.Uint64
	backgroundReceived atomic.Uint64
}

// Recorder counts the HTTP traffic of the round trippers it instruments.
type Recorder struct {
	clock       Clock
	attribution *Attribution
	inspector   Inspector
	enabled     atomic.Bool
	background  atomic.Bool
	hosts       atomic.Pointer[hostTable]
	// generation counts Resets. An exchange that started before the latest one
	// records nothing more, and a reset is told apart by it rather than by
	// the time, which a coarse clock may not advance.
	generation atomic.Uint64
	// enablings counts the times recording was turned on. An exchange that
	// started before the latest one records nothing more.
	enablings atomic.Uint64

	// mu guards everything below.
	mu              sync.Mutex
	since           time.Time
	endpoints       map[rawKey]*rawEndpoint
	overflowCallers int
	series          []IntervalTotals
	// backgroundSince is when the app went to the background, or the last
	// Reset if that came later.
	backgroundSince   time.Time
	backgroundSeconds float64
	// backgroundSeen marks that the current sampling interval saw the background.
	backgroundSeen bool
}

// NewRecorder returns an enabled recorder.
func NewRecorder(opts ...Option) *Recorder {
	r := &Recorder{clock: systemClock{}}
	for _, opt := range opts {
		opt(r)
	}
	r.enabled.Store(true)
	r.Reset()
	return r
}

// SetEnabled turns recording on or off. While off the instrumented transports
// pass requests straight through and the counters stay as they are.
func (r *Recorder) SetEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if enabled && !r.enabled.Load() {
		r.enablings.Add(1)
	}
	r.enabled.Store(enabled)
}

// Enabled tells whether the recorder is recording.
func (r *Recorder) Enabled() bool {
	return r.enabled.Load()
}

// SetBackground tells the recorder whether the app is in the background, so
// that it can tell that share of the traffic apart.
func (r *Recorder) SetBackground(background bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if background == r.background.Load() {
		return
	}
	now := r.clock.Now()
	if background {
		r.backgroundSince = now
		r.backgroundSeen = true
	} else {
		r.backgroundSeconds += now.Sub(r.backgroundSince).Seconds()
	}
	r.background.Store(background)
}

// Reset drops everything recorded so far.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generation.Add(1)
	r.since = r.clock.Now()
	r.hosts.Store(&hostTable{})
	r.endpoints = make(map[rawKey]*rawEndpoint)
	r.overflowCallers = 0
	r.series = nil
	r.backgroundSince = r.since
	r.backgroundSeconds = 0
	r.backgroundSeen = r.background.Load()
}

// addInterval appends a sampling interval, dropping the oldest beyond maxSeries.
// An interval measured across a Reset is dropped.
func (r *Recorder) addInterval(generation uint64, interval IntervalTotals) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if generation != r.generation.Load() {
		return
	}
	interval.Background = r.backgroundSeen
	r.backgroundSeen = r.background.Load()
	r.series = append(r.series, interval)
	if len(r.series) > maxSeries {
		r.series = append([]IntervalTotals(nil), r.series[len(r.series)-maxSeries:]...)
	}
}

// Snapshot returns the recorded traffic, analysed.
func (r *Recorder) Snapshot() Snapshot {
	return analyze(r.facts(), r.attribution)
}

// facts copies the raw counters out of the recorder.
func (r *Recorder) facts() facts {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock.Now()
	f := facts{
		generation:        r.generation.Load(),
		since:             r.since,
		until:             now,
		enabled:           r.Enabled(),
		inBackground:      r.background.Load(),
		backgroundSeconds: r.backgroundSeconds,
		series:            append([]IntervalTotals(nil), r.series...),
	}
	if f.inBackground {
		f.backgroundSeconds += now.Sub(r.backgroundSince).Seconds()
	}
	r.hosts.Load().m.Range(func(k, v any) bool {
		c := v.(*hostCounters)
		f.hosts = append(f.hosts, HostStats{
			Host:          k.(string),
			Connections:   c.connections.Load(),
			BytesSent:     c.bytesSent.Load(),
			BytesReceived: c.bytesReceived.Load(),
			Background: Background{
				BytesSent:     c.backgroundSent.Load(),
				BytesReceived: c.backgroundReceived.Load(),
			},
		})
		return true
	})
	for key, e := range r.endpoints {
		c := *e
		c.statusCodes = make(map[string]uint64, len(e.statusCodes))
		for code, n := range e.statusCodes {
			c.statusCodes[code] = n
		}
		f.endpoints = append(f.endpoints, rawFact{key: key, endpoint: c, latencies: e.latencies.samplesCopy()})
	}
	return f
}

func (r *Recorder) hostCounters(host string) *hostCounters {
	return r.hosts.Load().get(host)
}

func (r *Recorder) addConnection(host string) {
	r.hostCounters(host).connections.Add(1)
}

func (r *Recorder) addConnBytes(host string, sent, received int) {
	c := r.hostCounters(host)
	background := r.background.Load()
	if sent > 0 {
		c.bytesSent.Add(uint64(sent))
		if background {
			c.backgroundSent.Add(uint64(sent))
		}
	}
	if received > 0 {
		c.bytesReceived.Add(uint64(received))
		if background {
			c.backgroundReceived.Add(uint64(received))
		}
	}
}

func (r *Recorder) endpoint(key rawKey) *rawEndpoint {
	e, ok := r.endpoints[key]
	if ok {
		return e
	}
	if len(r.endpoints) >= maxEndpoints {
		key = rawKey{host: overflowHost, path: overflowPath, caller: key.caller}
		if e, ok = r.endpoints[key]; ok {
			return e
		}
		if key.caller != "" && r.overflowCallers >= maxOverflowCallers {
			key.caller = ""
			if e, ok = r.endpoints[key]; ok {
				return e
			}
		}
		if key.caller != "" {
			r.overflowCallers++
		}
	}
	e = &rawEndpoint{statusCodes: make(map[string]uint64)}
	r.endpoints[key] = e
	return e
}

// exchange is one request as it is recorded: what was known when it started,
// and what its response told once its headers arrived or it failed.
type exchange struct {
	key rawKey
	// generation, enablings and background are the recorder's when the request
	// started: a Reset or a pause in recording since drops the rest of the
	// exchange, and its bytes stay on the side of the background they started on.
	generation uint64
	enablings  uint64
	// maxBefore is the endpoint's largest request before this one was added.
	maxBefore           uint64
	background          bool
	requestBytes        uint64
	responseHeaderBytes uint64
	statusCode          int // 0 when the request got no answer
	inspection          Inspection
}

// current tells whether x still belongs to what the recorder records: no
// Reset since it started, and recording on ever since.
func (r *Recorder) current(x *exchange) bool {
	return x.generation == r.generation.Load() && x.enablings == r.enablings.Load() && r.Enabled()
}

func (r *Recorder) addExchange(x *exchange) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.current(x) {
		return
	}
	e := r.endpoint(x.key)
	e.requests++
	e.requestBytes += x.requestBytes
	x.maxBefore = e.maxRequestBytes
	e.maxRequestBytes = max(e.maxRequestBytes, x.requestBytes)
	e.responseBytes += x.responseHeaderBytes
	e.calls += x.inspection.Calls
	e.bundles += x.inspection.Bundles
	e.bundledCalls += x.inspection.BundledCalls
	e.maxBundledCalls = max(e.maxBundledCalls, x.inspection.MaxBundledCalls)
	if x.background {
		e.background.Requests++
		e.background.BytesSent += x.requestBytes
		e.background.BytesReceived += x.responseHeaderBytes
	}
	e.statusCodes[statusKey(x.statusCode)]++
	switch {
	case x.statusCode == 0:
		e.failed++
	case x.statusCode == 304:
		e.notModified++
	case x.statusCode < 200 || x.statusCode >= 300:
		e.failed++
	}
}

// addRequestBytes adds what a body of unknown length turned out to carry.
func (r *Recorder) addRequestBytes(x *exchange, n uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.current(x) || n == 0 {
		return
	}
	e := r.endpoint(x.key)
	x.requestBytes += n
	e.requestBytes += n
	e.maxRequestBytes = max(e.maxRequestBytes, x.requestBytes)
	if x.background {
		e.background.BytesSent += n
	}
}

func (r *Recorder) addResponseBytes(x *exchange, wire, decoded int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.current(x) {
		return
	}
	e := r.endpoint(x.key)
	e.responseBytes += uint64(wire)
	e.decodedBodyBytes += uint64(decoded)
	if x.background {
		e.background.BytesReceived += uint64(wire)
	}
}

// exactBytes replaces the estimates a request was recorded with by what its
// connection actually carried for it.
type exactBytes struct {
	estimatedSent, countedReceived uint64
	sent, received                 uint64
}

func (r *Recorder) finishExchange(x *exchange, latency time.Duration, exact *exactBytes) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.current(x) {
		return
	}
	e := r.endpoint(x.key)
	e.latencies.add(latency.Milliseconds())
	if exact == nil {
		return
	}
	e.exactRequests++
	e.requestBytes = e.requestBytes - min(e.requestBytes, exact.estimatedSent) + exact.sent
	// An estimate that turned out too high stops being the largest request.
	if e.maxRequestBytes == exact.estimatedSent {
		e.maxRequestBytes = x.maxBefore
	}
	e.maxRequestBytes = max(e.maxRequestBytes, exact.sent)
	e.responseBytes = e.responseBytes - min(e.responseBytes, exact.countedReceived) + exact.received
	if x.background {
		b := &e.background
		b.BytesSent = b.BytesSent - min(b.BytesSent, exact.estimatedSent) + exact.sent
		b.BytesReceived = b.BytesReceived - min(b.BytesReceived, exact.countedReceived) + exact.received
	}
}

// latencySamples is how many of the latest latencies an endpoint keeps.
const latencySamples = 256

type latencyRing struct {
	samples [latencySamples]int64
	n       int
}

func (l *latencyRing) add(ms int64) {
	l.samples[l.n%latencySamples] = ms
	l.n++
}

func (l *latencyRing) samplesCopy() []int64 {
	n := min(l.n, latencySamples)
	return append([]int64(nil), l.samples[:n]...)
}
