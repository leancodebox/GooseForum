package algorithm

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"strings"
)

const (
	saltLength     = 32
	hashIterations = 10000
	hashKeyLen     = 32
)

func MakePassword(password string) (string, error) {
	hash, salt, err := EncryptPassword(password)
	return hash + ":" + salt, err
}

func VerifyEncryptPassword(secretPassword, inputPassword string) error {
	passwordStore := strings.Split(secretPassword, ":")
	if len(passwordStore) != 2 {
		return errors.New("no pass")
	}
	return VerifyPassword(passwordStore[0], passwordStore[1], inputPassword)
}

// 加密函数，接收原始密码字符串，返回加密后的密码和盐值
func EncryptPassword(password string) (string, string, error) {
	// 生成随机盐
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", "", err
	}

	// 使用 PBKDF2 算法加密密码
	hash := pbkdf2SHA256([]byte(password), salt, hashIterations, hashKeyLen)
	encodedHash := base64.StdEncoding.EncodeToString(hash)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	return encodedHash, encodedSalt, nil
}

// 验证函数，接收密文、原始密码和盐值，返回验证结果
func VerifyPassword(encodedHash, encodedSalt, inputPassword string) error {
	// 解码密文和盐值
	hash, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		return errors.New("invalid password hash")
	}
	salt, err := base64.StdEncoding.DecodeString(encodedSalt)
	if err != nil {
		return errors.New("invalid password salt")
	}

	// 使用相同的盐值和迭代次数计算输入密码的散列值
	inputHash := pbkdf2SHA256([]byte(inputPassword), salt, hashIterations, hashKeyLen)

	// 比较密文和计算出的散列值
	if !equalHashes(hash, inputHash) {
		return errors.New("incorrect password")
	}

	return nil
}

// 判断两个字节数组是否相等，防止时序攻击
func equalHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// 使用 PBKDF2 和 SHA256 哈希函数加密密码
func pbkdf2SHA256(password []byte, salt []byte, iterations int, keyLen int) []byte {
	hashFunc := sha256.New
	dk := pbkdf2.Key(password, salt, iterations, keyLen, hashFunc)
	return dk
}
