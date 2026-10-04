package oidcProviderStore

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"time"
)

func RevokeAccountTokens(subject string, now time.Time) error {
	if err := dbconnect.Connect().Model(&TokenEntity{}).Where("user_id = ? AND revoked_at IS NULL", subject).Update("revoked_at", now).Error; err != nil {
		return err
	}
	return dbconnect.Connect().Model(&AuthorizationCodeEntity{}).Where("user_id = ? AND used = ?", subject, false).Update("used", true).Error
}
