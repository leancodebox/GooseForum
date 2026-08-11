package accessGroups

import "time"

const tableName = "access_groups"

const (
	SystemKeyEveryone   = "everyone"
	SystemKeyRegistered = "registered"
)

const (
	JoinModeSystem      = "system"
	JoinModeInviteOnly  = "invite_only"
	JoinModeApplication = "application"
)

const (
	StatusDisabled int8 = 0
	StatusEnabled  int8 = 1
)

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(128);not null;default:'';" json:"name"`
	SystemKey *string   `gorm:"column:system_key;type:varchar(32);uniqueIndex:uniq_access_group_system_key;" json:"systemKey,omitempty"`
	JoinMode  string    `gorm:"column:join_mode;type:varchar(32);not null;default:'invite_only';" json:"joinMode"`
	Status    int8      `gorm:"column:status;type:tinyint;not null;default:1;index:idx_access_groups_status,priority:1;" json:"status"`
	CreatedBy uint64    `gorm:"column:created_by;type:bigint unsigned;not null;default:0;" json:"createdBy"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
