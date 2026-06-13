package closer

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testCloser struct {
	closed *bool
}

func (c testCloser) Close() error {
	*c.closed = true
	return nil
}

func resetCloserForTest(t *testing.T) {
	t.Helper()
	mu.Lock()
	entries = nil
	nextSeq = 0
	mu.Unlock()
	closeTimeout = 10 * time.Second
}

func TestCloseAllRunsCallbacksInReverseOrderWithinPriorityAndClears(t *testing.T) {
	resetCloserForTest(t)
	t.Cleanup(func() {
		resetCloserForTest(t)
	})

	var order []int
	Register(func() error {
		order = append(order, 1)
		return nil
	})
	Register(func() error {
		order = append(order, 2)
		return errors.New("ignored")
	})

	CloseAll()

	if want := []int{2, 1}; !reflect.DeepEqual(order, want) {
		t.Fatalf("close order = %#v, want %#v", order, want)
	}

	order = nil
	CloseAll()
	if len(order) != 0 {
		t.Fatalf("CloseAll should clear callbacks after first run")
	}
}

func TestCloseAllRunsCallbacksByPriority(t *testing.T) {
	resetCloserForTest(t)
	t.Cleanup(func() {
		resetCloserForTest(t)
	})

	var order []int
	RegisterPriority(PriorityLogger, func() error {
		order = append(order, 4)
		return nil
	})
	RegisterPriority(PriorityProducer, func() error {
		order = append(order, 1)
		return nil
	})
	RegisterPriority(PriorityFlush, func() error {
		order = append(order, 2)
		return nil
	})
	RegisterPriority(PriorityDatabase, func() error {
		order = append(order, 3)
		return nil
	})

	CloseAll()

	if want := []int{1, 2, 3, 4}; !reflect.DeepEqual(order, want) {
		t.Fatalf("close order = %#v, want %#v", order, want)
	}
}

func TestBindRegistersCloseMethod(t *testing.T) {
	resetCloserForTest(t)
	t.Cleanup(func() {
		resetCloserForTest(t)
	})

	closed := false
	Bind(testCloser{closed: &closed})
	CloseAll()

	if !closed {
		t.Fatalf("bound closer was not closed")
	}
}

func TestCloseAllTimesOutBlockedCallback(t *testing.T) {
	resetCloserForTest(t)
	t.Cleanup(func() {
		resetCloserForTest(t)
	})

	closeTimeout = 20 * time.Millisecond
	Register(func() error {
		select {}
	})

	start := time.Now()
	CloseAll()

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("CloseAll waited too long for blocked callback: %s", elapsed)
	}

	ran := false
	Register(func() error {
		ran = true
		return nil
	})
	CloseAll()
	if !ran {
		t.Fatalf("CloseAll should continue working after a timed out callback")
	}
}
