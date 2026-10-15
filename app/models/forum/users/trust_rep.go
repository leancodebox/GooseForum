package users

import (
	"encoding/json"

	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"gorm.io/gorm"
)

type TrustState struct {
	AccountState      `gorm:"embedded"`
	TrustVisitDays    uint32
	TrustLastVisitDay string
	TrustTopicIds     []uint64 `gorm:"serializer:json"`
	TrustVersion      uint64
}

func GetTrustState(id uint64) (TrustState, error) {
	return getTrustState(builder(), id)
}
func getTrustState(db *gorm.DB, id uint64) (TrustState, error) {
	var s TrustState
	err := db.Model(&EntityComplete{}).Where("id = ?", id).First(&s).Error
	return s, err
}

func UpdateTrust(old, next TrustState) (bool, error) {
	return updateTrust(builder(), old, next)
}
func updateTrust(db *gorm.DB, old, next TrustState) (bool, error) {
	ids, err := json.Marshal(next.TrustTopicIds)
	if err != nil {
		return false, err
	}
	q := db.Model(&EntityComplete{}).Where("id = ? AND trust_version = ? AND trust_level = ? AND trust_level_locked = ?", old.Id, old.TrustVersion, old.TrustLevel, old.TrustLevelLocked).
		Where("restriction_status = ? AND is_activated = ? AND requires_email_verification = ?", old.RestrictionStatus, old.IsActivated, old.RequiresEmailVerification)
	if old.RestrictionUntil == nil {
		q = q.Where("restriction_until IS NULL")
	} else {
		q = q.Where("restriction_until = ?", old.RestrictionUntil)
	}
	r := q.Updates(map[string]any{"trust_level": next.TrustLevel, "trust_level_locked": next.TrustLevelLocked, "trust_visit_days": next.TrustVisitDays, "trust_last_visit_day": next.TrustLastVisitDay, "trust_topic_ids": string(ids), "trust_version": gorm.Expr("trust_version + 1")})
	return r.RowsAffected == 1, r.Error
}

type TrustIdentity struct {
	Id         uint64
	TrustLevel trustlevel.Level
}

func TrustIdentities(ids []uint64) ([]TrustIdentity, error) {
	var rows []TrustIdentity
	if len(ids) == 0 {
		return rows, nil
	}
	err := builder().Model(&EntityComplete{}).Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}
