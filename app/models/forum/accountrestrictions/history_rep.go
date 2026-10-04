package accountrestrictions

import (
	"errors"

	"gorm.io/gorm"
)

// EnsureHistory lets a repeated administrative save finish an interrupted audit write.
func EnsureHistory(entry *History) error {
	var latest History
	err := builder().Where("user_id = ?", entry.UserId).Order("id DESC").First(&latest).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	sameUntil := latest.Until == nil && entry.Until == nil || latest.Until != nil && entry.Until != nil && latest.Until.Equal(*entry.Until)
	if err == nil && latest.ActorId == entry.ActorId && latest.Status == entry.Status && latest.Reason == entry.Reason && sameUntil {
		return nil
	}
	return CreateHistory(entry)
}

func CreateHistory(entry *History) error { return builder().Create(entry).Error }

func HistoryPage(userID uint64, page, size int) ([]History, int64, error) {
	query := builder().Model(&History{}).Where("user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []History
	err := query.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&records).Error
	return records, total, err
}
