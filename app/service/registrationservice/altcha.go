package registrationservice

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/netip"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
)

var altchaIssueLimits = limiter{entries: make(map[string]counter)}
var altchaPending = sharedcache.Cache[bool]{Name: "altcha-pending"}

const altchaLifetime = 3 * time.Minute

func altchaKeys() (signing, solution string, err error) {
	secret := preferences.SecretKey()
	if secret == "" {
		return "", "", errors.New("ALTCHA requires a site secret")
	}
	derive := func(label string) string {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(label))
		return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	}
	return derive("gooseforum/altcha/challenge/v1"), derive("gooseforum/altcha/solution/v1"), nil
}

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
	signing, solution, err := altchaKeys()
	if err != nil {
		return altcha.Challenge{}, err
	}
	offset, err := rand.Int(rand.Reader, big.NewInt(160))
	if err != nil {
		return altcha.Challenge{}, err
	}
	counter := 80 + int(offset.Int64())
	expires := now.Add(altchaLifetime)
	challenge, err := altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm: "PBKDF2/SHA-256", Cost: 1000, Counter: &counter,
		DeriveKey: altcha.DeriveKeyPBKDF2(), ExpiresAt: &expires,
		Data:                map[string]any{"action": action},
		HMACSignatureSecret: signing, HMACKeySignatureSecret: solution,
	})
	if err != nil {
		return altcha.Challenge{}, err
	}
	ttl := time.Until(time.Unix(challenge.Parameters.ExpiresAt, 0))
	if ttl <= 0 {
		return altcha.Challenge{}, errors.New("ALTCHA challenge expired during issuance")
	}
	if err := altchaPending.Set(challenge.Parameters.Nonce, true, ttl); err != nil {
		return altcha.Challenge{}, err
	}
	return challenge, nil
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
	signing, solution, err := altchaKeys()
	if err != nil {
		return false
	}
	result, err := altcha.VerifySolution(altcha.VerifySolutionOptions{
		Challenge: payload.Challenge, Solution: payload.Solution,
		HMACSignatureSecret: signing, HMACKeySignatureSecret: solution,
	})
	if err != nil || !result.Verified {
		return false
	}
	_, found, err := altchaPending.Take(params.Nonce)
	return err == nil && found
}
