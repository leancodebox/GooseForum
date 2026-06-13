package logincrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	keyBits       = 2048
	payloadMaxAge = 3 * time.Minute
)

var (
	keyMu      sync.RWMutex
	privateKey *rsa.PrivateKey
	publicPEM  string
)

type PasswordPayload struct {
	Password string `json:"password"`
	Ts       int64  `json:"ts"`
}

func init() {
	if err := rotateKey(); err != nil {
		panic(fmt.Sprintf("generate login crypto key failed: %v", err))
	}
}

func PublicKeyPEM() string {
	keyMu.RLock()
	defer keyMu.RUnlock()
	return publicPEM
}

func DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", errors.New("encrypted password is required")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", errors.New("invalid encrypted password")
	}

	keyMu.RLock()
	key := privateKey
	keyMu.RUnlock()

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, ciphertext, nil)
	if err != nil {
		return "", errors.New("decrypt password failed")
	}

	var payload PasswordPayload
	if err = json.Unmarshal(plaintext, &payload); err != nil {
		return "", errors.New("invalid password payload")
	}

	if payload.Password == "" {
		return "", errors.New("password is required")
	}
	if payload.Ts <= 0 {
		return "", errors.New("password payload timestamp is required")
	}

	age := time.Since(time.UnixMilli(payload.Ts))
	if age < -payloadMaxAge || age > payloadMaxAge {
		return "", errors.New("password payload expired")
	}

	return payload.Password, nil
}

func rotateKey() error {
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return err
	}

	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}

	block := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})
	if block == nil {
		return errors.New("encode public key failed")
	}

	keyMu.Lock()
	privateKey = key
	publicPEM = string(block)
	keyMu.Unlock()
	return nil
}
