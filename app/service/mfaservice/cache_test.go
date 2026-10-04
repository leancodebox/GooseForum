package mfaservice

import (
	"net/http"
	"testing"
	"time"
)

func TestBudgetUpdatesPreserveDeadline(t *testing.T) {
	if err := budgets.Clear(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = budgets.Clear() })
	expires := time.Now().Add(time.Minute)
	if err := budgets.Set("user:12345", budget{Attempts: 1, Expires: expires}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if !allow(12345, "192.0.2.100") {
		t.Fatal("allowed attempt rejected")
	}
	item, found, err := budgets.Get("user:12345")
	if err != nil || !found || !item.Expires.Equal(expires) || item.Attempts != 2 {
		t.Fatalf("counter or deadline changed: %+v found=%v error=%v", item, found, err)
	}
}

func TestBusyChallengeCannotBeConsumedConcurrently(t *testing.T) {
	raw := "12345:busy-regression"
	c, _ := testContext(&http.Cookie{Name: challengeCookie, Value: raw})
	item := challenge{UserID: 12345, IP: c.ClientIP(), UA: c.Request.UserAgent(), Attempts: 1, Busy: true, Expires: time.Now().Add(time.Minute)}
	if err := challenges.Set(raw, item, time.Minute); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = challenges.Delete(raw) })
	if _, err := Login(c, "123456"); err == nil {
		t.Fatal("busy challenge accepted")
	}
	after, found, err := challenges.Get(raw)
	if err != nil || !found || !after.Busy || after.Attempts != item.Attempts || !after.Expires.Equal(item.Expires) {
		t.Fatalf("rejected concurrent request changed challenge: %+v found=%v error=%v", after, found, err)
	}
}
