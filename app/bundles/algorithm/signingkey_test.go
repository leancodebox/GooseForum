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
	decoded, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(key)
	if err != nil {
		t.Fatalf("generated key should be URL-safe base64: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded signing key length = %d, want 32", len(decoded))
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
