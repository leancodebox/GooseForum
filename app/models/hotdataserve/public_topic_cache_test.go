package hotdataserve

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestPublicTopicMaxCacheCoalescesAndInvalidates(t *testing.T) {
	ClearPublicTopicMaxCache()
	t.Cleanup(ClearPublicTopicMaxCache)
	var calls atomic.Int32
	load := func() (uint64, error) { calls.Add(1); return 42, nil }
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := loadPublicTopicMax(load)
			if id != 42 || err != nil {
				t.Errorf("id=%d err=%v", id, err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("loads=%d", calls.Load())
	}
	ClearTopicListCache()
	if _, err := loadPublicTopicMax(load); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("invalidation did not reload: %d", calls.Load())
	}
}
