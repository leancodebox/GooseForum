package users

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrSecurityChanged = errors.New("user security state changed; reload and retry")

func UpdateEmailByVersion(userID, version uint64, email string) error {
	return updateEmailByVersion(userID, version, email, false)
}

// UpdateVerifiedEmailByVersion is reserved for operator-verified recovery.
func UpdateVerifiedEmailByVersion(userID, version uint64, email string) error {
	return updateEmailByVersion(userID, version, email, true)
}

func updateEmailByVersion(userID, version uint64, email string, verified bool) error {
	email = strings.TrimSpace(email)
	return WithIdentityWriteLock(func() error {
		db := builder().Session(&gorm.Session{})
		if err := CheckIdentityAvailable("", email, userID); err != nil {
			return err
		}
		changes := map[string]any{
			"email": email, "email_key": identityKey(email), "is_activated": ActivationPending, "activated_at": nil,
			"requires_email_verification": true,
			"token_version":               gorm.Expr("token_version + 1"),
		}
		if verified {
			changes["is_activated"] = ActivationSuccess
			changes["activated_at"] = time.Now()
			changes["requires_email_verification"] = false
		}
		result := db.Where("id = ? AND token_version = ?", userID, version).Updates(changes)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrSecurityChanged
		}
		return nil
	})
}

func UpdateUsernameByVersion(userID, version uint64, username string) error {
	username = strings.TrimSpace(username)
	return WithIdentityWriteLock(func() error {
		db := builder().Session(&gorm.Session{})
		if err := CheckIdentityAvailable(username, "", userID); err != nil {
			return err
		}
		result := db.Where("id = ? AND token_version = ?", userID, version).Updates(map[string]any{"username": username, "username_key": identityKey(username)})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrSecurityChanged
		}
		return nil
	})
}

func ActivateByVersion(userID, version uint64, email string) (bool, error) {
	result := builder().Where("id = ? AND email = ? AND token_version = ?", userID, email, version).Updates(map[string]any{"is_activated": ActivationSuccess, "activated_at": time.Now()})
	return result.RowsAffected == 1, result.Error
}

func AssignInitialRole(userID, roleID uint64) error {
	return builder().Where("id = ?", userID).Update("role_id", roleID).Error
}

func AssignOperatorRole(userID, roleID uint64) error {
	return builder().Where("id = ?", userID).Updates(map[string]any{"role_id": roleID, "token_version": gorm.Expr("token_version + 1")}).Error
}
