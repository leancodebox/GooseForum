package userBadges

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Grant(userID uint64, badgeCode string, source string, reason string, grantedBy uint64, metadata Metadata) (bool, error) {
	if userID == 0 || badgeCode == "" {
		return false, nil
	}
	if source == "" {
		source = SourceAuto
	}
	if source == SourceAuto {
		return GrantAuto(userID, badgeCode, reason, metadata)
	}
	return GrantManual(userID, badgeCode, source, reason, grantedBy, metadata)
}

func GrantAuto(userID uint64, badgeCode string, reason string, metadata Metadata) (bool, error) {
	return grantWithDB(dbconnect.Connect(), userID, badgeCode, SourceAuto, reason, 0, metadata, false)
}

func GrantManual(userID uint64, badgeCode string, source string, reason string, grantedBy uint64, metadata Metadata) (bool, error) {
	return grantWithDB(dbconnect.Connect(), userID, badgeCode, source, reason, grantedBy, metadata, true)
}

func grantWithDB(conn *gorm.DB, userID uint64, badgeCode, source, reason string, grantedBy uint64, metadata Metadata, restore bool) (bool, error) {
	if userID == 0 || badgeCode == "" {
		return false, nil
	}
	changed := false
	err := conn.Transaction(func(tx *gorm.DB) error {
		// The stats row is also locked by definition deletion. A missing row cannot
		// be recreated here: it may represent a deleted badge or an unfinished migration.
		if _, err := badgeStats.LockWithDB(tx, badgeCode); err != nil {
			return err
		}
		if active, err := activeBadgeUser(tx, userID); err != nil {
			return err
		} else if !active {
			return gorm.ErrRecordNotFound
		}
		var existing Entity
		err := tx.Where("user_id = ? AND badge_code = ?", userID, badgeCode).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && (existing.RevokedAt == nil || !restore) {
			return nil
		}
		entity := Entity{UserId: userID, BadgeCode: badgeCode, Source: source, Reason: reason, Metadata: metadata, GrantedBy: grantedBy, GrantedAt: time.Now()}
		var result *gorm.DB
		if errors.Is(err, gorm.ErrRecordNotFound) {
			result = tx.Create(&entity)
		} else {
			result = tx.Model(&Entity{}).Where("id = ? AND revoked_at IS NOT NULL", existing.Id).
				Select("Source", "Reason", "Metadata", "GrantedBy", "GrantedAt", "RevokedAt").Updates(&entity)
		}
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		if err := badgeStats.AdjustWithDB(tx, badgeCode, 1, true); err != nil {
			return err
		}
		changed = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return changed, nil
}

// Keep the user row stable while deciding whether this grant contributes to the
// count. Future account deletion/restoration must maintain stats in its transaction.
func activeBadgeUser(tx *gorm.DB, userID uint64) (bool, error) {
	var user users.Identity
	query := tx.Model(&users.EntityComplete{}).Where("id = ?", userID)
	if tx.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func Exists(userID uint64, badgeCode string) bool {
	var entity Entity
	return builder().
		Select("id").
		Where(queryopt.Eq("user_id", userID)).
		Where(queryopt.Eq("badge_code", badgeCode)).
		Limit(1).
		First(&entity).Error == nil && entity.Id != 0
}

func GetActiveByUserID(userID uint64) (entities []*Entity) {
	entities, _ = GetActiveByUserIDWithError(userID)
	return
}

func GetActiveByUserIDWithError(userID uint64) ([]*Entity, error) {
	entities := make([]*Entity, 0)
	err := builder().
		Where(queryopt.Eq("user_id", userID)).
		Where("revoked_at IS NULL").
		Order("granted_at DESC, id DESC").
		Find(&entities).Error
	return entities, err
}

func GetActiveByUserIDsAndCodes(userIDs []uint64, badgeCodes []string) (entities []*Entity) {
	if len(userIDs) == 0 || len(badgeCodes) == 0 {
		return
	}
	builder().
		Where(queryopt.In("user_id", userIDs)).
		Where(queryopt.In("badge_code", badgeCodes)).
		Where("revoked_at IS NULL").
		Find(&entities)
	return
}

func Revoke(userID uint64, badgeCode string) error {
	_, err := revokeWithDB(dbconnect.Connect(), userID, badgeCode)
	return err
}

func revokeWithDB(conn *gorm.DB, userID uint64, badgeCode string) (bool, error) {
	if userID == 0 || badgeCode == "" {
		return false, nil
	}
	changed := false
	err := conn.Transaction(func(tx *gorm.DB) error {
		if _, err := badgeStats.LockWithDB(tx, badgeCode); err != nil {
			return err
		}
		active, err := activeBadgeUser(tx, userID)
		if err != nil {
			return err
		}
		result := tx.Model(&Entity{}).Where("user_id = ? AND badge_code = ? AND revoked_at IS NULL", userID, badgeCode).Update("revoked_at", time.Now())
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		// Historical records for deleted users were excluded by the migration.
		if active {
			if err := badgeStats.AdjustWithDB(tx, badgeCode, -1, false); err != nil {
				return err
			}
		}
		changed = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return changed, nil
}
