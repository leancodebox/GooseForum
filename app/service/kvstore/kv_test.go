package kvstore

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func useTempStore(t *testing.T) {
	t.Helper()
	resetForTest()
	preferences.Set("badger.path", t.TempDir())
	t.Cleanup(Close)
}

func resetForTest() {
	Close()
	current = nil
	connectOnce = sync.OnceValues(connect)
}

func TestSetGetAndCopiesBytes(t *testing.T) {
	useTempStore(t)

	input := []byte("one")
	if err := SetBytes("copy", input, 0); err != nil {
		t.Fatalf("SetBytes() error = %v", err)
	}
	input[0] = 'x'

	got, err := GetBytes("copy")
	if err != nil {
		t.Fatalf("GetBytes() error = %v", err)
	}
	if !bytes.Equal(got, []byte("one")) {
		t.Fatalf("GetBytes() = %q, want one", got)
	}

	got[0] = 'x'
	gotAgain, err := Get("copy")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if gotAgain != "one" {
		t.Fatalf("Get() after caller change = %q, want one", gotAgain)
	}
}

func TestGetManyBytesReturnsExistingKeysAndCopiesValues(t *testing.T) {
	useTempStore(t)

	if err := SetBytes("one", []byte("first"), 0); err != nil {
		t.Fatalf("SetBytes(one) error = %v", err)
	}
	if err := SetBytes("two", []byte("second"), 0); err != nil {
		t.Fatalf("SetBytes(two) error = %v", err)
	}

	got, err := GetManyBytes([]string{"one", "missing", "two", "one"})
	if err != nil {
		t.Fatalf("GetManyBytes() error = %v", err)
	}
	if len(got) != 2 || string(got["one"]) != "first" || string(got["two"]) != "second" {
		t.Fatalf("GetManyBytes() = %#v, want existing values", got)
	}

	got["one"][0] = 'x'
	again, err := Get("one")
	if err != nil {
		t.Fatalf("Get(one) error = %v", err)
	}
	if again != "first" {
		t.Fatalf("stored value after caller mutation = %q, want first", again)
	}
}

func TestGetManyBytesRejectsInvalidKeys(t *testing.T) {
	useTempStore(t)

	_, err := GetManyBytes([]string{"valid", ""})
	if !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("GetManyBytes() error = %v, want ErrInvalidKey", err)
	}
}

func TestUpdateBytesSetKeepAndCopiesCurrent(t *testing.T) {
	useTempStore(t)

	err := UpdateBytes("mutate", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
		if exists {
			t.Fatal("updater exists = true before first write")
		}
		return UpdateSet, []byte("created"), nil
	})
	if err != nil {
		t.Fatalf("UpdateBytes() set error = %v", err)
	}

	var seen []byte
	err = UpdateBytes("mutate", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
		if !exists || !bytes.Equal(current, []byte("created")) {
			t.Fatalf("updater current = %q exists = %v, want created/true", current, exists)
		}
		seen = append([]byte{}, current...)
		current[0] = 'x'
		return UpdateKeep, nil, nil
	})
	if err != nil {
		t.Fatalf("UpdateBytes() keep error = %v", err)
	}
	if !bytes.Equal(seen, []byte("created")) {
		t.Fatalf("updater saw current = %q, want created", seen)
	}

	err = UpdateBytes("mutate", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
		if !exists || !bytes.Equal(current, []byte("created")) {
			t.Fatalf("updater current after caller change = %q exists = %v, want created/true", current, exists)
		}
		return UpdateKeep, nil, nil
	})
	if err != nil {
		t.Fatalf("UpdateBytes() second keep error = %v", err)
	}
}

func TestUpdateBytesTTLExpiry(t *testing.T) {
	useTempStore(t)

	ttl := 1100 * time.Millisecond
	if err := UpdateBytes("ttl", ttl, func(current []byte, exists bool) (UpdateAction, []byte, error) {
		if exists {
			t.Fatal("updater exists = true before first write")
		}
		return UpdateSet, []byte("value"), nil
	}); err != nil {
		t.Fatalf("UpdateBytes() set ttl error = %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		expired := false
		err := UpdateBytes("ttl", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
			expired = !exists
			return UpdateKeep, nil, nil
		})
		if err != nil {
			t.Fatalf("UpdateBytes() ttl read error = %v", err)
		}
		if expired {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("UpdateBytes() did not observe ttl expiry")
}

func TestClosePreventsReconnect(t *testing.T) {
	useTempStore(t)

	if err := UpdateBytes("persist", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
		return UpdateSet, []byte("value"), nil
	}); err != nil {
		t.Fatalf("UpdateBytes() error = %v", err)
	}
	Close()
	if err := Set("after-close", "value", 0); !errors.Is(err, ErrClosed) {
		t.Fatalf("Set() after Close() error = %v, want ErrClosed", err)
	}
}

func TestCloseBeforeConnectDoesNotOpenStore(t *testing.T) {
	resetForTest()
	preferences.Set("badger.path", t.TempDir())

	Close()
	currentMu.RLock()
	got := current
	currentMu.RUnlock()
	if got != nil {
		t.Fatal("Close() before Connect opened store")
	}
}

func TestConcurrentUpdateBytes(t *testing.T) {
	useTempStore(t)

	const workers = 16
	const increments = 50

	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for range workers {
		wg.Go(func() {
			for range increments {
				err := UpdateBytes("concurrent", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
					var value uint64
					if exists && len(current) >= 8 {
						value = binary.BigEndian.Uint64(current)
					}
					value++
					next := make([]byte, 8)
					binary.BigEndian.PutUint64(next, value)
					return UpdateSet, next, nil
				})
				if err != nil {
					errs <- err
					return
				}
			}
		})
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("UpdateBytes() concurrent error = %v", err)
		}
	}

	var got uint64
	err := UpdateBytes("concurrent", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
		if exists && len(current) >= 8 {
			got = binary.BigEndian.Uint64(current)
		}
		return UpdateKeep, nil, nil
	})
	if err != nil {
		t.Fatalf("UpdateBytes() final read error = %v", err)
	}

	if got != workers*increments {
		t.Fatalf("counter = %d, want %d", got, workers*increments)
	}
}

func TestInvalidKey(t *testing.T) {
	useTempStore(t)

	err := UpdateBytes("", 0, func(current []byte, exists bool) (UpdateAction, []byte, error) {
		return UpdateSet, []byte("value"), nil
	})
	if !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("UpdateBytes() error = %v, want ErrInvalidKey", err)
	}
}
