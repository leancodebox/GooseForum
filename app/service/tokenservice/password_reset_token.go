package tokenservice

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// PasswordResetClaims is the JWT payload used for password reset links.
type PasswordResetClaims struct {
	UserId       uint64  `json:"userId"`
	Email        string  `json:"email"`
	TokenVersion *uint64 `json:"tokenVersion"`
	jwt.RegisteredClaims
}

// GeneratePasswordResetToken creates a signed password reset token.
func GeneratePasswordResetToken(userId uint64, email string, tokenVersion uint64) (string, error) {
	claims := PasswordResetClaims{
		UserId:       userId,
		Email:        email,
		TokenVersion: &tokenVersion,
		ExpiresAt:    jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		IssuedAt:     jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey())
}

// ParsePasswordResetToken parses and validates a password reset token.
func ParsePasswordResetToken(tokenString string) (*PasswordResetClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &PasswordResetClaims{}, func(token *jwt.Token) (any, error) {
		return signingKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*PasswordResetClaims); ok && token.Valid {
		if claims.TokenVersion == nil {
			return nil, jwt.ErrTokenInvalidClaims
		}
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
