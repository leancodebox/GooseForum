package sharedcache

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestAtomicUpdateSerializesCounters(t *testing.T) {
	c := Cache[int]{Name: "atomic-counter"}
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.AtomicUpdate("key", func(value int, found bool) (int, time.Duration, error) {
				return value + 1, time.Minute, nil
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	value, found, err := c.Get("key")
	if err != nil || !found || value != 100 {
		t.Fatalf("counter=%d found=%v error=%v", value, found, err)
	}
}

func TestAtomicUpdateErrorPreservesExistingEntry(t *testing.T) {
	c := Cache[int]{Name: "atomic-error"}
	if err := c.Set("key", 9, time.Minute); err != nil {
		t.Fatal(err)
	}
	rejected := errors.New("rejected")
	if err := c.AtomicUpdate("key", func(value int, found bool) (int, time.Duration, error) {
		return 0, 0, rejected
	}); !errors.Is(err, rejected) {
		t.Fatalf("update error=%v", err)
	}
	value, found, err := c.Get("key")
	if err != nil || !found || value != 9 {
		t.Fatalf("rejected update altered entry: value=%d found=%v error=%v", value, found, err)
	}
}
