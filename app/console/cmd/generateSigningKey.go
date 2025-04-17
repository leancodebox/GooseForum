package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "generate:SigningKey",
		Short: "生成 signingKey",
		Run:   runSigningKey,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
}
func runSigningKey(_ *cobra.Command, _ []string) {
	signingKey, err := generateSigningKey(32)
	if err != nil {
		fmt.Println("Failed to generate signingKey:", err)
		return
	}

	fmt.Println("Generated signingKey:", signingKey)
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("随机数生成失败: %w", err)
	}
	return b, nil
}

// 生成符合JWT规范的签名密钥
func generateSigningKey(keyLength int) (string, error) {
	bytes, err := generateRandomBytes(keyLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}
