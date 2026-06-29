package badges

import (
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

const tableName = "badges"

const (
	TypeSystem = "system"
	TypeCustom = "custom"

	GrantModeAuto   = "auto"
	GrantModeManual = "manual"

	IconTypeKey   = "key"
	IconTypeAsset = "asset"
)

type Entity struct {
	Id          uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	Code        string    `gorm:"column:code;type:varchar(64);not null;uniqueIndex;" json:"code"`
	Type        string    `gorm:"column:type;type:varchar(16);not null;default:'custom';index;" json:"type"`
	GrantMode   string    `gorm:"column:grant_mode;type:varchar(16);not null;default:'manual';" json:"grantMode"`
	Name        string    `gorm:"column:name;type:varchar(64);not null;default:'';" json:"name"`
	Description string    `gorm:"column:description;type:varchar(255);not null;default:'';" json:"description"`
	IconType    string    `gorm:"column:icon_type;type:varchar(16);not null;default:'key';" json:"iconType"`
	IconKey     string    `gorm:"column:icon_key;type:varchar(64);not null;default:'';" json:"iconKey"`
	IconURL     string    `gorm:"column:icon_url;type:varchar(255);not null;default:'';" json:"iconUrl"`
	Color       string    `gorm:"column:color;type:varchar(32);not null;default:'';" json:"color"`
	Level       string    `gorm:"column:level;type:varchar(32);not null;default:'';" json:"level"`
	IsEnabled   bool      `gorm:"column:is_enabled;type:boolean;not null;default:true;index;" json:"isEnabled"`
	IsWearable  bool      `gorm:"column:is_wearable;type:boolean;not null;default:true;" json:"isWearable"`
	SortOrder   int       `gorm:"column:sort_order;type:int;not null;default:0;index;" json:"sortOrder"`
	CreatedAt   time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func builder() *gorm.DB {
	return db.Connect().Table(tableName)
}

func (itself *Entity) TableName() string {
	return tableName
}
