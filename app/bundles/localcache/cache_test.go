package localcache

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestCache_GetOrLoad(t *testing.T) {
	c := Cache[string]{}
	defer stopTestCache(&c)
	loads := 0
	a, _ := c.GetOrLoadE("", func() (string, error) {
		loads++
		return "a", nil
	}, time.Minute)
	b := c.GetOrLoad("", func() (string, error) {
		loads++
		return "b", nil
	}, time.Minute)

	if a != "a" || b != "a" || loads != 1 {
		t.Fatalf("expected cached value after first load, got %q/%q with %d loads", a, b, loads)
	}
}

func TestCache_LimitsEntries(t *testing.T) {
	c := Cache[int]{}
	defer stopTestCache(&c)
	for i := 0; i < int(defaultMaxEntries)+10; i++ {
		key := "key-" + strconv.Itoa(i)
		_, _ = c.GetOrLoadE(key, func() (int, error) {
			return i, nil
		}, time.Minute)
	}

	count := c.cache.Len()
	if count > int(defaultMaxEntries) {
		t.Fatalf("cache retained too many entries: %d", count)
	}
}

func TestCache_UsesCustomMaxEntries(t *testing.T) {
	c := Cache[int]{MaxEntries: 3}
	defer stopTestCache(&c)
	for i := 0; i < 10; i++ {
		key := "key-" + strconv.Itoa(i)
		_, _ = c.GetOrLoadE(key, func() (int, error) {
			return i, nil
		}, time.Minute)
	}

	count := c.cache.Len()
	if count > int(c.MaxEntries) {
		t.Fatalf("cache retained too many entries: %d", count)
	}
}

func TestCache_ExpiresEntries(t *testing.T) {
	c := Cache[string]{}
	defer stopTestCache(&c)

	loads := 0
	_, _ = c.GetOrLoadE("key", func() (string, error) {
		loads++
		return "a", nil
	}, 20*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	got, _ := c.GetOrLoadE("key", func() (string, error) {
		loads++
		return "b", nil
	}, time.Minute)

	if got != "b" || loads != 2 {
		t.Fatalf("expected expired entry to reload, got %q with %d loads", got, loads)
	}
}

func TestCache_GetOrLoadE_ReturnsLoaderError(t *testing.T) {
	c := Cache[string]{}
	defer stopTestCache(&c)

	want := errors.New("boom")
	_, err := c.GetOrLoadE("key", func() (string, error) {
		return "", want
	}, time.Minute)
	if !errors.Is(err, want) {
		t.Fatalf("GetOrLoadE() error = %v, want %v", err, want)
	}
}

func TestCache_Set(t *testing.T) {
	c := Cache[string]{}
	defer stopTestCache(&c)

	c.Set("key", "preset", time.Minute)
	got, _ := c.GetOrLoadE("key", func() (string, error) {
		return "loaded", nil
	}, time.Minute)

	if got != "preset" {
		t.Fatalf("Set() value = %q", got)
	}
}

func TestCache_UpdateIfPresent(t *testing.T) {
	c := Cache[int]{}
	defer stopTestCache(&c)

	updated := c.UpdateIfPresent("missing", func(value int) int {
		return value + 1
	}, time.Minute)
	if updated {
		t.Fatal("UpdateIfPresent() updated a missing key")
	}

	c.Set("key", 1, time.Minute)
	updated = c.UpdateIfPresent("key", func(value int) int {
		return value + 1
	}, time.Minute)
	if !updated {
		t.Fatal("UpdateIfPresent() did not update an existing key")
	}

	got, _ := c.GetOrLoadE("key", func() (int, error) {
		return 0, nil
	}, time.Minute)
	if got != 2 {
		t.Fatalf("UpdateIfPresent() value = %d, want 2", got)
	}
}

func TestCache_DeleteAndClear(t *testing.T) {
	c := Cache[string]{}
	defer stopTestCache(&c)

	c.Set("first", "a", time.Minute)
	c.Set("second", "b", time.Minute)
	c.Delete("first")

	got := c.GetOrLoad("first", func() (string, error) {
		return "loaded", nil
	}, time.Minute)
	if got != "loaded" {
		t.Fatalf("Delete() did not remove key, got %q", got)
	}

	c.Clear()
	got = c.GetOrLoad("second", func() (string, error) {
		return "after-clear", nil
	}, time.Minute)
	if got != "after-clear" {
		t.Fatalf("Clear() did not remove entries, got %q", got)
	}
}

func stopTestCache[V any](c *Cache[V]) {
	if c.cache != nil {
		c.cache.Stop()
	}
}
