package users

import (
	"gorm.io/gorm"
)

type MFAAuthState struct {
	AccountState `gorm:"embedded"`
}

func GetMFAAuthState(userID, version uint64) (MFAAuthState, error) {
	var state MFAAuthState
	err := builder().Model(&EntityComplete{}).Where("id = ? AND token_version = ?", userID, version).First(&state).Error
	return state, err
}

func CompareAndAdvanceTokenVersion(userID, version uint64) (bool, error) {
	result := builder().Model(&EntityComplete{}).Where("id = ? AND token_version = ?", userID, version).
		Update("token_version", gorm.Expr("token_version + 1"))
	return result.RowsAffected == 1, result.Error
}

func AdvanceTokenVersion(userID uint64) (bool, error) {
	result := builder().Model(&EntityComplete{}).Where("id = ?", userID).Update("token_version", gorm.Expr("token_version + 1"))
	return result.RowsAffected == 1, result.Error
}
