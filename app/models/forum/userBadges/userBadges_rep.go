package userBadges

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
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
	if userID == 0 || badgeCode == "" {
		return false, nil
	}
	entity := Entity{
		UserId:    userID,
		BadgeCode: badgeCode,
		Source:    SourceAuto,
		Reason:    reason,
		Metadata:  metadata,
		GrantedAt: time.Now(),
	}
	result := builder().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "badge_code"}},
		DoNothing: true,
	}).Create(&entity)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func GrantManual(userID uint64, badgeCode string, source string, reason string, grantedBy uint64, metadata Metadata) (bool, error) {
	if userID == 0 || badgeCode == "" {
		return false, nil
	}
	now := time.Now()
	var existing Entity
	err := builder().
		Where(queryopt.Eq("user_id", userID)).
		Where(queryopt.Eq("badge_code", badgeCode)).
		First(&existing).Error
	if err == nil {
		if existing.RevokedAt == nil {
			return false, nil
		}
		result := builder().Model(&Entity{}).
			Where(queryopt.Eq("id", existing.Id)).
			Updates(map[string]any{
				"source":     source,
				"reason":     reason,
				"metadata":   metadata,
				"granted_by": grantedBy,
				"granted_at": now,
				"revoked_at": nil,
			})
		return result.RowsAffected > 0, result.Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	entity := Entity{
		UserId:    userID,
		BadgeCode: badgeCode,
		Source:    source,
		Reason:    reason,
		Metadata:  metadata,
		GrantedBy: grantedBy,
		GrantedAt: now,
	}
	result := builder().Create(&entity)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return false, nil
	}
	return result.RowsAffected > 0, result.Error
}

func HasActive(userID uint64, badgeCode string) bool {
	var entity Entity
	return builder().
		Select("id").
		Where(queryopt.Eq("user_id", userID)).
		Where(queryopt.Eq("badge_code", badgeCode)).
		Where("revoked_at IS NULL").
		Limit(1).
		First(&entity).Error == nil && entity.Id != 0
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
	builder().
		Where(queryopt.Eq("user_id", userID)).
		Where("revoked_at IS NULL").
		Order("granted_at DESC, id DESC").
		Find(&entities)
	return
}

func GetActiveByUserIDs(userIDs []uint64) (entities []*Entity) {
	if len(userIDs) == 0 {
		return
	}
	builder().
		Where("user_id IN ?", userIDs).
		Where("revoked_at IS NULL").
		Order("granted_at DESC, id DESC").
		Find(&entities)
	return
}

func Revoke(userID uint64, badgeCode string) error {
	now := time.Now()
	return builder().
		Where(queryopt.Eq("user_id", userID)).
		Where(queryopt.Eq("badge_code", badgeCode)).
		Where("revoked_at IS NULL").
		Update("revoked_at", now).Error
}

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}

func FirstActive(userID uint64, badgeCode string) (Entity, error) {
	var entity Entity
	err := builder().
		Where(queryopt.Eq("user_id", userID)).
		Where(queryopt.Eq("badge_code", badgeCode)).
		Where("revoked_at IS NULL").
		First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity, nil
	}
	return entity, err
}
