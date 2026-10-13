package httptraffic

import (
	"cmp"
	"maps"
	"sort"
	"strconv"
	"strings"
	"time"
)

// EndpointStats counts the requests of one endpoint: a host, method, path and
// source.
type EndpointStats struct {
	Host   string `json:"host"`
	Method string `json:"method"`
	Path   string `json:"path"`
	// Source names the feature that made the requests, e.g. "Balances".
	Source string `json:"source"`
	// Caller is the function behind most of the requests, empty when they
	// were tagged with their source.
	Caller string `json:"caller"`

	Requests uint64 `json:"requests"`
	// NotModified counts 304 answers to conditional requests.
	NotModified uint64 `json:"notModified"`
	// Failed counts transport errors and answers outside 2xx/304.
	Failed uint64 `json:"failed"`
	// StatusCodes counts answers by HTTP status, "error" for no answer.
	StatusCodes map[string]uint64 `json:"statusCodes"`

	// RequestBytes and ResponseBytes are what the requests and responses took
	// on the wire. For ExactRequests of them, all over HTTP/1.1, they are the
	// connection's own byte counts; for the rest they are estimates: the
	// HTTP/1.1 size of the request, an upper bound over HTTP/2, and the
	// response as read off the connection.
	RequestBytes    uint64 `json:"requestBytes"`
	MaxRequestBytes uint64 `json:"maxRequestBytes"`
	ResponseBytes   uint64 `json:"responseBytes"`
	ExactRequests   uint64 `json:"exactRequests"`
	// DecodedBodyBytes is the size of the response bodies once decompressed.
	DecodedBodyBytes uint64 `json:"decodedBodyBytes"`

	// Calls counts the calls the requests carried, as the Inspector read
	// them; a batch counts each.
	Calls uint64 `json:"calls"`
	// Bundles counts the calls that bundled further calls, such as eth_calls to
	// Multicall3, and BundledCalls and MaxBundledCalls what they bundled.
	Bundles         uint64 `json:"bundles"`
	BundledCalls    uint64 `json:"bundledCalls"`
	MaxBundledCalls uint64 `json:"maxBundledCalls"`

	// Latency is measured from sending the request to the end of its response
	// body, over the latest latencySamples requests.
	Latency LatencyStats `json:"latencyMs"`

	// Background is the share of the requests and bytes above that went while
	// the app was in the background.
	Background Background `json:"background"`
}

// Background is the part of a count that happened while the app was in the
// background, as told by Recorder.SetBackground. A request counts by where the
// app was when it started; a host's connection bytes by where it was as each
// byte moved, so the two can differ across a switch.
type Background struct {
	Requests      uint64 `json:"requests"`
	BytesSent     uint64 `json:"bytesSent"`
	BytesReceived uint64 `json:"bytesReceived"`
}

func (b Background) sub(p Background) Background {
	return Background{
		Requests:      b.Requests - min(b.Requests, p.Requests),
		BytesSent:     b.BytesSent - min(b.BytesSent, p.BytesSent),
		BytesReceived: b.BytesReceived - min(b.BytesReceived, p.BytesReceived),
	}
}

func (b *Background) add(o Background) {
	b.Requests += o.Requests
	b.BytesSent += o.BytesSent
	b.BytesReceived += o.BytesReceived
}

// LatencyStats holds latency percentiles and the maximum, in milliseconds.
type LatencyStats struct {
	P50 int64 `json:"p50"`
	P95 int64 `json:"p95"`
	Max int64 `json:"max"`
}

// HostStats counts the connections and bytes of one host.
type HostStats struct {
	Host          string `json:"host"`
	Connections   uint64 `json:"connections"`
	BytesSent     uint64 `json:"bytesSent"`
	BytesReceived uint64 `json:"bytesReceived"`
	// Background counts connection bytes only; its Requests stay zero.
	Background Background `json:"background"`
}

