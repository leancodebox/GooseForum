package userBadges

import (
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

const tableName = "user_badges"

const (
	SourceAuto      = "auto"
	SourceManual    = "manual"
	SourceMigration = "migration"
)

type Metadata map[string]any

type Entity struct {
	Id        uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	UserId    uint64     `gorm:"column:user_id;type:bigint unsigned;not null;default:0;uniqueIndex:idx_user_badge_code;index:idx_user_badge_granted;" json:"userId"`
	BadgeCode string     `gorm:"column:badge_code;type:varchar(64);not null;default:'';uniqueIndex:idx_user_badge_code;index;" json:"badgeCode"`
	Source    string     `gorm:"column:source;type:varchar(16);not null;default:'auto';" json:"source"`
	Reason    string     `gorm:"column:reason;type:varchar(255);not null;default:'';" json:"reason"`
	Metadata  Metadata   `gorm:"column:metadata;type:json;serializer:json" json:"metadata"`
	GrantedBy uint64     `gorm:"column:granted_by;type:bigint unsigned;not null;default:0;" json:"grantedBy"`
	GrantedAt time.Time  `gorm:"column:granted_at;index:idx_user_badge_granted;" json:"grantedAt"`
	RevokedAt *time.Time `gorm:"column:revoked_at;type:datetime;index;" json:"revokedAt"`
	CreatedAt time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func builder() *gorm.DB {
	return db.Connect().Table(tableName)
}

func (itself *Entity) TableName() string {
	return tableName
}
