package algorithm

import (
	"encoding/base64"
	"testing"
)

func TestGenerateSecretKey(t *testing.T) {
	bytes, err := GenerateRandomBytes(32)
	if err != nil {
		t.Fatal(err)
	}
	if len(bytes) != 32 {
		t.Fatalf("GenerateRandomBytes length = %d, want 32", len(bytes))
	}

	key, err := GenerateSecretKey(32)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(key)
	if err != nil {
		t.Fatalf("generated key should be URL-safe base64: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded secret key length = %d, want 32", len(decoded))
	}
	if len(key) != 43 {
		t.Fatalf("encoded secret key length = %d, want 43", len(key))
	}
	other, err := GenerateSecretKey(32)
	if err != nil || other == key {
		t.Fatal("successive secret keys must be independently generated")
	}
}

func TestGenerateSecretKeyRejectsInvalidLength(t *testing.T) {
	for _, length := range []int{-1, 0} {
		if _, err := GenerateSecretKey(length); err == nil {
			t.Fatal("invalid key length accepted")
		}
	}
}

func TestSafeGenerateSecretKey(t *testing.T) {
	key := SafeGenerateSecretKey(32)
	decoded, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(key)
	if err != nil {
		t.Fatalf("safe key should be URL-safe base64: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded safe secret key length = %d, want 32", len(decoded))
	}
}
