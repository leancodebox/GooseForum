package oidcprovider

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"sort"
	"sync"
	"time"
)

// RSAKeySet signs ID tokens with its active RS256 key and publishes every key
// retained for verification. Retain old keys until all tokens they signed have
// expired.
type RSAKeySet struct {
	mu         sync.RWMutex
	activeKID  string
	activeKey  *rsa.PrivateKey
	publicKeys map[string]*rsa.PublicKey
	retireAt   map[string]time.Time
}

func NewRSAKeySet(kid string, key *rsa.PrivateKey) (*RSAKeySet, error) {
	if kid == "" || key == nil || key.N.BitLen() < 2048 {
		return nil, errors.New("kid and an RSA key of at least 2048 bits are required")
	}
	if err := key.Validate(); err != nil {
		return nil, err
	}
	owned, err := cloneRSAPrivateKey(key)
	if err != nil {
		return nil, err
	}
	return &RSAKeySet{activeKID: kid, activeKey: owned, publicKeys: map[string]*rsa.PublicKey{kid: cloneRSAPublicKey(&owned.PublicKey)}, retireAt: map[string]time.Time{}}, nil
}

func GenerateRSAKeySet(kid string, bits int) (*RSAKeySet, error) {
	if bits < 2048 {
		return nil, errors.New("RSA key size must be at least 2048 bits")
	}
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return NewRSAKeySet(kid, key)
}

// Add installs a verification key and optionally makes it active for signing.
func (s *RSAKeySet) Add(kid string, key *rsa.PrivateKey, active bool) error {
	if kid == "" || key == nil || key.N.BitLen() < 2048 {
		return errors.New("kid and an RSA key of at least 2048 bits are required")
	}
	if err := key.Validate(); err != nil {
		return err
	}
	owned, err := cloneRSAPrivateKey(key)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.publicKeys[kid]; exists {
		return errors.New("signing key ID already exists")
	}
	s.publicKeys[kid] = cloneRSAPublicKey(&owned.PublicKey)
	if active {
		s.activeKID, s.activeKey = kid, owned
	}
	return nil
}

// AddVerificationKey retains a public key for validating tokens signed before
// rotation without retaining the corresponding private key.
func (s *RSAKeySet) AddVerificationKey(kid string, key *rsa.PublicKey) error {
	if kid == "" || key == nil || key.N == nil || key.N.Sign() <= 0 || key.N.BitLen() < 2048 || key.E < 3 || key.E%2 == 0 {
		return errors.New("kid and a valid RSA public key of at least 2048 bits are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.publicKeys[kid]; exists {
		return errors.New("signing key ID already exists")
	}
	s.publicKeys[kid] = cloneRSAPublicKey(key)
	return nil
}

// Remove removes an old verification key. The active signing key cannot be
// removed until another key has been activated.
func (s *RSAKeySet) Remove(kid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if kid == s.activeKID {
		return errors.New("cannot remove active signing key")
	}
	delete(s.publicKeys, kid)
	delete(s.retireAt, kid)
	return nil
}

// Retire schedules a non-active verification key for removal after every ID
// token signed by it has expired.
func (s *RSAKeySet) Retire(kid string, at time.Time) error {
	if at.IsZero() {
		return errors.New("retirement time is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if kid == s.activeKID {
		return errors.New("cannot retire active signing key")
	}
	if _, exists := s.publicKeys[kid]; !exists {
		return errors.New("signing key ID does not exist")
	}
	if current, exists := s.retireAt[kid]; exists && at.Before(current) {
		return errors.New("retirement time cannot be moved earlier")
	}
	s.retireAt[kid] = at
	return nil
}

// Prune removes verification keys whose retirement time has arrived.
func (s *RSAKeySet) Prune(now time.Time) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := make([]string, 0)
	for kid, at := range s.retireAt {
		if !now.Before(at) {
			delete(s.publicKeys, kid)
			delete(s.retireAt, kid)
			removed = append(removed, kid)
		}
	}
	sort.Strings(removed)
	return removed
}

func (s *RSAKeySet) SignIDToken(_ context.Context, claims IDTokenClaims) (string, error) {
	if err := claims.Validate(); err != nil {
		return "", err
	}
	s.mu.RLock()
	kid, key := s.activeKID, s.activeKey
	s.mu.RUnlock()
	if key == nil {
		return "", errors.New("active signing key is unavailable")
	}
	header, err := json.Marshal(map[string]string{"alg": "RS256", "kid": kid, "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(encoded))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return encoded + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (s *RSAKeySet) PublicKeys(context.Context) ([]JWK, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	kids := make([]string, 0, len(s.publicKeys))
	for kid := range s.publicKeys {
		kids = append(kids, kid)
	}
	sort.Strings(kids)
	result := make([]JWK, 0, len(kids))
	for _, kid := range kids {
		key := s.publicKeys[kid]
		exponent := big.NewInt(int64(key.E)).Bytes()
		result = append(result, JWK{KTY: "RSA", Use: "sig", Alg: "RS256", Kid: kid, N: base64.RawURLEncoding.EncodeToString(key.N.Bytes()), E: base64.RawURLEncoding.EncodeToString(exponent)})
	}
	return result, nil
}

// MarshalRSAPrivateKeyPEM encodes a signing key as unencrypted PKCS#8 PEM.
// Encryption at rest is the responsibility of the host's key store.
func MarshalRSAPrivateKeyPEM(key *rsa.PrivateKey) ([]byte, error) {
	if key == nil {
		return nil, errors.New("private key is required")
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	encoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	clear(der)
	return encoded, nil
}

func ParseRSAPrivateKeyPEM(data []byte) (*rsa.PrivateKey, error) {
	block, rest := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" || len(rest) != 0 {
		return nil, errors.New("invalid PKCS#8 private key PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("PKCS#8 key is not RSA")
	}
	if key.N.BitLen() < 2048 {
		return nil, errors.New("RSA key size must be at least 2048 bits")
	}
	if err := key.Validate(); err != nil {
		return nil, err
	}
	return key, nil
}

func cloneRSAPrivateKey(key *rsa.PrivateKey) (*rsa.PrivateKey, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	clear(der)
	if err != nil {
		return nil, err
	}
	return parsed.(*rsa.PrivateKey), nil
}

func cloneRSAPublicKey(key *rsa.PublicKey) *rsa.PublicKey {
	return &rsa.PublicKey{N: new(big.Int).Set(key.N), E: key.E}
}
