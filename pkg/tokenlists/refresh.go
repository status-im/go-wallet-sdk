package tokenlists

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// refreshWorker handles the background refresh of token lists.
type refreshWorker struct {
	config  *Config
	fetcher *tokenListsFetcher
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool
}

// newRefreshWorker creates a new refresh worker.
func newRefreshWorker(config *Config) *refreshWorker {
	fetcher := newTokenListsFetcher(config)
	return &refreshWorker{
		config:  config,
		fetcher: fetcher,
	}
}

// start begins the background refresh worker.
func (w *refreshWorker) start(ctx context.Context) (refreshCh chan struct{}) {
	if w.running.Load() {
		return
	}

	refreshCh = make(chan struct{})

	childCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.running.Store(true)
	w.wg.Add(1)
	go w.run(childCtx, refreshCh)

	return refreshCh
}

// stop stops the background refresh worker.
func (w *refreshWorker) stop() {
	if !w.running.Load() {
		return
	}

	if w.cancel != nil {
		w.cancel()
	}

	w.wg.Wait()
	w.running.Store(false)
}

func (w *refreshWorker) run(ctx context.Context, refreshCh chan struct{}) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.AutoRefreshCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(refreshCh)
			return
		case <-ticker.C:
			w.checkAndRefresh(ctx, refreshCh)
		}
	}
}

func (w *refreshWorker) checkAndRefresh(ctx context.Context, refreshCh chan struct{}) {
	privacyOn, err := w.config.PrivacyGuard.IsPrivacyOn()
	if err != nil {
		w.config.logger.Error("failed to get privacy mode", zap.Error(err))
		return
	}
	if privacyOn {
		return
	}

	lastRefresh, err := w.config.LastTokenListsUpdateTimeStore.Get()
	if err != nil {
		w.config.logger.Error("failed to get last token lists update time", zap.Error(err))
		return
	}

	if time.Since(lastRefresh) < w.config.AutoRefreshInterval {
		return
	}

	storedListsCount, err := w.fetcher.fetchAndStore(ctx)
	if err != nil {
		w.config.logger.Error("failed to fetch and store token lists", zap.Error(err))
		// Just log the error and don't return, let program continue, cause we have to store last tokens update timestamp
	}

	if storedListsCount > 0 {
		w.config.logger.Info("updated token lists", zap.Int("count", storedListsCount))
	}

	currentTimestamp := time.Unix(time.Now().Unix(), 0)
	err = w.config.LastTokenListsUpdateTimeStore.Set(currentTimestamp)
	if err != nil {
		w.config.logger.Error("failed to save last tokens update time", zap.Error(err))
	}

	refreshCh <- struct{}{}
}