// SourceStats sums the endpoints of one source.
type SourceStats struct {
	Source        string     `json:"source"`
	Requests      uint64     `json:"requests"`
	NotModified   uint64     `json:"notModified"`
	Failed        uint64     `json:"failed"`
	BytesSent     uint64     `json:"bytesSent"`
	BytesReceived uint64     `json:"bytesReceived"`
	Background    Background `json:"background"`
}

// Totals sums the connection-level bytes of all hosts and the requests of
// all endpoints.
type Totals struct {
	Requests      uint64 `json:"requests"`
	Failed        uint64 `json:"failed"`
	BytesSent     uint64 `json:"bytesSent"`
	BytesReceived uint64 `json:"bytesReceived"`
	// Background holds connection bytes and requests while in the background.
	Background Background `json:"background"`
	// BackgroundSeconds is how long the app was in the background.
	BackgroundSeconds float64 `json:"backgroundSeconds"`
	// InBackground tells whether the app is in the background right now.
	InBackground bool `json:"inBackground"`
}

// IntervalTotals is the traffic of one sampling interval that ended At.
type IntervalTotals struct {
	At            time.Time `json:"at"`
	Requests      uint64    `json:"requests"`
	BytesSent     uint64    `json:"bytesSent"`
	BytesReceived uint64    `json:"bytesReceived"`
	// Background tells whether the app spent any of the interval in the background.
	Background bool `json:"background"`
}

// Snapshot is what a Recorder has counted between Since and Until.
type Snapshot struct {
	Since     time.Time       `json:"since"`
	Until     time.Time       `json:"until"`
	Totals    Totals          `json:"totals"`
	Hosts     []HostStats     `json:"hosts"`
	Sources   []SourceStats   `json:"sources"`
	Endpoints []EndpointStats `json:"endpoints"`
	// Insights answers "where does the traffic go?" from the figures above.
	Insights Insights `json:"insights"`
	// Enabled tells whether the recorder is recording.
	Enabled bool `json:"enabled"`
	// Series holds the latest sampling intervals, oldest first, when a
	// sampler is sampling the recorder.
	Series []IntervalTotals `json:"series"`

	// generation counts the Resets of the recorder before this snapshot.
	generation uint64
}

// facts are the raw counters of a recorder at one moment.
type facts struct {
	generation        uint64
	since, until      time.Time
	enabled           bool
	inBackground      bool
	backgroundSeconds float64
	hosts             []HostStats
	endpoints         []rawFact
	series            []IntervalTotals
}

type rawFact struct {
	key       rawKey
	endpoint  rawEndpoint
	latencies []int64
}

type endpointKey struct {
	host, method, path, source string
}

// analyze turns raw facts into a snapshot: it classifies each raw endpoint
// into its source, merges the raw endpoints of one source, sums the totals and
// orders hosts, sources and endpoints by the bytes they moved, largest first.
func analyze(f facts, a *Attribution) Snapshot {
	type merged struct {
		stats      EndpointStats
		latencies  []int64
		callerSeen uint64
	}
	byKey := make(map[endpointKey]*merged)
	for _, raw := range f.endpoints {
		source := a.classify(raw.key.caller, raw.key.path)
		key := endpointKey{raw.key.host, raw.key.method, raw.key.path, source}
		m, ok := byKey[key]
		if !ok {
			m = &merged{stats: EndpointStats{
				Host: key.host, Method: key.method, Path: key.path, Source: source,
				StatusCodes: make(map[string]uint64),
			}}
			byKey[key] = m
		}
		e, s := raw.endpoint, &m.stats
		s.Requests += e.requests
		s.NotModified += e.notModified
		s.Failed += e.failed
		s.ExactRequests += e.exactRequests
		s.RequestBytes += e.requestBytes
		s.MaxRequestBytes = max(s.MaxRequestBytes, e.maxRequestBytes)
		s.ResponseBytes += e.responseBytes
		s.DecodedBodyBytes += e.decodedBodyBytes
		s.Calls += e.calls
		s.Bundles += e.bundles
		s.BundledCalls += e.bundledCalls
		s.MaxBundledCalls = max(s.MaxBundledCalls, e.maxBundledCalls)
		s.Background.add(e.background)
		for code, n := range e.statusCodes {
			s.StatusCodes[code] += n
		}
		// Raw endpoints come in map order: equal counts go to the smaller name.
		if function := callerFunction(raw.key.caller); function != "" {
			short := shortFunction(function)
			if e.requests > m.callerSeen || (e.requests == m.callerSeen && short < s.Caller) {
				s.Caller = short
				m.callerSeen = e.requests
			}
		}
		m.latencies = append(m.latencies, raw.latencies...)
	}

	s := Snapshot{
		generation: f.generation,
		Since:      f.since,
		Until:      f.until,
		Hosts:      append([]HostStats(nil), f.hosts...),
		Endpoints:  make([]EndpointStats, 0, len(byKey)),
		Enabled:    f.enabled,
		Series:     f.series,
	}
	for _, m := range byKey {
		m.stats.Latency = latencyStats(m.latencies)
		s.Endpoints = append(s.Endpoints, m.stats)
	}
	s.summarize()
	s.Totals.BackgroundSeconds = f.backgroundSeconds
	s.Totals.InBackground = f.inBackground
	s.Insights = insightsOf(s)
	return s
}

