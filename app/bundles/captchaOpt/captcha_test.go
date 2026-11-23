package captchaOpt

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestImageAnswerConsumedOnce(t *testing.T) {
	store := customStore{&captchaStore{data: make(map[string]captchaInfo)}}
	if err := store.Set("challenge", "123456"); err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for range 32 {
		wg.Go(func() {
			if store.Verify("challenge", "123456", true) {
				successes.Add(1)
			}
		})
	}
	wg.Wait()
	if successes.Load() != 1 || store.Verify("challenge", "", true) {
		t.Fatal("invalid consumption of image challenge")
	}
}
