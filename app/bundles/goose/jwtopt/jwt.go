package jwtopt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type CustomClaims struct {
	UserId uint64
	jwt.RegisteredClaims
}

var (
	TokenInvalid = errors.New("Couldn't handle this token: ")
)

type JWT struct {
	SigningKey []byte
}

func NewJWT(signingKey []byte) *JWT {
	return &JWT{
		signingKey,
	}
}


func GetBaseRegisteredClaims(expireTime time.Duration) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		NotBefore: jwt.NewNumericDate(time.Now().Add(-10)),        // 签名生效时间
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)), // 过期时间 7天  配置文件
		Issuer:    "thh",                                          // 签名的发行者
	}
}

// CreateToken 创建一个token
func (itself *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(itself.SigningKey)
}

// ParseToken 解析 token
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
	return nil, TokenInvalid
}
