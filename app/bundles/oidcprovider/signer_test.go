package oidcprovider

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestRSAKeySetSignsAndPublishesVerifiableKeys(t *testing.T) {
	keys, err := GenerateRSAKeySet("key-1", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if keys.publicKeys["key-1"].N == keys.activeKey.N {
		t.Fatal("public key aliases private key material")
	}
	token, err := keys.SignIDToken(context.Background(), IDTokenClaims{Issuer: "https://issuer.example", Subject: "user-1", Audience: "client", IssuedAt: 1, ExpiresAt: 2})
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid compact JWT: %s", token)
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	var header map[string]string
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		t.Fatal(err)
	}
	if header["alg"] != "RS256" || header["kid"] != "key-1" || header["typ"] != "JWT" {
		t.Fatalf("unexpected header: %+v", header)
	}
	jwks, err := keys.PublicKeys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(jwks) != 1 || jwks[0].Kid != "key-1" {
		t.Fatalf("unexpected JWKS: %+v", jwks)
	}
	modulus, err := base64.RawURLEncoding.DecodeString(jwks[0].N)
	if err != nil {
		t.Fatal(err)
	}
	exponent, err := base64.RawURLEncoding.DecodeString(jwks[0].E)
	if err != nil {
		t.Fatal(err)
	}
	pub := rsa.PublicKey{N: new(big.Int).SetBytes(modulus), E: int(new(big.Int).SetBytes(exponent).Int64())}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(&pub, crypto.SHA256, digest[:], signature); err != nil {
		t.Fatalf("signature verification failed: %v", err)
	}
}

func TestRSAKeyRotationRetainsOldPublicKey(t *testing.T) {
	keys, err := GenerateRSAKeySet("old", 2048)
	if err != nil {
		t.Fatal(err)
	}
	newKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	if err := keys.Add("new", newKey, true); err != nil {
		t.Fatal(err)
	}
	if keys.publicKeys["new"].N == keys.activeKey.N {
		t.Fatal("rotated public key retains active private key")
	}
	if err := keys.Add("new", newKey, false); err == nil {
		t.Fatal("duplicate key ID was accepted")
	}
	jwks, err := keys.PublicKeys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(jwks) != 2 || jwks[0].Kid != "new" || jwks[1].Kid != "old" {
		t.Fatalf("rotation did not retain keys: %+v", jwks)
	}
	if err := keys.Remove("new"); err == nil {
		t.Fatal("active key removal must fail")
	}
	if err := keys.Remove("old"); err != nil {
		t.Fatal(err)
	}
}

func TestRSAKeyRetirementPrunesOnlyAtBoundary(t *testing.T) {
	keys, err := GenerateRSAKeySet("old", 2048)
	if err != nil {
		t.Fatal(err)
	}
	newKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	if err := keys.Add("new", newKey, true); err != nil {
		t.Fatal(err)
	}
	retireAt := time.Unix(1_700_000_000, 0)
	if err := keys.Retire("old", retireAt); err != nil {
		t.Fatal(err)
	}
	if err := keys.Retire("old", retireAt.Add(-time.Second)); err == nil {
		t.Fatal("key retirement was moved earlier")
	}
	if removed := keys.Prune(retireAt.Add(-time.Nanosecond)); len(removed) != 0 {
		t.Fatalf("key pruned early: %v", removed)
	}
	if removed := keys.Prune(retireAt); len(removed) != 1 || removed[0] != "old" {
		t.Fatalf("key not pruned at boundary: %v", removed)
	}
	if err := keys.Retire("new", retireAt); err == nil {
		t.Fatal("active signing key retirement accepted")
	}
}

func TestRSAPrivateKeyPKCS8PEMRoundTrip(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := MarshalRSAPrivateKeyPEM(key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseRSAPrivateKeyPEM(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(key.N) != 0 || parsed.D.Cmp(key.D) != 0 {
		t.Fatal("private key changed during PEM round trip")
	}
	set, err := NewRSAKeySet("active", parsed)
	if err != nil {
		t.Fatal(err)
	}
	verificationKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	if err := set.AddVerificationKey("old", &verificationKey.PublicKey); err != nil {
		t.Fatal(err)
	}
	if err := set.AddVerificationKey("old", &verificationKey.PublicKey); err == nil {
		t.Fatal("duplicate verification key ID was accepted")
	}
	jwks, err := set.PublicKeys(context.Background())
	if err != nil || len(jwks) != 2 {
		t.Fatalf("verification-only key not published: %+v, %v", jwks, err)
	}
}

func TestRSAKeySetRejectsInvalidIDTokenClaims(t *testing.T) {
	keys, err := GenerateRSAKeySet("key", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := keys.SignIDToken(context.Background(), IDTokenClaims{Subject: "user"}); err == nil {
		t.Fatal("invalid required claims were signed")
	}
	if _, err := keys.SignIDToken(context.Background(), IDTokenClaims{Issuer: "https://issuer.example", Subject: "user", Audience: "client", IssuedAt: 10, ExpiresAt: 9}); err == nil {
		t.Fatal("invalid timestamps were signed")
	}
}
