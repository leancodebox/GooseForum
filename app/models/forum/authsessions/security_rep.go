package authsessions

import (
	"time"
)

func CreateToken(token *Token) error {
	return builder().Create(token).Error
}

func RevokeAllTokens(userID uint64, now time.Time) error {
	return builder().Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error
}

func CreateSecurityLog(log *Log) error {
	return logBuilder().Create(log).Error
}
