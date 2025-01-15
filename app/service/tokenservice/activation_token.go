package tokenservice

import (
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ActivationClaims struct {
	UserId uint64 `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateActivationTokenByUser(entity users.Entity) (string, error) {
	return GenerateActivationToken(entity.Id, entity.Email)
}

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
	return token.SignedString([]byte(preferences.Get("jwtopt.key")))
}

func ParseActivationToken(tokenString string) (*ActivationClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ActivationClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(preferences.Get("jwtopt.key")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ActivationClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
