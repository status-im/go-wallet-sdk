package httptraffic

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecorder_SeparatesTheBackgroundTraffic(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	get(t, client, server.URL+"/a")
	rec.SetBackground(true)
	get(t, client, server.URL+"/a")
	get(t, client, server.URL+"/b")
	rec.SetBackground(false)
	get(t, client, server.URL+"/b")

	s := rec.Snapshot()
	a, b := endpointFor(t, s, "/a"), endpointFor(t, s, "/b")
	require.Equal(t, uint64(1), a.Background.Requests)
	require.Equal(t, a.RequestBytes/2, a.Background.BytesSent, "one of the two requests went in the background")
	require.Equal(t, a.ResponseBytes/2, a.Background.BytesReceived)
	require.Equal(t, uint64(1), b.Background.Requests)

	require.Equal(t, uint64(2), s.Totals.Background.Requests)
	require.Positive(t, s.Totals.Background.BytesSent, "connection bytes are split too")
	require.Less(t, s.Totals.Background.BytesSent, s.Totals.BytesSent)
	require.Equal(t, uint64(2), s.Sources[0].Background.Requests)
	require.False(t, s.Totals.InBackground)
}

func TestRecorder_CountsTheTimeInTheBackground(t *testing.T) {
	clock := newFakeClock()
	rec := NewRecorder(WithClock(clock))
	rec.SetBackground(true)
	clock.Advance(90 * time.Second)
	s := rec.Snapshot()
	require.True(t, s.Totals.InBackground)
	require.Equal(t, 90.0, s.Totals.BackgroundSeconds)

	rec.SetBackground(false)
	clock.Advance(time.Hour)
	require.Equal(t, 90.0, rec.Snapshot().Totals.BackgroundSeconds, "the clock stops in the foreground")

	rec.Reset()
	require.Zero(t, rec.Snapshot().Totals.BackgroundSeconds)
}

func TestRecorder_FlagsIntervalsThatSawTheBackground(t *testing.T) {
	rec := NewRecorder()
	generation := rec.Snapshot().generation

	rec.addInterval(generation, IntervalTotals{})
	rec.SetBackground(true)
	rec.SetBackground(false)
	rec.addInterval(generation, IntervalTotals{})
	rec.SetBackground(true)
	rec.addInterval(generation, IntervalTotals{})
	rec.addInterval(generation, IntervalTotals{})
	rec.SetBackground(false)
	rec.addInterval(generation, IntervalTotals{})
	rec.addInterval(generation, IntervalTotals{})

	var flags []bool
	for _, p := range rec.Snapshot().Series {
		flags = append(flags, p.Background)
	}
	require.Equal(t, []bool{false, true, true, true, true, false}, flags,
		"an interval counts as background when any of it was")
}

func TestSnapshot_SubKeepsTheBackgroundInBetween(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	rec.SetBackground(true)
	get(t, client, server.URL+"/a")
	before := rec.Snapshot()
	get(t, client, server.URL+"/a")

	d := rec.Snapshot().Sub(before)
	require.Equal(t, uint64(1), d.Totals.Background.Requests)
	require.Equal(t, uint64(1), d.Endpoints[0].Background.Requests)
	require.True(t, d.Totals.InBackground)
	require.Positive(t, d.Totals.BackgroundSeconds)
}
