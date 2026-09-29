package sharedcache

import (
	"errors"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func TestMemoryIsolationAndInvalidation(t *testing.T) {
	a := Cache[int]{Name: "first", MaxEntries: 3}
	b := Cache[int]{Name: "second", MaxEntries: 3}
	a.Set("user:1:audience:a", 1, time.Minute)
	a.Set("user:1:audience:b", 2, time.Minute)
	a.Set("user:10:audience:a", 10, time.Minute)
	b.Set("user:1:audience:a", 20, time.Minute)
	if err := a.DeletePrefix("user:1:audience:"); err != nil {
		t.Fatal(err)
	}
	load := func() (int, error) { return 99, nil }
	for _, key := range []string{"user:1:audience:a", "user:1:audience:b"} {
		if got := a.GetOrLoad(key, load, time.Minute); got != 99 {
			t.Fatalf("prefix invalidation retained %s: %d", key, got)
		}
	}
	if got := a.GetOrLoad("user:10:audience:a", load, time.Minute); got != 10 {
		t.Fatalf("prefix invalidation affected another user: %d", got)
	}
	a.Clear()
	if got := b.GetOrLoad("user:1:audience:a", load, time.Minute); got != 20 {
		t.Fatalf("clear affected another namespace: %d", got)
	}
	if got := a.GetOrLoad("user:10:audience:a", load, time.Minute); got != 99 {
		t.Fatalf("clear retained entry: %d", got)
	}
}

func TestInvalidationDuringLoadPreventsRefill(t *testing.T) {
	for _, operation := range []string{"delete", "prefix", "clear", "set", "missing-update"} {
		t.Run(operation, func(t *testing.T) {
			c := Cache[int]{Name: operation}
			started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
			go func() {
				defer close(done)
				c.GetOrLoad("user:1", func() (int, error) {
					close(started)
					<-release
					return 1, nil
				}, time.Minute)
			}()
			<-started
			want := 2
			switch operation {
			case "delete":
				c.Delete("user:1")
			case "prefix":
				c.DeletePrefix("user:")
			case "clear":
				c.Clear()
			case "set":
				want = 3
				c.Set("user:1", want, time.Minute)
			case "missing-update":
				if c.UpdateIfPresent("user:1", func(v int) int { return v + 1 }, time.Minute) {
					t.Fatal("updated a missing entry")
				}
			}
			close(release)
			<-done
			if got := c.GetOrLoad("user:1", func() (int, error) { return 2, nil }, time.Minute); got != want {
				t.Fatalf("old load overwrote invalidation/write: %d, want %d", got, want)
			}
		})
	}
}

func TestConcurrentLoadsAndLoaderError(t *testing.T) {
	c := Cache[int]{Name: "concurrent"}
	var loads atomic.Int64
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			value, err := c.GetOrLoadE("key", func() (int, error) {
				loads.Add(1)
				return 7, nil
			}, time.Minute)
			if err != nil || value != 7 {
				t.Errorf("load = %d, %v", value, err)
			}
		})
	}
	wg.Wait()
	if loads.Load() != 1 {
		t.Fatalf("loader called %d times", loads.Load())
	}
	want := errors.New("loader failed")
	if _, err := c.GetOrLoadE("error", func() (int, error) { return 0, want }, time.Minute); !errors.Is(err, want) {
		t.Fatalf("loader error = %v", err)
	}
	if got := c.GetOrLoad("error", func() (int, error) { return 8, nil }, time.Minute); got != 8 {
		t.Fatalf("loader failure was cached: %d", got)
	}
}

func TestNewRequestAfterInvalidationDoesNotJoinOldLoad(t *testing.T) {
	for _, operation := range []string{"delete", "prefix", "clear", "set", "missing-update"} {
		t.Run(operation, func(t *testing.T) {
			c := Cache[int]{Name: "flight-invalidation"}
			started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
			go func() {
				defer close(done)
				c.GetOrLoad("key", func() (int, error) {
					close(started)
					<-release
					return 1, nil
				}, time.Minute)
			}()
			<-started
			switch operation {
			case "delete":
				c.Delete("key")
			case "prefix":
				c.DeletePrefix("ke")
			case "clear":
				c.Clear()
			case "set":
				c.Set("key", 2, time.Minute)
			case "missing-update":
				if c.UpdateIfPresent("key", func(v int) int { return v + 1 }, time.Minute) {
					t.Fatal("updated a missing entry")
				}
			}
			fresh := make(chan int, 1)
			go func() {
				fresh <- c.GetOrLoad("key", func() (int, error) { return 2, nil }, time.Minute)
			}()
			select {
			case value := <-fresh:
				if value != 2 {
					t.Errorf("new request returned old value: %d", value)
				}
			case <-time.After(time.Second):
				t.Error("new request joined the invalidated load")
			}
			close(release)
			<-done
			if got := c.GetOrLoad("key", func() (int, error) { return 3, nil }, time.Minute); got != 2 {
				t.Fatalf("old load replaced fresh cache entry: %d", got)
			}
		})
	}
}

