package httptraffic

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecorder_PassesThroughWhileDisabled(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder()
	client := &http.Client{Transport: rec.Instrument(nil)}

	rec.SetEnabled(false)
	get(t, client, server.URL+"/a")
	s := rec.Snapshot()
	require.False(t, s.Enabled)
	require.Empty(t, s.Endpoints)
	require.Zero(t, s.Totals.BytesSent+s.Totals.BytesReceived, "a disabled recorder counts no connection bytes")

	rec.SetEnabled(true)
	get(t, client, server.URL+"/a")
	s = rec.Snapshot()
	require.True(t, s.Enabled)
	require.Equal(t, uint64(1), endpointFor(t, s, "/a").Requests)
}

func TestNewRecorder_RecordsAtOnce(t *testing.T) {
	require.True(t, NewRecorder().Enabled())
}

func TestRecorder_ReportSkipsWhileDisabled(t *testing.T) {
	rec := NewRecorder()
	rec.SetEnabled(false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rec.Report(ctx, 10*time.Millisecond, 1000, func(Snapshot) {})

	time.Sleep(50 * time.Millisecond)
	require.Empty(t, rec.Snapshot().Series, "no samples while disabled")

	rec.SetEnabled(true)
	require.Eventually(t, func() bool { return len(rec.Snapshot().Series) > 0 }, 2*time.Second, 5*time.Millisecond)
}
