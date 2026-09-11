package topicrankservice

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	paniclog "github.com/leancodebox/GooseForum/app/bundles/recovery"
	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"gorm.io/gorm"
)

const (
	defaultRankRate      = 20
	rankBatchSize        = 20
	rankIdleWait         = 5 * time.Second
	rankRetryWait        = 30 * time.Second
	rankDatabaseBackoff  = time.Second
	rankOperationTimeout = 5 * time.Second
)

var rankWorkerOnce sync.Once

// StartWorker runs at most once in this process. Shared-database deployments
// must enable it on only one instance; it does not acquire distributed leases.
func StartWorker() {
	rankWorkerOnce.Do(func() {
		if !preferences.GetBool("ranking.worker_enabled", true) {
			return
		}
		rate := preferences.GetInt("ranking.max_per_second", defaultRankRate)
		if rate < 1 || rate > 1000 {
			slog.Warn("invalid ranking.max_per_second; using default", "value", rate, "default", defaultRankRate)
			rate = defaultRankRate
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		worker := newRankWorker(rate)
		closer.RegisterPriority(closer.PriorityProducer, func() error {
			cancel()
			select {
			case <-done:
				return nil
			case <-time.After(5 * time.Second):
				return errors.New("timed out stopping topic ranking worker")
			}
		})
		go func() {
			defer close(done)
			defer paniclog.Recover("topic_ranking_worker")
			worker.run(ctx, topicrank.Wakeups())
		}()
		slog.Info("topic ranking worker started", "maxPerSecond", rate)
	})
}

// Clock functions keep timing tests deterministic without accelerating the
// production loop or adding an alternative queue implementation.
type rankWorker struct {
	interval  time.Duration
	batchSize int
	now       func() time.Time
	wait      func(context.Context, time.Duration, <-chan struct{}) bool
}

func newRankWorker(rate int) rankWorker {
	return rankWorker{
		interval: time.Second / time.Duration(rate), batchSize: min(rankBatchSize, rate),
		now: time.Now, wait: waitRankWorker,
	}
}

func (w rankWorker) run(ctx context.Context, wake <-chan struct{}) {
	nextStart := w.now()
	lastReport := w.now()
	processed, failed := 0, 0
	var observedLag time.Duration
	for ctx.Err() == nil {
		readCtx, cancel := context.WithTimeout(ctx, rankOperationTimeout)
		entries, err := topicrank.Pending(readCtx, w.batchSize)
		cancel()
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			slog.Error("read topic ranking schedule", "err", err)
			// Events cannot bypass database failure backoff.
			if !w.wait(ctx, rankDatabaseBackoff, nil) {
				return
			}
			continue
		}
		if len(entries) == 0 {
			if !w.wait(ctx, rankIdleWait, wake) {
				return
			}
			continue
		}
		for _, entry := range entries {
			now := w.now()
			if entry.At.After(now) {
				if !w.wait(ctx, min(entry.At.Sub(now), rankIdleWait), wake) {
					return
				}
				// A wakeup may have inserted an earlier entry; reread the indexed head.
				break
			}
			// No burst tokens and no catch-up after a slow operation or a long idle.
			// Wakeups never bypass this pacing delay.
			if delay := nextStart.Sub(now); delay > 0 && !w.wait(ctx, delay, nil) {
				return
			}
			if ctx.Err() != nil {
				return
			}
			observedLag = max(observedLag, w.now().Sub(entry.At))
			retryFailed, err := w.process(ctx, entry)
			nextStart = w.now().Add(w.interval)
			if ctx.Err() != nil {
				return
			}
			processed++
			if err != nil {
				failed++
				slog.Warn("recalculate topic ranking", "topicID", entry.ID, "err", err)
			}
			if w.now().Sub(lastReport) >= 30*time.Second {
				slog.Info("topic ranking progress", "processed", processed, "failed", failed, "observedDueLag", observedLag)
				lastReport, processed, failed, observedLag = w.now(), 0, 0, 0
			}
			if retryFailed {
				if !w.wait(ctx, rankDatabaseBackoff, nil) {
					return
				}
				break
			}
		}
	}
}

func (w rankWorker) process(ctx context.Context, entry topicrank.ScheduledTopic) (bool, error) {
	operationCtx, cancel := context.WithTimeout(ctx, rankOperationTimeout)
	err := recalculateScheduled(operationCtx, entry, w.now())
	cancel()
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	retryCtx, cancelRetry := context.WithTimeout(ctx, rankOperationTimeout)
	defer cancelRetry()
	retryErr := topicrank.Defer(retryCtx, entry, w.now().Add(rankRetryWait))
	return retryErr != nil, errors.Join(err, retryErr)
}

func waitRankWorker(ctx context.Context, delay time.Duration, wake <-chan struct{}) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-wake:
		return true
	case <-timer.C:
		return true
	}
}
