package jwtopt

import (
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"sync"
	"time"
)

var (
	once       sync.Once
	std        *JWT
	signingKey = preferences.Get("jwtopt.signingKey", "mq+ZeGafL+b1xdC0u9vSVg==")
	validTime  = time.Duration(preferences.GetInt64("jwtopt.validTime", 86400*7)) * time.Second
)

func Std() *JWT {
	once.Do(func() {
		std = NewJWT([]byte(signingKey))
	})
	return std
}

func CreateNewToken(userId uint64, expireTime time.Duration) (string, error) {
	cc := CustomClaims{
		UserId:           userId,
		RegisteredClaims: GetBaseRegisteredClaims(expireTime),
	}
	return Std().CreateToken(cc)
}

// VerifyTokenWithFresh 验证token 并刷新， 如果token还有1天就过期则生成新的token，否则还是用原来的
func VerifyTokenWithFresh(tokenStr string) (userId uint64, newToken string, err error) {
	claims, err := Std().ParseToken(tokenStr)
	if err != nil {
		return 0, "", err
	}
	eTime, err := claims.GetExpirationTime()
	if err == nil && time.Now().Add(time.Second*86400*1).After(eTime.Time) {
		claims.RegisteredClaims = GetBaseRegisteredClaims(validTime)
		tokenStr, err = Std().CreateToken(*claims)
	}
	return claims.UserId, tokenStr, err
}

func VerifyToken(tokenStr string) (userId uint64, err error) {
	claims, err := Std().ParseToken(tokenStr)
	if err != nil {
		return 0, err
	}
	return claims.UserId, err
}
