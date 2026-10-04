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

// GenerateSecretKey returns a URL-safe base64 secret key with no padding.
func GenerateSecretKey(keyLength int) (string, error) {
	bytes, err := GenerateRandomBytes(keyLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

// SafeGenerateSecretKey returns a secure key or stops if secure randomness is unavailable.
func SafeGenerateSecretKey(keyLength int) string {
	if keyLength <= 0 {
		keyLength = 32
	}
	secretKey, err := GenerateSecretKey(keyLength)
	if err == nil {
		return secretKey
	}

	panic(fmt.Errorf("secure secret key generation failed: %w", err))
}
