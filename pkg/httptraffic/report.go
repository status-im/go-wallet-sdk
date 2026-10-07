package httptraffic

import (
	"context"
	"fmt"
	"time"
)

// Report samples the recorder every sampleInterval into its series until ctx
// is done, and every samplesPerReport samples hands the traffic since the
// previous report to report. Reports without traffic are skipped, and so is
// the time recording is off.
func (r *Recorder) Report(ctx context.Context, sampleInterval time.Duration, samplesPerReport int, report func(Snapshot)) {
	ticker := time.NewTicker(sampleInterval)
	defer ticker.Stop()

	prevSample := r.Snapshot()
	prevReport := prevSample
	samples := 0
	// stale is set while recording is off: the baseline then predates
	// whatever the counters missed, so it is taken again once it is back on.
	stale := !r.Enabled()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !r.Enabled() {
				stale = true
				continue
			}
			current := r.Snapshot()
			if stale {
				prevSample, prevReport, samples, stale = current, current, 0, false
				continue
			}
			if current.Since != prevSample.Since {
				// Reset in between: the counters started over.
				prevSample = Snapshot{Since: current.Since, Until: current.Since}
				prevReport = prevSample
				samples = 0
			}
			d := current.Sub(prevSample)
			r.addInterval(current.Since, IntervalTotals{
				At:            current.Until,
				Requests:      d.Totals.Requests,
				BytesSent:     d.Totals.BytesSent,
				BytesReceived: d.Totals.BytesReceived,
			})
			prevSample = current

			samples++
			if samples >= samplesPerReport {
				if delta := current.Sub(prevReport); len(delta.Hosts) > 0 || len(delta.Endpoints) > 0 {
					report(delta)
				}
				prevReport = current
				samples = 0
			}
		}
	}
}

// Summary is a snapshot in a few short lines, heaviest first, e.g. for a log.
type Summary struct {
	Sources   []string
	Hosts     []string
	Endpoints []string
}

// Summary lists at most the given number of sources, hosts and endpoints.
func (s Snapshot) Summary(sources, hosts, endpoints int) Summary {
	return Summary{
		Sources: lines(s.Sources, sources, func(src SourceStats) string {
			return fmt.Sprintf("%s %d req %d↑ %d↓", src.Source, src.Requests, src.BytesSent, src.BytesReceived)
		}),
		Hosts: lines(s.Hosts, hosts, func(h HostStats) string {
			return fmt.Sprintf("%s %d↑ %d↓", h.Host, h.BytesSent, h.BytesReceived)
		}),
		Endpoints: lines(s.Endpoints, endpoints, func(e EndpointStats) string {
			return fmt.Sprintf("%s %s%s [%s] %d req %d↑ %d↓", e.Method, e.Host, e.Path, e.Source, e.Requests, e.RequestBytes, e.ResponseBytes)
		}),
	}
}

func lines[T any](items []T, limit int, line func(T) string) []string {
	out := make([]string, 0, min(len(items), limit))
	for _, item := range items[:min(len(items), limit)] {
		out = append(out, line(item))
	}
	return out
}
