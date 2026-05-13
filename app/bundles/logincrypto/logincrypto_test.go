package logincrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"testing"
	"time"
)

func TestDecryptPassword(t *testing.T) {
	encrypted := encryptForTest(t, PasswordPayload{
		Password: "secret123",
		Ts:       time.Now().UnixMilli(),
	})

	password, err := DecryptPassword(encrypted)
	if err != nil {
		t.Fatalf("DecryptPassword returned error: %v", err)
	}
	if password != "secret123" {
		t.Fatalf("expected password to round-trip, got %q", password)
	}
}

func TestDecryptPasswordRejectsExpiredPayload(t *testing.T) {
	encrypted := encryptForTest(t, PasswordPayload{
		Password: "secret123",
		Ts:       time.Now().Add(-payloadMaxAge - time.Minute).UnixMilli(),
	})

	if _, err := DecryptPassword(encrypted); err == nil {
		t.Fatal("expected expired payload to fail")
	}
}

func encryptForTest(t *testing.T, payload PasswordPayload) string {
	t.Helper()

	block, _ := pem.Decode([]byte(PublicKeyPEM()))
	if block == nil {
		t.Fatal("failed to decode public key pem")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("ParsePKIXPublicKey returned error: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		t.Fatal("public key is not RSA")
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, plaintext, nil)
	if err != nil {
		t.Fatalf("EncryptOAEP returned error: %v", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext)
}
