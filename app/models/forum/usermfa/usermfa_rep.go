package usermfa

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

func GetFactor(userID uint64) (Factor, error) { return GetFactorWithDB(dbconnect.Connect(), userID) }

func HasFactor(userID uint64) (bool, error) {
	var count int64
	err := dbconnect.Connect().Model(&Factor{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

func CountUnusedRecoveryCodes(userID uint64) (int64, error) {
	var count int64
	err := dbconnect.Connect().Model(&RecoveryCode{}).Where("user_id = ? AND used_at IS NULL", userID).Count(&count).Error
	return count, err
}

func CreateFactor(factor *Factor) error { return CreateFactorWithDB(dbconnect.Connect(), factor) }

func DeleteFactor(userID uint64) (bool, error) {
	return DeleteFactorWithDB(dbconnect.Connect(), userID)
}

func DeleteRecoveryCodes(userID uint64) error {
	return DeleteRecoveryCodesWithDB(dbconnect.Connect(), userID)
}

func CreateRecoveryCodes(codes []RecoveryCode) error {
	return CreateRecoveryCodesWithDB(dbconnect.Connect(), codes)
}

// AcceptStep prevents replay across all challenges and security operations.
func AcceptStep(userID uint64, step int64) (bool, error) {
	return AcceptStepWithDB(dbconnect.Connect(), userID, step)
}

func ConsumeRecoveryCode(userID uint64, hash string, now time.Time) (bool, error) {
	return ConsumeRecoveryCodeWithDB(dbconnect.Connect(), userID, hash, now)
}
func GetFactorWithDB(db *gorm.DB, userID uint64) (Factor, error) {
	var factor Factor
	err := db.Where("user_id = ?", userID).First(&factor).Error
	return factor, err
}

func CreateFactorWithDB(db *gorm.DB, factor *Factor) error {
	return db.Create(factor).Error
}

func DeleteFactorWithDB(db *gorm.DB, userID uint64) (bool, error) {
	result := db.Where("user_id = ?", userID).Delete(&Factor{})
	return result.RowsAffected == 1, result.Error
}

func DeleteRecoveryCodesWithDB(db *gorm.DB, userID uint64) error {
	return db.Where("user_id = ?", userID).Delete(&RecoveryCode{}).Error
}

func CreateRecoveryCodesWithDB(db *gorm.DB, codes []RecoveryCode) error {
	return db.Create(&codes).Error
}

func AcceptStepWithDB(db *gorm.DB, userID uint64, step int64) (bool, error) {
	result := db.Model(&Factor{}).Where("user_id = ? AND last_step < ?", userID, step).Update("last_step", step)
	return result.RowsAffected == 1, result.Error
}

func ConsumeRecoveryCodeWithDB(db *gorm.DB, userID uint64, hash string, now time.Time) (bool, error) {
	result := db.Model(&RecoveryCode{}).Where("user_id = ? AND hash = ? AND used_at IS NULL", userID, hash).Update("used_at", now)
	return result.RowsAffected == 1, result.Error
}
