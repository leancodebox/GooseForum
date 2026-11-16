package badgeStats

import (
	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

const tableName = "badge_stats"

type Entity struct {
	BadgeCode       string `gorm:"column:badge_code;type:varchar(64);primaryKey;not null" json:"badgeCode"`
	GrantedCount    int64  `gorm:"column:granted_count;not null;default:0" json:"grantedCount"`
	HasGrantHistory bool   `gorm:"column:has_grant_history;not null;default:false" json:"hasGrantHistory"`
}

func (Entity) TableName() string { return tableName }

func builder() *gorm.DB { return db.Connect().Table(tableName) }
