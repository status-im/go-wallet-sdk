package httptraffic

// Insights answers "where does the traffic go?" from a snapshot's figures, so
// that every client tells the same story.
type Insights struct {
	// TopSource is the source that moved the most bytes, empty with no traffic.
	TopSource string `json:"topSource"`
	// TopShare is its part of all request and response bytes, from 0 to 1.
	TopShare float64 `json:"topShare"`
	// TopUpload tells whether it mostly sends rather than receives.
	TopUpload bool `json:"topUpload"`
	// BytesPerRequest is what one request of its heaviest endpoint sends, when
	// TopUpload, or receives otherwise.
	BytesPerRequest uint64 `json:"bytesPerRequest"`
	// CallsPerBundle is how many calls a bundling call of that endpoint, such
	// as an eth_call to Multicall3, bundles; zero when it makes none.
	CallsPerBundle uint64 `json:"callsPerBundle"`
	// NextSource is the runner-up and NextBytes what it moved.
	NextSource string `json:"nextSource"`
	NextBytes  uint64 `json:"nextBytes"`
}

func insightsOf(s Snapshot) Insights {
	var in Insights
	if len(s.Sources) == 0 {
		return in
	}
	var all uint64
	for _, src := range s.Sources {
		all += src.BytesSent + src.BytesReceived
	}
	top := s.Sources[0]
	topBytes := top.BytesSent + top.BytesReceived
	if all == 0 || topBytes == 0 {
		return in
	}
	in.TopSource = top.Source
	in.TopShare = float64(topBytes) / float64(all)
	in.TopUpload = top.BytesSent >= top.BytesReceived

	// Endpoints are ordered heaviest first, so the first of the source wins.
	for _, e := range s.Endpoints {
		if e.Source != top.Source || e.Requests == 0 {
			continue
		}
		if in.TopUpload {
			in.BytesPerRequest = e.RequestBytes / e.Requests
		} else {
			in.BytesPerRequest = e.ResponseBytes / e.Requests
		}
		if e.Bundles > 0 {
			in.CallsPerBundle = e.BundledCalls / e.Bundles
		}
		break
	}
	if len(s.Sources) > 1 {
		in.NextSource = s.Sources[1].Source
		in.NextBytes = s.Sources[1].BytesSent + s.Sources[1].BytesReceived
	}
	return in
}
