package users

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"

	"gorm.io/gorm"
)

// The key identifies the provider identity, never a matching email address.
func OAuthRegistrationKey(provider, uid string) string {
	encoded, _ := json.Marshal([2]string{provider, uid})
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func FindOAuthRegistration(key string) (*EntityComplete, error) {
	var user EntityComplete
	err := dbconnect.Connect().Where("oauth_registration_key = ?", key).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func CompleteOAuthRegistration(id uint64) error {
	return dbconnect.Connect().Model(&EntityComplete{}).Where("id = ?", id).Update("oauth_registration_key", nil).Error
}
