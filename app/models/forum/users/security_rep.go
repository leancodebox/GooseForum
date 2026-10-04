package users

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

type MFAAuthState struct {
	AccountState `gorm:"embedded"`
}

func GetMFAAuthState(userID, version uint64) (MFAAuthState, error) {
	return GetMFAAuthStateWithDB(dbconnect.Connect(), userID, version)
}

func CompareAndAdvanceTokenVersion(userID, version uint64) (bool, error) {
	return CompareAndAdvanceTokenVersionWithDB(dbconnect.Connect(), userID, version)
}

func AdvanceTokenVersion(userID uint64) (bool, error) {
	result := dbconnect.Connect().Model(&EntityComplete{}).Where("id = ?", userID).Update("token_version", gorm.Expr("token_version + 1"))
	return result.RowsAffected == 1, result.Error
}
func GetMFAAuthStateWithDB(db *gorm.DB, userID, version uint64) (MFAAuthState, error) {
	var state MFAAuthState
	err := db.Model(&EntityComplete{}).Where("id = ? AND token_version = ?", userID, version).First(&state).Error
	return state, err
}

func CompareAndAdvanceTokenVersionWithDB(db *gorm.DB, userID, version uint64) (bool, error) {
	result := db.Model(&EntityComplete{}).Where("id = ? AND token_version = ?", userID, version).
		Update("token_version", gorm.Expr("token_version + 1"))
	return result.RowsAffected == 1, result.Error
}
