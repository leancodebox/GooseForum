package agenttokens

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func List(userID uint64) ([]Entity, error) {
	rows := []Entity{}
	err := dbconnect.Connect().Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func GetByDigest(hash string) (Entity, error) {
	var row Entity
	err := dbconnect.Connect().Where("hash = ?", hash).Take(&row).Error
	return row, err
}

func Revoke(userID uint64, id string) error {
	return dbconnect.Connect().Model(&Entity{}).Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).Update("revoked_at", time.Now()).Error
}

func RevokeAll(userID uint64) error {
	return dbconnect.Connect().Model(&Entity{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", time.Now()).Error
}

func Delete(userID uint64, id string) error {
	return dbconnect.Connect().Where("id = ? AND user_id = ?", id, userID).Delete(&Entity{}).Error
}

func CreateToken(userID, version uint64, name string, scopes []string, days int) (Entity, string, error) {
	return Create(dbconnect.Connect(), userID, version, name, scopes, days)
}
