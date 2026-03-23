package dailyStats

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StatType 统计项枚举
type StatType string

const (
	StatTypeRegCount     StatType = "reg_count"     // 注册用户数
	StatTypeArticleCount StatType = "article_count" // 帖子发布数
	StatTypeReplyCount   StatType = "reply_count"   // 回复发布数
)

// Increment 增加统计值 (Upsert)
func Increment(date time.Time, key StatType, delta int64) error {
	dateStr := date.Format("2006-01-02")

	return builder().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "stat_date"}, {Name: "stat_key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"stat_value": gorm.Expr("stat_value + ?", delta),
		}),
	}).Create(map[string]any{
		"stat_date":  dateStr,
		"stat_key":   string(key),
		"stat_value": delta,
	}).Error
}

// InitStats 初始化某天的统计项（如果不存在则创建，初始值为 0）
func InitStats(date time.Time, key StatType) error {
	dateStr := date.Format("2006-01-02")

	return builder().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(map[string]any{
		"stat_date":  dateStr,
		"stat_key":   string(key),
		"stat_value": 0,
	}).Error
}

// SetValue 直接设置统计值 (用于修复数据)
func SetValue(date time.Time, key StatType, value int64) error {
	dateStr := date.Format("2006-01-02")

	return builder().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "stat_date"}, {Name: "stat_key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"stat_value": value,
		}),
	}).Create(map[string]any{
		"stat_date":  dateStr,
		"stat_key":   string(key),
		"stat_value": value,
	}).Error
}

// GetStatsInRange 获取指定时间范围内的统计项列表
func GetStatsInRange(keys []StatType, startDate, endDate string) ([]*Entity, error) {
	var results []*Entity
	err := builder().
		Where("stat_key IN ?", keys).
		Where("stat_date >= ?", startDate).
		Where("stat_date <= ?", endDate).
		Order("stat_date ASC").
		Find(&results).Error
	return results, err
}

// GetSumInRange 获取指定统计项在一段时间内的总和
func GetSumInRange(key StatType, days int) int64 {
	var sum int64
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	builder().
		Where("stat_key = ?", string(key)).
		Where("stat_date >= ?", startDate).
		Select("COALESCE(SUM(stat_value), 0)").
		Scan(&sum)
	return sum
}

// GetCurrentMonthSum 获取指定统计项在本月内的总和
func GetCurrentMonthSum(key StatType) int64 {
	var sum int64
	// 获取本月第一天，例如 "2026-02-01"
	startDate := time.Now().Format("2006-01") + "-01"
	builder().
		Where("stat_key = ?", string(key)).
		Where("stat_date >= ?", startDate).
		Select("COALESCE(SUM(stat_value), 0)").
		Scan(&sum)
	return sum
}
