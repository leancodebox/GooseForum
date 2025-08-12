package tokenservice

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"

	"github.com/golang-jwt/jwt/v5"
)

type PasswordResetClaims struct {
	UserId uint64 `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GeneratePasswordResetToken(userId uint64, email string) (string, error) {
	claims := PasswordResetClaims{
		UserId: userId,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)), // 30分钟有效期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(preferences.Get("jwtopt.key")))
}

func ParsePasswordResetToken(tokenString string) (*PasswordResetClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &PasswordResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(preferences.Get("jwtopt.key")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*PasswordResetClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}