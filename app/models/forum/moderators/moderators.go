package moderators

import (
	"time"
)

const tableName = "moderators"

const (
	ScopeGlobal   = "global"
	ScopeCategory = "category"
)

const (
	StatusDisabled = 0
	StatusEnabled  = 1
)

const (
	fieldUserId    = "user_id"
	fieldScopeType = "scope_type"
	fieldScopeId   = "scope_id"
	fieldStatus    = "status"
)

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	UserId    uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;uniqueIndex:idx_moderator_scope_user;index:idx_moderator_user_status,priority:1;" json:"userId"`
	ScopeType string    `gorm:"column:scope_type;type:varchar(32);not null;default:'category';uniqueIndex:idx_moderator_scope_user;index:idx_moderator_scope_status,priority:1;" json:"scopeType"`
	ScopeId   uint64    `gorm:"column:scope_id;type:bigint unsigned;not null;default:0;uniqueIndex:idx_moderator_scope_user;index:idx_moderator_scope_status,priority:2;" json:"scopeId"`
	Status    int       `gorm:"column:status;type:int;not null;default:1;index:idx_moderator_scope_status,priority:3;index:idx_moderator_user_status,priority:2;" json:"status"`
	CreatedBy uint64    `gorm:"column:created_by;type:bigint unsigned;not null;default:0;" json:"createdBy"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