type failingBackend struct {
	Backend[int]
	readErr  error
	writeErr error
}

func (b *failingBackend) Get(string) (int, bool, error)        { return 0, false, b.readErr }
func (b *failingBackend) Set(string, int, time.Duration) error { return b.writeErr }

func TestInjectedBackendErrors(t *testing.T) {
	want := errors.New("backend unavailable")
	backend := &failingBackend{readErr: want}
	c := NewWithBackend[int]("injected", backend)
	loaded := false
	load := func() (int, error) { loaded = true; return 5, nil }
	if _, err := c.GetOrLoadE("key", load, time.Minute); !errors.Is(err, want) || loaded {
		t.Fatalf("backend read error = %v, loaded = %v", err, loaded)
	}
	backend.readErr, backend.writeErr = nil, want
	if got, err := c.GetOrLoadE("key", load, time.Minute); err != nil || got != 5 {
		t.Fatalf("fill failure discarded loaded data: %d, %v", got, err)
	}
	if err := c.Set("key", 6, time.Minute); !errors.Is(err, want) {
		t.Fatalf("explicit write error = %v", err)
	}
}

func TestMemoryExpirationCapacityAndUpdate(t *testing.T) {
	c := Cache[int]{Name: "bounded", MaxEntries: 2}
	for _, key := range []string{"a", "b", "c"} {
		c.Set(key, 1, time.Minute)
	}
	if got := c.GetOrLoad("a", func() (int, error) { return 9, nil }, time.Minute); got != 9 {
		t.Fatalf("capacity did not evict oldest entry: %d", got)
	}
	if c.UpdateIfPresent("missing", func(v int) int { return v + 1 }, time.Minute) {
		t.Fatal("update created a missing entry")
	}
	if !c.UpdateIfPresent("a", func(v int) int { return v + 1 }, 10*time.Millisecond) {
		t.Fatal("update failed for existing entry")
	}
	if got := c.GetOrLoad("a", func() (int, error) { return 0, nil }, time.Minute); got != 10 {
		t.Fatalf("updated value = %d", got)
	}
	time.Sleep(40 * time.Millisecond)
	if got := c.GetOrLoad("a", func() (int, error) { return 11, nil }, time.Minute); got != 11 {
		t.Fatalf("expired entry did not reload: %d", got)
	}
}

func TestConfigurationIsFrozenBeforeLazyCacheInitialization(t *testing.T) {
	original := preferences.GetString("cache.driver", "memory")
	defer preferences.Set("cache.driver", original)
	originalSelection := selection
	selection = &backendSelection{driver: "memory"}
	defer func() { selection = originalSelection }()
	preferences.Set("cache.driver", "redis")
	if err := ConfigureFromPreferences(); err == nil {
		t.Fatal("unimplemented redis driver was accepted")
	}
	preferences.Set("cache.driver", "memory")
	if err := ConfigureFromPreferences(); err != nil {
		t.Fatal(err)
	}
	preferences.Set("cache.driver", "redis")
	if err := ConfigureFromPreferences(); err != nil {
		t.Fatalf("running process reread backend configuration: %v", err)
	}
	c := Cache[int]{Name: "lazy-after-config-change"}
	for range 2 {
		if value, err := c.GetOrLoadE("key", func() (int, error) { return 7, nil }, time.Minute); err != nil || value != 7 {
			t.Fatalf("lazy cache initialization after config change = %d, %v", value, err)
		}
	}
	// A new process must validate the changed setting again.
	selection = &backendSelection{driver: "memory"}
	if err := ConfigureFromPreferences(); err == nil {
		t.Fatal("new startup accepted an unsupported driver")
	}
}

