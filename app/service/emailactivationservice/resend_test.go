package emailactivationservice

import (
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/service/kvstore"
)

func useTempKV(t *testing.T) {
	t.Helper()
	kvstore.Close()
	preferences.Set("badger.path", t.TempDir())
	t.Cleanup(kvstore.Close)
}

func TestReserveCooldownAndDailyLimit(t *testing.T) {
	useTempKV(t)

	userID := uint64(1001)
	now := time.Date(2026, 6, 2, 10, 0, 0, 0, time.Local)
	first, err := consumeQuota(userID, now)
	if err != nil {
		t.Fatalf("consumeQuota() first error = %v", err)
	}
	if first.RemainingToday != 2 {
		t.Fatalf("consumeQuota() first remaining = %d, want 2", first.RemainingToday)
	}

	cooldown, err := consumeQuota(userID, now.Add(30*time.Second))
	if !errors.Is(err, ErrCooldown) {
		t.Fatalf("consumeQuota() cooldown error = %v, want ErrCooldown", err)
	}
	if cooldown.RetryAfterSeconds <= 0 {
		t.Fatalf("consumeQuota() retryAfterSeconds = %d, want positive", cooldown.RetryAfterSeconds)
	}

	second, err := consumeQuota(userID, now.Add(61*time.Second))
	if err != nil {
		t.Fatalf("consumeQuota() second send error = %v", err)
	}
	if second.RemainingToday != 1 {
		t.Fatalf("consumeQuota() second remaining = %d, want 1", second.RemainingToday)
	}
	third, err := consumeQuota(userID, now.Add(122*time.Second))
	if err != nil {
		t.Fatalf("consumeQuota() third send error = %v", err)
	}
	if third.RemainingToday != 0 {
		t.Fatalf("consumeQuota() third remaining = %d, want 0", third.RemainingToday)
	}
	if _, err = consumeQuota(userID, now.Add(183*time.Second)); !errors.Is(err, ErrDailyLimit) {
		t.Fatalf("consumeQuota() daily limit error = %v, want ErrDailyLimit", err)
	}
}
