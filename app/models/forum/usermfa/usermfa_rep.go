package usermfa

import (
	"time"
)

func GetFactor(userID uint64) (Factor, error) {
	var factor Factor
	err := builder().Where("user_id = ?", userID).First(&factor).Error
	return factor, err
}

func HasFactor(userID uint64) (bool, error) {
	var count int64
	err := builder().Model(&Factor{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

func CountUnusedRecoveryCodes(userID uint64) (int64, error) {
	var count int64
	err := recoveryBuilder().Where("user_id = ? AND used_at IS NULL", userID).Count(&count).Error
	return count, err
}

func CreateFactor(factor *Factor) error { return builder().Create(factor).Error }

func DeleteFactor(userID uint64) (bool, error) {
	result := builder().Where("user_id = ?", userID).Delete(&Factor{})
	return result.RowsAffected == 1, result.Error
}

func DeleteRecoveryCodes(userID uint64) error {
	return recoveryBuilder().Where("user_id = ?", userID).Delete(&RecoveryCode{}).Error
}

func CreateRecoveryCodes(codes []RecoveryCode) error {
	return recoveryBuilder().Create(&codes).Error
}

func DeleteOtherRecoveryCodes(userID uint64, codes []RecoveryCode) error {
	hashes := make([]string, len(codes))
	for i, code := range codes {
		hashes[i] = code.Hash
	}
	return recoveryBuilder().Where("user_id = ? AND hash NOT IN ?", userID, hashes).Delete(&RecoveryCode{}).Error
}

// AcceptStep prevents replay across all challenges and security operations.
func AcceptStep(userID uint64, step int64) (bool, error) {
	result := builder().Where("user_id = ? AND last_step < ?", userID, step).Update("last_step", step)
	return result.RowsAffected == 1, result.Error
}

func ConsumeRecoveryCode(userID uint64, hash string, now time.Time) (bool, error) {
	result := recoveryBuilder().Where("user_id = ? AND hash = ? AND used_at IS NULL", userID, hash).Update("used_at", now)
	return result.RowsAffected == 1, result.Error
}
