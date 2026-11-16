package badgeStats

import (
	"errors"
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInvalidCount = errors.New("badge grant count would be invalid")

func GetByCodes(codes []string) (map[string]Entity, error) {
	result := make(map[string]Entity, len(codes))
	if len(codes) == 0 {
		return result, nil
	}
	var rows []Entity
	if err := builder().Where("badge_code IN ?", codes).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.BadgeCode] = row
	}
	return result, nil
}

func Get(code string) (Entity, error) {
	var row Entity
	err := builder().Where("badge_code = ?", code).First(&row).Error
	return row, err
}

// EnsureWithDB initializes a new badge without scanning grants. Existing values
// are preserved; historic values are initialized by the versioned migration.
func EnsureWithDB(tx *gorm.DB, code string) error {
	if code == "" {
		return errors.New("badge code is required")
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "badge_code"}}, DoNothing: true,
	}).Create(&Entity{BadgeCode: code}).Error
}

// LockWithDB must run inside the caller's transaction, before changing grants
// or deleting a definition. Every writer locks this same persistent row.
func LockWithDB(tx *gorm.DB, code string) (Entity, error) {
	var row Entity
	if tx.Dialector.Name() == "sqlite" {
		// SQLite has no FOR UPDATE; acquire its write lock before reading.
		if err := tx.Model(&Entity{}).Where("badge_code = ?", code).
			UpdateColumn("granted_count", gorm.Expr("granted_count")).Error; err != nil {
			return row, err
		}
	}
	query := tx.Where("badge_code = ?", code)
	if tx.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&row).Error
	return row, err
}

// AdjustWithDB changes the stored count atomically in the grant transaction.
// A failed adjustment must be returned by the caller so the grant rolls back.
func AdjustWithDB(tx *gorm.DB, code string, delta int64, markHistory bool) error {
	if delta == math.MinInt64 {
		return ErrInvalidCount
	}
	query := tx.Model(&Entity{}).Where("badge_code = ? AND granted_count >= 0", code)
	if delta < 0 {
		query = query.Where("granted_count >= ?", -delta)
	} else if delta > 0 {
		query = query.Where("granted_count <= ?", math.MaxInt64-delta)
	}
	values := map[string]any{"granted_count": gorm.Expr("granted_count + ?", delta)}
	if markHistory {
		values["has_grant_history"] = true
	}
	result := query.Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	// MySQL reports zero changed rows for an idempotent zero adjustment.
	var row Entity
	if err := tx.Where("badge_code = ?", code).First(&row).Error; err != nil {
		return err
	}
	if delta == 0 && row.GrantedCount >= 0 {
		return nil
	}
	return ErrInvalidCount
}
