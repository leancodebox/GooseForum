package authsessions

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"time"

	"gorm.io/gorm"
)

func CreateToken(token *Token) error {
	return dbconnect.Connect().Create(token).Error
}

func RevokeAllTokens(userID uint64, now time.Time) error {
	return RevokeAllTokensWithDB(dbconnect.Connect(), userID, now)
}

func CreateSecurityLog(log *Log) error {
	return dbconnect.Connect().Create(log).Error
}
func RevokeAllTokensWithDB(db *gorm.DB, userID uint64, now time.Time) error {
	return db.Model(&Token{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error
}
