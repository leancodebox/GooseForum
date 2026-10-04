package algorithm

import (
	"encoding/base64"
	"testing"
)

func TestGenerateSigningKey(t *testing.T) {
	bytes, err := GenerateRandomBytes(32)
	if err != nil {
		t.Fatal(err)
	}
	if len(bytes) != 32 {
		t.Fatalf("GenerateRandomBytes length = %d, want 32", len(bytes))
	}

	key, err := GenerateSigningKey(32)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(key)
	if err != nil {
		t.Fatalf("generated key should be URL-safe base64: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded signing key length = %d, want 32", len(decoded))
	}
	if len(key) != 43 {
		t.Fatalf("encoded signing key length = %d, want 43", len(key))
	}
	other, err := GenerateSigningKey(32)
	if err != nil || other == key {
		t.Fatal("successive signing keys must be independently generated")
	}
}

func TestGenerateSigningKeyRejectsInvalidLength(t *testing.T) {
	for _, length := range []int{-1, 0} {
		if _, err := GenerateSigningKey(length); err == nil {
			t.Fatal("invalid key length accepted")
		}
	}
}

func TestSafeGenerateSigningKey(t *testing.T) {
	key := SafeGenerateSigningKey(32)
	decoded, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(key)
	if err != nil {
		t.Fatalf("safe key should be URL-safe base64: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded safe signing key length = %d, want 32", len(decoded))
	}
}
