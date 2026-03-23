package dailyStats

import (
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

func builder() *gorm.DB {
	return db.Connect().Table(tableName)
}

const tableName = "daily_stats"

// pid 日期，格式 YYYY-MM-DD
const pid = "stat_date"

// pid 统计项的 Key (带前缀)
const stat_pid = "stat_key"

// fieldStatValue 统计数值
const fieldStatValue = "stat_value"

type Entity struct {
	StatDate  time.Time `gorm:"primaryKey;column:stat_date;type:date;not null;" json:"statDate"`       // 日期，格式 YYYY-MM-DD
	StatKey   string    `gorm:"primaryKey;column:stat_key;type:varchar(128);not null;" json:"statKey"` // 统计项的 Key (带前缀)
	StatValue int64     `gorm:"column:stat_value;type:bigint;not null;default:0;" json:"statValue"`    // 统计数值
}

func (itself *Entity) TableName() string {
	return tableName
}

// func (itself *Entity) BeforeSave(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeCreate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterCreate(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeUpdate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterUpdate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterSave(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeDelete(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterDelete(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterFind(tx *gorm.DB) (err error) {}
