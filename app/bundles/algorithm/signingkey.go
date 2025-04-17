package algorithm

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mRand "math/rand"
	"time"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("随机数生成失败: %w", err)
	}
	return b, nil
}

func GenerateSigningKey(keyLength int) (string, error) {
	bytes, err := GenerateRandomBytes(keyLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

func SafeGenerateSigningKey(keyLength int) string {
	signingKey, err := GenerateSigningKey(keyLength)
	if err != nil {
		bytes := make([]byte, keyLength)
		fallbackSource := mRand.New(mRand.NewSource(time.Now().UnixNano()))
		for i := 0; i < keyLength; i++ {
			bytes[i] = byte(fallbackSource.Intn(256))
		}
		return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)
	}
	// 强制生成 Base64 字符串，无错误返回
	return signingKey
}