func TestInitializationFailureCanBeRetried(t *testing.T) {
	c := Cache[int]{}
	load := func() (int, error) { return 4, nil }
	for range 2 {
		if _, err := c.GetOrLoadE("key", load, time.Minute); err == nil {
			t.Fatal("unnamed cache initialized")
		}
	}
	if err := c.Set("key", 4, time.Minute); err == nil {
		t.Fatal("write swallowed initialization error")
	}
	if err := c.Delete("key"); err == nil {
		t.Fatal("delete swallowed initialization error")
	}
	if err := c.DeletePrefix("ke"); err == nil {
		t.Fatal("prefix deletion swallowed initialization error")
	}
	if err := c.Clear(); err == nil {
		t.Fatal("clear swallowed initialization error")
	}
	if c.UpdateIfPresent("key", func(v int) int { return v }, time.Minute) {
		t.Fatal("update succeeded after initialization error")
	}
	c.Name = "repaired"
	if value, err := c.GetOrLoadE("key", load, time.Minute); err != nil || value != 4 {
		t.Fatalf("initialization retry = %d, %v", value, err)
	}
}

// Wait for callers to attach to an active flight without timing the loader.
func waitFlightUsers(t *testing.T, c *Cache[int], key string, users int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		pending := c.flights[key]
		found := pending != nil && pending.users == users
		c.mu.Unlock()
		if found {
			return
		}
		runtime.Gosched()
	}
	t.Fatalf("flight %q did not reach %d callers", key, users)
}

func TestUnrelatedMutationPreservesFlightAndFill(t *testing.T) {
	for _, operation := range []string{"set", "delete", "prefix", "update", "missing-update"} {
		t.Run(operation, func(t *testing.T) {
			c := Cache[int]{Name: "unrelated"}
			c.Set("other", 1, time.Minute)
			started, release := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			t.Cleanup(unblock)
			var loads atomic.Int64
			load := func() (int, error) {
				if loads.Add(1) == 1 {
					close(started)
				}
				<-release
				return 7, nil
			}
			results := make(chan int, 2)
			go func() { results <- c.GetOrLoad("target", load, time.Minute) }()
			<-started
			c.mu.Lock()
			originalFlight := c.flights["target"]
			c.mu.Unlock()
			switch operation {
			case "set":
				c.Set("other", 2, time.Minute)
			case "delete":
				c.Delete("other")
			case "prefix":
				c.DeletePrefix("other")
			case "update":
				c.UpdateIfPresent("other", func(v int) int { return v + 1 }, time.Minute)
			case "missing-update":
				if c.UpdateIfPresent("missing", func(v int) int { return v + 1 }, time.Minute) {
					t.Fatal("updated a missing entry")
				}
			}
			c.mu.Lock()
			unchanged := c.flights["target"] == originalFlight
			c.mu.Unlock()
			if !unchanged {
				t.Fatal("unrelated mutation detached active flight")
			}
			go func() { results <- c.GetOrLoad("target", load, time.Minute) }()
			waitFlightUsers(t, &c, "target", 2)
			unblock()
			for range 2 {
				if value := <-results; value != 7 {
					t.Fatalf("load returned %d", value)
				}
			}
			if value := c.GetOrLoad("target", load, time.Minute); value != 7 || loads.Load() != 1 {
				t.Fatalf("flight/fill lost: value=%d, loads=%d", value, loads.Load())
			}
			c.mu.Lock()
			remaining := len(c.flights)
			c.mu.Unlock()
			if remaining != 0 {
				t.Fatalf("retained %d completed flight records", remaining)
			}
		})
	}
}

