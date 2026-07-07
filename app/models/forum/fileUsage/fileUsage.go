package fileUsage

import "time"

const tableName = "file_usages"

const (
	TargetTopic       = "topic"
	TargetPost        = "post"
	TargetUser        = "user"
	TargetAdminUpload = "admin_upload"
)

const (
	UsageInlineImage = "inline_image"
	UsageAvatar      = "avatar"
	UsageAdminUpload = "admin_upload"
)

type Entity struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	FileName   string    `gorm:"column:file_name;type:varchar(512);not null;default:'';uniqueIndex:idx_file_usage_target_file,priority:4;" json:"fileName"`
	TargetType string    `gorm:"column:target_type;type:varchar(32);not null;default:'';uniqueIndex:idx_file_usage_target_file,priority:1;" json:"targetType"`
	TargetId   uint64    `gorm:"column:target_id;type:bigint unsigned;not null;default:0;uniqueIndex:idx_file_usage_target_file,priority:2;" json:"targetId"`
	UsageType  string    `gorm:"column:usage_type;type:varchar(32);not null;default:'';uniqueIndex:idx_file_usage_target_file,priority:3;" json:"usageType"`
	UserId     uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
