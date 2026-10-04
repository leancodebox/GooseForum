package authsessions

import (
	"time"

	"gorm.io/gorm"
)

func ExcessTokens(userID uint64, now time.Time, limit int) ([]Token, error) {
	var tokens []Token
	err := builder().Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Order("last_seen_at DESC, id DESC").Offset(limit).Find(&tokens).Error
	return tokens, err
}

func RenewToken(id uint64, now, expiry time.Time) (bool, error) {
	result := builder().Model(&Token{}).Where("id = ? AND revoked_at IS NULL AND expires_at > ? AND expires_at < ?", id, now, expiry).
		Update("expires_at", expiry)
	return result.RowsAffected > 0, result.Error
}

func TouchToken(id uint64, threshold, now time.Time) (bool, error) {
	result := builder().Model(&Token{}).Where("id = ? AND revoked_at IS NULL AND last_seen_at < ?", id, threshold).Update("last_seen_at", now)
	return result.RowsAffected > 0, result.Error
}

func GetValidToken(hash string, now time.Time) (Token, error) {
	var token Token
	err := builder().Model(&Token{}).
		Joins("JOIN users ON users.id = user_auth_tokens.user_id AND users.deleted_at IS NULL AND users.token_version = user_auth_tokens.token_version").
		Where("user_auth_tokens.token_hash = ? AND user_auth_tokens.revoked_at IS NULL AND user_auth_tokens.expires_at > ?", hash, now).
		First(&token).Error
	return token, err
}

func GetTokenByHash(hash string) (Token, error) {
	var token Token
	err := builder().Where("token_hash = ?", hash).First(&token).Error
	return token, err
}

func GetUserToken(userID, id uint64) (Token, error) {
	var token Token
	err := builder().Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).First(&token).Error
	return token, err
}

func RevokeToken(userID, id uint64, now time.Time) (bool, error) {
	result := builder().Model(&Token{}).Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).Update("revoked_at", now)
	return result.RowsAffected == 1, result.Error
}

func RevokeOtherTokens(userID, currentID uint64, now time.Time) error {
	return builder().Model(&Token{}).Where("user_id = ? AND id <> ? AND revoked_at IS NULL", userID, currentID).Update("revoked_at", now).Error
}

func ListUserTokens(userID, version uint64, now time.Time) ([]Token, error) {
	var tokens []Token
	err := builder().Where("user_id = ? AND token_version = ? AND revoked_at IS NULL AND expires_at > ?", userID, version, now).
		Order("last_seen_at DESC, id DESC").Find(&tokens).Error
	return tokens, err
}

const CleanupBatchSize = 500

func CleanupTokens(cutoff time.Time) error {
	return deleteExpiredInBatches(builder(), &Token{}, "expires_at < ? OR revoked_at < ?", cutoff, cutoff)
}

func CleanupLogs(cutoff time.Time) error {
	return deleteExpiredInBatches(logBuilder(), &Log{}, "created_at < ?", cutoff)
}

func deleteExpiredInBatches(db *gorm.DB, model any, condition string, args ...any) error {
	type expiredID struct{ Id uint64 }
	for {
		var rows []expiredID
		if err := db.Session(&gorm.Session{}).Model(model).Where(condition, args...).Order("id").Limit(CleanupBatchSize).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		ids := make([]uint64, len(rows))
		for index, row := range rows {
			ids[index] = row.Id
		}
		if err := db.Session(&gorm.Session{}).Where("id IN ?", ids).Delete(model).Error; err != nil {
			return err
		}
		if len(ids) < CleanupBatchSize {
			return nil
		}
	}
}
