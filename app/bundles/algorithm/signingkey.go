package algorithm

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

var fallbackSigningKeyCounter atomic.Uint64

// GenerateRandomBytes returns n cryptographically secure random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
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

// SafeGenerateSigningKey returns a URL-safe signing key, falling back to best-effort entropy.
func SafeGenerateSigningKey(keyLength int) string {
	if keyLength <= 0 {
		keyLength = 32
	}
	signingKey, err := GenerateSigningKey(keyLength)
	if err == nil {
		return signingKey
	}

	slog.Warn("secure signing key generation failed, using fallback entropy", "err", err)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(fallbackSigningKeyBytes(keyLength))
}

func fallbackSigningKeyBytes(keyLength int) []byte {
	hostname, _ := os.Hostname()
	seed := fmt.Sprintf("%d:%d:%d:%s", time.Now().UnixNano(), os.Getpid(), fallbackSigningKeyCounter.Add(1), hostname)
	bytes := make([]byte, 0, keyLength)
	for len(bytes) < keyLength {
		sum := sha256.Sum256([]byte(seed))
		bytes = append(bytes, sum[:]...)
		seed = base64.StdEncoding.EncodeToString(sum[:]) + ":" + seed
	}
	return bytes[:keyLength]
}
