// Package jwtopt creates and validates GooseForum access tokens.
package jwtopt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims stores the GooseForum user ID in an access token.
type CustomClaims struct {
	UserId uint64
	jwt.RegisteredClaims
}

var (
	// ErrTokenInvalid is returned when a token cannot be parsed into CustomClaims.
	ErrTokenInvalid = errors.New("couldn't handle this token")
)

// JWT signs and parses tokens with a fixed signing key.
type JWT struct {
	SigningKey []byte
}

// NewJWT creates a JWT helper with signingKey.
func NewJWT(signingKey []byte) *JWT {
	return &JWT{
		signingKey,
	}
}

// GetBaseRegisteredClaims returns standard claims for a token lifetime.
func GetBaseRegisteredClaims(expireTime time.Duration) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		NotBefore: jwt.NewNumericDate(time.Now().Add(-10)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)),
		Issuer:    "GooseForum",
	}
}

// CreateToken signs claims as a JWT string.
func (itself *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(itself.SigningKey)
}

// ParseToken parses tokenString into CustomClaims.
func (itself *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},
		func(token *jwt.Token) (i any, e error) {
			return itself.SigningKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, ErrTokenInvalid
}
