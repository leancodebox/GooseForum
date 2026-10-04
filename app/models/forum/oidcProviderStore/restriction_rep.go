package oidcProviderStore

import (
	"time"
)

func RevokeAccountTokens(subject string, now time.Time) error {
	if err := tokenBuilder().Where("user_id = ? AND revoked_at IS NULL", subject).Update("revoked_at", now).Error; err != nil {
		return err
	}
	return codeBuilder().Where("user_id = ? AND used = ?", subject, false).Update("used", true).Error
}
