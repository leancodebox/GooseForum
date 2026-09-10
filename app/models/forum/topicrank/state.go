package topicrank

import (
	"gorm.io/gorm"
	"time"
)

// State is the writable ranking projection of topics. Normal topic saves keep
// these derived fields read-only; ranking writes have no content save hooks.
type State struct {
	Id          uint64     `gorm:"column:id;primaryKey"`
	RankScore   int64      `gorm:"column:rank_score"`
	NextRankAt  *time.Time `gorm:"column:next_rank_at"`
	PublishedAt *time.Time `gorm:"column:published_at"`
}

func (State) TableName() string { return "topics" }

func query(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true, SkipDefaultTransaction: true}).Model(&State{})
}
func Save(db *gorm.DB, id uint64, score int64, next *time.Time) error {
	return query(db).Where("id = ?", id).UpdateColumns(map[string]any{"rank_score": score, "next_rank_at": next}).Error
}
func SetPublishedAt(db *gorm.DB, id uint64, at time.Time) error {
	return query(db).Where("id = ?", id).UpdateColumn("published_at", at).Error
}
func DueIDs(db *gorm.DB, now time.Time, limit int) ([]uint64, error) {
	var ids []uint64
	err := query(db).Where("next_rank_at <= ?", now).Order("next_rank_at ASC, id ASC").Limit(limit).Pluck("id", &ids).Error
	return ids, err
}
func MaxID(db *gorm.DB) (uint64, error) {
	var ids []uint64
	err := query(db).Order("id DESC").Limit(1).Pluck("id", &ids).Error
	if err != nil || len(ids) == 0 {
		return 0, err
	}
	return ids[0], nil
}
func IDsAfter(db *gorm.DB, cursor, upper uint64, limit int) ([]uint64, error) {
	var ids []uint64
	err := query(db).Where("id > ? AND id <= ?", cursor, upper).Order("id ASC").Limit(limit).Pluck("id", &ids).Error
	return ids, err
}
