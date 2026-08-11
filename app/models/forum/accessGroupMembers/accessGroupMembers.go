package accessGroupMembers

import "time"

const tableName = "access_group_members"

const (
	MemberRoleMember  = "member"
	MemberRoleManager = "manager"
)

const (
	StatusDisabled int8 = 0
	StatusEnabled  int8 = 1
	StatusPending  int8 = 2
)

type Entity struct {
	Id            uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	AccessGroupId uint64    `gorm:"column:access_group_id;type:bigint unsigned;not null;uniqueIndex:uniq_access_group_member,priority:1;" json:"accessGroupId"`
	UserId        uint64    `gorm:"column:user_id;type:bigint unsigned;not null;uniqueIndex:uniq_access_group_member,priority:2;index:idx_access_group_members_user_status_group,priority:1;" json:"userId"`
	MemberRole    string    `gorm:"column:member_role;type:varchar(32);not null;default:'member';" json:"memberRole"`
	Status        int8      `gorm:"column:status;type:tinyint;not null;default:1;index:idx_access_group_members_user_status_group,priority:2;" json:"status"`
	CreatedBy     uint64    `gorm:"column:created_by;type:bigint unsigned;not null;default:0;" json:"createdBy"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
