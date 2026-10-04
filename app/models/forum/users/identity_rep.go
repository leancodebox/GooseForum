package users

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

func (user EntityComplete) NeedsEmailVerification(policyEnabled bool) bool {
	return user.IsActivated == ActivationPending && (policyEnabled || user.RequiresEmailVerification)
}

var (
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")
)

func CheckIdentityAvailable(username, email string, excludeUserID uint64) error {
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
		result := builder().Model(&EntityComplete{}).Unscoped().Where("id <> ? AND "+identity.column+" = ?", excludeUserID, identityKey(identity.value)).Limit(1).Find(&match)
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
	user.Username = identityKey(user.Username)
	user.Email = identityKey(user.Email)
	return nil
}

// Resolve a write race against the database constraint using domain errors.
func identityWriteError(err error, username, email string, excludeUserID uint64) error {
	if err == nil {
		return nil
	}
	conflict := CheckIdentityAvailable(username, email, excludeUserID)
	if errors.Is(conflict, ErrUsernameExists) || errors.Is(conflict, ErrEmailExists) {
		return conflict
	}
	return err
}
