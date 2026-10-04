package algorithm

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRandomBytes returns n cryptographically secure random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, fmt.Errorf("random byte length must be positive")
	}
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("随机数生成失败: %w", err)
	}
	return b, nil
}

// GenerateSigningKey returns a URL-safe base64 signing key with no padding.
func GenerateSigningKey(keyLength int) (string, error) {
	bytes, err := GenerateRandomBytes(keyLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

// SafeGenerateSigningKey returns a secure key or stops if secure randomness is unavailable.
func SafeGenerateSigningKey(keyLength int) string {
	if keyLength <= 0 {
		keyLength = 32
	}
	signingKey, err := GenerateSigningKey(keyLength)
	if err == nil {
		return signingKey
	}

	panic(fmt.Errorf("secure signing key generation failed: %w", err))
}