func TestOldCompletionPreservesReplacementFlight(t *testing.T) {
	c := Cache[int]{Name: "replacement"}
	oldStarted, oldRelease, oldDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	newStarted, newRelease := make(chan struct{}), make(chan struct{})
	var oldOnce, newOnce sync.Once
	unblockOld := func() { oldOnce.Do(func() { close(oldRelease) }) }
	unblockNew := func() { newOnce.Do(func() { close(newRelease) }) }
	t.Cleanup(unblockOld)
	t.Cleanup(unblockNew)
	go func() {
		defer close(oldDone)
		c.GetOrLoad("key", func() (int, error) {
			close(oldStarted)
			<-oldRelease
			return 1, nil
		}, time.Minute)
	}()
	<-oldStarted
	c.Delete("key")
	var loads atomic.Int64
	load := func() (int, error) {
		if loads.Add(1) == 1 {
			close(newStarted)
		}
		<-newRelease
		return 2, nil
	}
	results := make(chan int, 2)
	go func() { results <- c.GetOrLoad("key", load, time.Minute) }()
	<-newStarted
	unblockOld()
	<-oldDone
	go func() { results <- c.GetOrLoad("key", load, time.Minute) }()
	waitFlightUsers(t, &c, "key", 2)
	unblockNew()
	for range 2 {
		if value := <-results; value != 2 {
			t.Fatalf("replacement returned %d", value)
		}
	}
	if value := c.GetOrLoad("key", load, time.Minute); value != 2 || loads.Load() != 1 {
		t.Fatalf("replacement flight lost: value=%d, loads=%d", value, loads.Load())
	}
}

func TestNilInterfaceValueCanBeCached(t *testing.T) {
	c := Cache[any]{Name: "nil-value"}
	loads := 0
	for range 2 {
		value, err := c.GetOrLoadE("key", func() (any, error) { loads++; return nil, nil }, time.Minute)
		if value != nil || err != nil {
			t.Fatalf("nil load = %v, %v", value, err)
		}
	}
	if loads != 1 {
		t.Fatalf("nil value loaded %d times", loads)
	}
}

func TestDifferentKeysLoadConcurrently(t *testing.T) {
	c := Cache[int]{Name: "different-keys"}
	started, release, done := make(chan string, 2), make(chan struct{}), make(chan struct{}, 2)
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	for _, key := range []string{"a", "b"} {
		go func() {
			defer func() { done <- struct{}{} }()
			c.GetOrLoad(key, func() (int, error) {
				started <- key
				<-release
				return 1, nil
			}, time.Minute)
		}()
	}
	for range 2 {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatal("one key's loader blocked another key")
		}
	}
	unblock()
	for range 2 {
		<-done
	}
}

func TestLoaderPanicDoesNotPoisonFlight(t *testing.T) {
	c := Cache[int]{Name: "panic-retry"}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("loader panic was not propagated")
			}
		}()
		c.GetOrLoad("key", func() (int, error) { panic("loader panic") }, time.Minute)
	}()
	c.mu.Lock()
	remaining := len(c.flights)
	c.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("panic retained %d flight records", remaining)
	}
	if got := c.GetOrLoad("key", func() (int, error) { return 8, nil }, time.Minute); got != 8 {
		t.Fatalf("load after panic returned %d", got)
	}
}

func TestConcurrentReadsAndInvalidationsConverge(t *testing.T) {
	const keyCount = 64
	c := Cache[uint64]{Name: "invalidation-stress", MaxEntries: keyCount}
	var source [keyCount]atomic.Uint64
	keys := make([]string, keyCount)
	for i := range keys {
		keys[i] = "user:" + strconv.Itoa(i) + ":profile"
		source[i].Store(1)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := range 32 {
		wg.Go(func() {
			<-start
			for round := range 500 {
				i := (round + worker) % keyCount
				value, err := c.GetOrLoadE(keys[i], func() (uint64, error) {
					value := source[i].Load()
					// Allow invalidation to interleave after the source was read.
					runtime.Gosched()
					return value, nil
				}, time.Minute)
				if err != nil || value == 0 {
					t.Errorf("concurrent read = %d, %v", value, err)
					return
				}
			}
		})
	}
	wg.Go(func() {
		<-start
		for round := range 1000 {
			i := round % keyCount
			source[i].Add(1)
			var err error
			switch round % 3 {
			case 0:
				err = c.Delete(keys[i])
			case 1:
				err = c.DeletePrefix("user:" + strconv.Itoa(i) + ":")
			case 2:
				err = c.Clear()
			}
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	close(start)
	wg.Wait()
	for i, key := range keys {
		value, err := c.GetOrLoadE(key, func() (uint64, error) { return source[i].Load(), nil }, time.Minute)
		if err != nil || value != source[i].Load() {
			t.Fatalf("stale refill after writes settled for %q: got %d, want %d, err %v", key, value, source[i].Load(), err)
		}
	}
	c.mu.Lock()
	remaining := len(c.flights)
	c.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("stress run retained %d flight records", remaining)
	}
}
