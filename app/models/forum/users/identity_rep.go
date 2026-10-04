package users

import (
	"errors"
	"strings"
	"sync"

	"gorm.io/gorm"
)

var identityWriteMu sync.Mutex

func (user EntityComplete) NeedsEmailVerification(policyEnabled bool) bool {
	return user.IsActivated == ActivationPending && (policyEnabled || user.RequiresEmailVerification)
}

var (
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")
)

// WithIdentityWriteLock serializes identity creation and changes in this process.
func WithIdentityWriteLock(fn func() error) error {
	identityWriteMu.Lock()
	defer identityWriteMu.Unlock()
	return fn()
}

func CheckIdentityAvailable(username, email string, excludeUserID uint64) error {
	return checkIdentityAvailable(builder(), username, email, excludeUserID)
}

func checkIdentityAvailable(tx *gorm.DB, username, email string, excludeUserID uint64) error {
	for _, identity := range []struct {
		column, value string
		conflict      error
	}{
		{"username", username, ErrUsernameExists}, {"email", email, ErrEmailExists},
	} {
		if strings.TrimSpace(identity.value) == "" {
			continue
		}
		var match Identity
		result := tx.Session(&gorm.Session{}).Model(&EntityComplete{}).Where("id <> ? AND "+identity.column+"_key = ?", excludeUserID, identityKey(identity.value)).Limit(1).Find(&match)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			return identity.conflict
		}
	}
	return nil
}

func identityKey(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func (user *EntityComplete) BeforeCreate(_ *gorm.DB) error {
	user.UsernameKey = identityKey(user.Username)
	user.EmailKey = identityKey(user.Email)
	return nil
}

// BackfillIdentityKeys preserves legacy duplicate accounts while indexing lookups.
func BackfillIdentityKeys(db *gorm.DB) error {
	type identityRow struct {
		Id              uint64
		Username, Email string
	}
	var afterID uint64
	for {
		var rows []identityRow
		if err := db.Model(&EntityComplete{}).Unscoped().Where("id > ? AND ((username_key = '' AND username <> '') OR (email_key = '' AND email <> ''))", afterID).Order("id").Limit(500).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		for _, row := range rows {
			if err := db.Model(&EntityComplete{}).Unscoped().Where("id = ?", row.Id).UpdateColumns(map[string]any{"username_key": identityKey(row.Username), "email_key": identityKey(row.Email)}).Error; err != nil {
				return err
			}
		}
		afterID = rows[len(rows)-1].Id
	}
}
