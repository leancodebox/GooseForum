package topicviewservice

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

const (
	viewFlushInterval = 10 * time.Second
	viewQueueSize     = 4096
)

type ViewCounter struct {
	ticker    *time.Ticker
	requestCh chan uint64
	closeCh   chan struct{}
	flushFn   func(map[uint64]uint64) error
	closed    bool
	mu        sync.RWMutex
	wg        sync.WaitGroup
}

var (
	counter     *ViewCounter
	counterOnce sync.Once
)

func GetViewCounter() *ViewCounter {
	counterOnce.Do(func() {
		counter = &ViewCounter{
			ticker:    time.NewTicker(viewFlushInterval),
			requestCh: make(chan uint64, viewQueueSize),
			closeCh:   make(chan struct{}),
			flushFn:   topics.IncrementViews,
		}
		counter.start()
		closer.RegisterPriority(closer.PriorityFlush, CloseViewCounter)
		slog.Info("topic view counter started", "flushInterval", viewFlushInterval.String(), "queueSize", viewQueueSize)
	})
	return counter
}

func RecordView(topicID uint64) {
	if topicID == 0 {
		return
	}
	GetViewCounter().Record(topicID)
}

func (c *ViewCounter) Record(topicID uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.closed {
		return
	}

	select {
	case c.requestCh <- topicID:
	default:
		slog.Warn("topic view counter queue full, drop view", "topicId", topicID)
	}
}

func (c *ViewCounter) start() {
	c.wg.Go(func() {
		pending := make(map[uint64]uint64)
		for {
			select {
			case topicID := <-c.requestCh:
				pending[topicID]++
			case <-c.ticker.C:
				c.flush(pending)
				pending = make(map[uint64]uint64)
			case <-c.closeCh:
				c.drain(pending)
				c.flush(pending)
				return
			}
		}
	})
}

func (c *ViewCounter) drain(pending map[uint64]uint64) {
	for {
		select {
		case topicID := <-c.requestCh:
			pending[topicID]++
		default:
			return
		}
	}
}

func (c *ViewCounter) flush(pending map[uint64]uint64) {
	if len(pending) == 0 {
		return
	}
	if c.flushFn == nil {
		slog.Error("flush topic view counts failed", "err", errMissingFlushFn)
		return
	}
	if err := c.flushFn(pending); err != nil {
		slog.Error("flush topic view counts failed", "count", len(pending), "err", err)
		return
	}
	slog.Debug("topic view counts flushed", "count", len(pending))
}

func (c *ViewCounter) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.ticker.Stop()
	close(c.closeCh)
	c.mu.Unlock()

	c.wg.Wait()
}

func CloseViewCounter() error {
	if counter != nil {
		counter.Close()
	}
	return nil
}

var errMissingFlushFn = errors.New("topic view counter flush function is nil")