// summarize fills in the totals and sources and orders the tables.
func (s *Snapshot) summarize() {
	s.Totals = Totals{}
	for _, h := range s.Hosts {
		s.Totals.BytesSent += h.BytesSent
		s.Totals.BytesReceived += h.BytesReceived
		s.Totals.Background.add(h.Background)
	}

	sources := make(map[string]*SourceStats)
	for _, e := range s.Endpoints {
		s.Totals.Requests += e.Requests
		s.Totals.Failed += e.Failed
		s.Totals.Background.Requests += e.Background.Requests
		src, ok := sources[e.Source]
		if !ok {
			src = &SourceStats{Source: e.Source}
			sources[e.Source] = src
		}
		src.Requests += e.Requests
		src.NotModified += e.NotModified
		src.Failed += e.Failed
		src.BytesSent += e.RequestBytes
		src.BytesReceived += e.ResponseBytes
		src.Background.add(e.Background)
	}
	s.Sources = make([]SourceStats, 0, len(sources))
	for _, src := range sources {
		s.Sources = append(s.Sources, *src)
	}

	sort.Slice(s.Hosts, func(i, j int) bool {
		a, b := s.Hosts[i], s.Hosts[j]
		if a.BytesSent+a.BytesReceived != b.BytesSent+b.BytesReceived {
			return a.BytesSent+a.BytesReceived > b.BytesSent+b.BytesReceived
		}
		return a.Host < b.Host
	})
	sort.Slice(s.Sources, func(i, j int) bool {
		a, b := s.Sources[i], s.Sources[j]
		if a.BytesSent+a.BytesReceived != b.BytesSent+b.BytesReceived {
			return a.BytesSent+a.BytesReceived > b.BytesSent+b.BytesReceived
		}
		return a.Source < b.Source
	})
	sort.Slice(s.Endpoints, func(i, j int) bool {
		a, b := s.Endpoints[i], s.Endpoints[j]
		if a.RequestBytes+a.ResponseBytes != b.RequestBytes+b.ResponseBytes {
			return a.RequestBytes+a.ResponseBytes > b.RequestBytes+b.ResponseBytes
		}
		return cmp.Or(
			strings.Compare(a.Host, b.Host),
			strings.Compare(a.Method, b.Method),
			strings.Compare(a.Path, b.Path),
			strings.Compare(a.Source, b.Source),
		) < 0
	})
}

