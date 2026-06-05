package closer

import (
	"errors"
	"reflect"
	"testing"
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
	mu.Unlock()
}

func TestCloseAllRunsCallbacksInReverseOrderAndClears(t *testing.T) {
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
