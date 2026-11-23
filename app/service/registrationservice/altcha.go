package registrationservice

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/netip"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
)

// Keys and redeemed nonces share a process lifetime, including across restarts.
var altchaSigningKey = rand.Text()
var altchaSolutionKey = rand.Text()
var altchaIssueLimits = limiter{entries: make(map[string]counter)}
var altchaRedemptions = limiter{entries: make(map[string]counter)}

const altchaLifetime = 3 * time.Minute

func NewAltchaChallenge(action, ip string) (altcha.Challenge, error) {
	if action != "login" && action != "register" {
		return altcha.Challenge{}, errors.New("invalid challenge action")
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return altcha.Challenge{}, ErrRateLimited
	}
	now := time.Now()
	if err := altchaIssueLimits.consume(now, []quota{{"ip:" + addr.Unmap().String(), 20, time.Minute}, {"global", 600, time.Minute}}); err != nil {
		return altcha.Challenge{}, err
	}
	offset, err := rand.Int(rand.Reader, big.NewInt(160))
	if err != nil {
		return altcha.Challenge{}, err
	}
	counter := 80 + int(offset.Int64())
	expires := now.Add(altchaLifetime)
	return altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm: "PBKDF2/SHA-256", Cost: 1000, Counter: &counter,
		DeriveKey: altcha.DeriveKeyPBKDF2(), ExpiresAt: &expires,
		Data:                map[string]any{"action": action},
		HMACSignatureSecret: altchaSigningKey, HMACKeySignatureSecret: altchaSolutionKey,
	})
}

func VerifyAltcha(token, action string) bool {
	if token == "" || len(token) > 16384 {
		return false
	}
	raw, err := base64.StdEncoding.DecodeString(token)
	var payload altcha.Payload
	if err != nil || json.Unmarshal(raw, &payload) != nil {
		return false
	}
	params := payload.Challenge.Parameters
	if params.Data["action"] != action || params.KeySignature == "" {
		return false
	}
	result, err := altcha.VerifySolution(altcha.VerifySolutionOptions{
		Challenge: payload.Challenge, Solution: payload.Solution,
		HMACSignatureSecret: altchaSigningKey, HMACKeySignatureSecret: altchaSolutionKey,
	})
	if err != nil || !result.Verified {
		return false
	}
	return altchaRedemptions.consume(time.Now(), []quota{{params.Nonce, 1, altchaLifetime}}) == nil
}