// Sub returns the traffic recorded between prev and s, leaving out hosts and
// endpoints that saw none. prev has to be an earlier snapshot of the same
// recorder with no Reset in between. Maxima and latencies are not differences:
// they stay as s has them. A count that exact bytes lowered since prev gives
// zero rather than a negative difference.
func (s Snapshot) Sub(prev Snapshot) Snapshot {
	prevHosts := make(map[string]HostStats, len(prev.Hosts))
	for _, h := range prev.Hosts {
		prevHosts[h.Host] = h
	}
	prevEndpoints := make(map[endpointKey]EndpointStats, len(prev.Endpoints))
	for _, e := range prev.Endpoints {
		prevEndpoints[endpointKey{e.Host, e.Method, e.Path, e.Source}] = e
	}

	d := Snapshot{Since: prev.Until, Until: s.Until, Enabled: s.Enabled, generation: s.generation}
	for _, h := range s.Hosts {
		p := prevHosts[h.Host]
		h.Connections -= min(h.Connections, p.Connections)
		h.BytesSent -= min(h.BytesSent, p.BytesSent)
		h.BytesReceived -= min(h.BytesReceived, p.BytesReceived)
		h.Background = h.Background.sub(p.Background)
		if h.Connections+h.BytesSent+h.BytesReceived > 0 {
			d.Hosts = append(d.Hosts, h)
		}
	}
	for _, e := range s.Endpoints {
		p := prevEndpoints[endpointKey{e.Host, e.Method, e.Path, e.Source}]
		e.Requests -= min(e.Requests, p.Requests)
		e.NotModified -= min(e.NotModified, p.NotModified)
		e.Failed -= min(e.Failed, p.Failed)
		e.ExactRequests -= min(e.ExactRequests, p.ExactRequests)
		e.RequestBytes -= min(e.RequestBytes, p.RequestBytes)
		e.ResponseBytes -= min(e.ResponseBytes, p.ResponseBytes)
		e.DecodedBodyBytes -= min(e.DecodedBodyBytes, p.DecodedBodyBytes)
		e.Calls -= min(e.Calls, p.Calls)
		e.Bundles -= min(e.Bundles, p.Bundles)
		e.BundledCalls -= min(e.BundledCalls, p.BundledCalls)
		e.Background = e.Background.sub(p.Background)
		codes := maps.Clone(e.StatusCodes)
		for code, n := range p.StatusCodes {
			codes[code] -= n
			if codes[code] == 0 {
				delete(codes, code)
			}
		}
		e.StatusCodes = codes
		if e.Requests+e.RequestBytes+e.ResponseBytes > 0 {
			d.Endpoints = append(d.Endpoints, e)
		}
	}
	d.summarize()
	d.Totals.BackgroundSeconds = s.Totals.BackgroundSeconds - prev.Totals.BackgroundSeconds
	d.Totals.InBackground = s.Totals.InBackground
	d.Insights = insightsOf(d)
	return d
}

// Query narrows the endpoints a report carries; everything else stays whole.
type Query struct {
	// Endpoints is "all", "none", "source" or "host".
	Endpoints string `json:"endpoints"`
	// Key is the source or host whose endpoints to keep.
	Key string `json:"key"`
}

// Filter keeps the endpoints the query asks for.
func (s Snapshot) Filter(q Query) Snapshot {
	switch q.Endpoints {
	case "all":
		return s
	case "source", "host":
		kept := make([]EndpointStats, 0)
		for _, e := range s.Endpoints {
			if (q.Endpoints == "source" && e.Source == q.Key) || (q.Endpoints == "host" && e.Host == q.Key) {
				kept = append(kept, e)
			}
		}
		s.Endpoints = kept
	default:
		s.Endpoints = []EndpointStats{}
	}
	return s
}

func statusKey(code int) string {
	if code == 0 {
		return statusTransportError
	}
	return strconv.Itoa(code)
}

func latencyStats(samples []int64) LatencyStats {
	if len(samples) == 0 {
		return LatencyStats{}
	}
	sorted := append([]int64(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	n := len(sorted)
	return LatencyStats{
		P50: sorted[(n-1)*50/100],
		P95: sorted[(n-1)*95/100],
		Max: sorted[n-1],
	}
}
