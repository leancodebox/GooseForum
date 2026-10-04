package registrationservice

import (
	"errors"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"sync"
	"testing"
	"time"
)

func TestDomainPolicyUsesNormalizedExactMatch(t *testing.T) {
	config := pageConfig.SecurityAndRegistration{AllowedDomains: []string{"EXAMPLE.com", "例子.中国"}}
	for _, email := range []string{"Alice@EXAMPLE.COM", "user@例子.中国"} {
		if _, err := ValidateEmail(email, config); err != nil {
			t.Fatalf("%s: %v", email, err)
		}
	}
	for _, email := range []string{"user@sub.example.com", "user@evilexample.com", "user@example.com.evil", "display <user@example.com>", "", "user@example.com."} {
		if _, err := ValidateEmail(email, config); err == nil {
			t.Fatalf("allowed %q", email)
		}
	}
}

func TestLimiterCountsConcurrentAttemptsAndExpires(t *testing.T) {
	l := limiter{entries: make(map[string]counter)}
	now := time.Unix(10000, 0)
	var mu sync.Mutex
	accepted := 0
	var wg sync.WaitGroup
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.consume(now, []quota{{"ip:local", 5}, {"global", 100}}) == nil {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if accepted != 5 {
		t.Fatalf("accepted %d requests", accepted)
	}
	if err := l.consume(now.Add(time.Hour), []quota{{"ip:local", 5}}); err != nil {
		t.Fatal(err)
	}
}

func TestLimiterRejectsWithoutConsumingOtherQuotas(t *testing.T) {
	l := limiter{entries: make(map[string]counter)}
	now := time.Now()
	if err := l.consume(now, []quota{{"email:a", 1}}); err != nil {
		t.Fatal(err)
	}
	if err := l.consume(now, []quota{{"global", 3}, {"email:a", 1}}); !errors.Is(err, ErrRateLimited) {
		t.Fatal(err)
	}
	if _, exists := l.entries["global"]; exists {
		t.Fatal("rejected request consumed global quota")
	}
	for range 20 {
		if err := l.consume(now, []quota{{"ip:a", 0}}); err != nil {
			t.Fatal(err)
		}
	}
	if _, exists := l.entries["ip:a"]; exists {
		t.Fatal("disabled quota stored counters")
	}
}

func TestSettingsRejectInvalidLimitsAndDomains(t *testing.T) {
	for _, config := range []pageConfig.SecurityAndRegistration{{RegistrationIPLimit: -1}, {RegistrationGlobalLimit: 100001}, {AllowedDomains: []string{"*.example.com"}}, {AllowedDomains: []string{".example.com"}}, {AllowedDomains: []string{"example..com"}}} {
		if ValidateSettings(&config) == nil {
			t.Fatalf("accepted %+v", config)
		}
	}
}
