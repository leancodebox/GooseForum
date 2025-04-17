package algorithm

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("随机数生成失败: %w", err)
	}
	return b, nil
}

// 生成符合JWT规范的签名密钥
func GenerateSigningKey(keyLength int) (string, error) {
	bytes, err := GenerateRandomBytes(keyLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}
