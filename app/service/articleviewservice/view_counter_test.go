package articleviewservice

import (
	"reflect"
	"testing"
	"time"
)

func TestViewCounterBatchesViewsOnClose(t *testing.T) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	flushed := map[uint64]uint64{}
	counter := &ViewCounter{
		ticker:    ticker,
		requestCh: make(chan uint64, 8),
		closeCh:   make(chan struct{}),
		flushFn: func(counts map[uint64]uint64) error {
			for articleID, count := range counts {
				flushed[articleID] += count
			}
			return nil
		},
	}
	counter.start()

	counter.Record(7)
	counter.Record(7)
	counter.Record(9)
	counter.Close()

	want := map[uint64]uint64{7: 2, 9: 1}
	if !reflect.DeepEqual(flushed, want) {
		t.Fatalf("flushed counts = %#v, want %#v", flushed, want)
	}
}

func TestViewCounterDropsWhenQueueIsFull(t *testing.T) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	counter := &ViewCounter{
		ticker:    ticker,
		requestCh: make(chan uint64, 1),
		closeCh:   make(chan struct{}),
		flushFn: func(map[uint64]uint64) error {
			return nil
		},
	}

	counter.Record(1)
	counter.Record(2)

	if len(counter.requestCh) != 1 {
		t.Fatalf("queue length = %d, want 1", len(counter.requestCh))
	}
}
