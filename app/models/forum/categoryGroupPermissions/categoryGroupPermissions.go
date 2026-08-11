package categoryGroupPermissions

import "time"

const tableName = "category_group_permissions"

const (
	PermissionRead   int8 = 1
	PermissionReply  int8 = 2
	PermissionCreate int8 = 3
	PermissionManage int8 = 4
)

const (
	StatusDisabled int8 = 0
	StatusEnabled  int8 = 1
)

type Entity struct {
	Id              uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	CategoryId      uint64    `gorm:"column:category_id;type:bigint unsigned;not null;uniqueIndex:uniq_category_group_permission,priority:1;" json:"categoryId"`
	AccessGroupId   uint64    `gorm:"column:access_group_id;type:bigint unsigned;not null;uniqueIndex:uniq_category_group_permission,priority:2;index:idx_category_group_permissions_group_status_category,priority:1;" json:"accessGroupId"`
	PermissionLevel int8      `gorm:"column:permission_level;type:tinyint;not null;default:0;" json:"permissionLevel"`
	Status          int8      `gorm:"column:status;type:tinyint;not null;default:1;index:idx_category_group_permissions_group_status_category,priority:2;" json:"status"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
