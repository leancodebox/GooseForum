// Package tokenservice creates and parses short-lived account tokens.
package tokenservice

import (
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/users"

	"github.com/golang-jwt/jwt/v5"
)

// ActivationClaims is the JWT payload used for email activation.
type ActivationClaims struct {
	UserId uint64 `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateActivationTokenByUser creates an activation token from a user entity.
func GenerateActivationTokenByUser(entity users.EntityComplete) (string, error) {
	return GenerateActivationToken(entity.Id, entity.Email)
}

// GenerateActivationToken creates a signed email activation token.
func GenerateActivationToken(userId uint64, email string) (string, error) {
	claims := ActivationClaims{
		UserId: userId,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey())
}

// ParseActivationToken parses and validates an activation token.
func ParseActivationToken(tokenString string) (*ActivationClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ActivationClaims{}, func(token *jwt.Token) (any, error) {
		return signingKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ActivationClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
