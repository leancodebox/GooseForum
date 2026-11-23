package registrationservice

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func encodeAltchaProof(t *testing.T, challenge altcha.Challenge, solution altcha.Solution) string {
	t.Helper()
	raw, err := json.Marshal(altcha.Payload{Challenge: challenge, Solution: solution})
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func TestAltchaRedemption(t *testing.T) {
	challenge, err := NewAltchaChallenge("login", "192.0.2.1")
	if err != nil {
		t.Fatal(err)
	}
	solution, err := altcha.SolveChallenge(altcha.SolveChallengeOptions{Challenge: challenge, DeriveKey: altcha.DeriveKeyPBKDF2()})
	if err != nil || solution == nil {
		t.Fatalf("solve: %v", err)
	}
	token := encodeAltchaProof(t, challenge, *solution)
	if VerifyAltcha(token, "register") {
		t.Fatal("accepted a login proof for registration")
	}
	if VerifyAltcha(encodeAltchaProof(t, challenge, altcha.Solution{}), "login") {
		t.Fatal("accepted an empty solution")
	}
	tampered := challenge
	tampered.Parameters.ExpiresAt++
	if VerifyAltcha(encodeAltchaProof(t, tampered, *solution), "login") {
		t.Fatal("accepted modified expiration")
	}
	var wg sync.WaitGroup
	var accepted atomic.Int32
	for range 20 {
		wg.Go(func() {
			if VerifyAltcha(token, "login") {
				accepted.Add(1)
			}
		})
	}
	wg.Wait()
	if accepted.Load() != 1 {
		t.Fatalf("redeemed %d times", accepted.Load())
	}
}

func TestAltchaExpired(t *testing.T) {
	signing, key, err := altchaKeys()
	if err != nil {
		t.Fatal(err)
	}
	then := time.Now().Add(-time.Minute)
	counter := 1
	challenge, err := altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm: "PBKDF2/SHA-256", Cost: 1, Counter: &counter, DeriveKey: altcha.DeriveKeyPBKDF2(),
		Data: map[string]any{"action": "register"}, ExpiresAt: &then,
		HMACSignatureSecret: signing, HMACKeySignatureSecret: key,
	})
	if err != nil {
		t.Fatal(err)
	}
	solution, err := altcha.SolveChallenge(altcha.SolveChallengeOptions{Challenge: challenge, DeriveKey: altcha.DeriveKeyPBKDF2()})
	if err != nil || solution == nil {
		t.Fatalf("solve: %v", err)
	}
	if err := altchaPending.Set(challenge.Parameters.Nonce, true, time.Minute); err != nil {
		t.Fatal(err)
	}
	if VerifyAltcha(encodeAltchaProof(t, challenge, *solution), "register") {
		t.Fatal("accepted expired challenge")
	}
}

func TestAltchaIssuanceHasSeparateQuota(t *testing.T) {
	ip := "192.0.2.92"
	t.Cleanup(func() {
		altchaIssueLimits.mu.Lock()
		delete(altchaIssueLimits.entries, "ip:"+ip)
		altchaIssueLimits.mu.Unlock()
		loginLimits.mu.Lock()
		delete(loginLimits.entries, "minute:"+ip)
		delete(loginLimits.entries, "hour:"+ip)
		loginLimits.mu.Unlock()
	})
	for range 20 {
		if _, err := NewAltchaChallenge("register", ip); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewAltchaChallenge("login", "::ffff:"+ip); !errors.Is(err, ErrRateLimited) {
		t.Fatal("missing challenge limit")
	}
	if err := AllowPasswordLogin(ip); err != nil {
		t.Fatal("issuance exhausted login budget")
	}
}

func TestAltchaRequiresSiteKeyAndPendingTicket(t *testing.T) {
	old := preferences.Get("app.secretKey")
	t.Cleanup(func() { preferences.Set("app.secretKey", old) })
	preferences.Set("app.secretKey", "")
	if _, err := NewAltchaChallenge("login", "192.0.2.55"); err == nil {
		t.Fatal("issued without a site key")
	}

	preferences.Set("app.secretKey", "site A")
	signing, solutionKey, err := altchaKeys()
	if err != nil || signing == solutionKey {
		t.Fatal("missing distinct purpose keys")
	}
	challenge, err := NewAltchaChallenge("register", "192.0.2.55")
	if err != nil {
		t.Fatal(err)
	}
	solution, err := altcha.SolveChallenge(altcha.SolveChallengeOptions{Challenge: challenge, DeriveKey: altcha.DeriveKeyPBKDF2()})
	if err != nil || solution == nil {
		t.Fatalf("solve: %v", err)
	}
	token := encodeAltchaProof(t, challenge, *solution)
	preferences.Set("app.secretKey", "site B")
	if VerifyAltcha(token, "register") {
		t.Fatal("accepted a ticket signed with a different key")
	}
	preferences.Set("app.secretKey", "site A")
	if pending, found, err := altchaPending.Take(challenge.Parameters.Nonce); err != nil || !found || !pending {
		t.Fatal("invalid signature consumed the pending ticket")
	}
	if VerifyAltcha(token, "register") {
		t.Fatal("accepted a ticket after cache loss")
	}
}
