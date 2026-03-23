package imUserChatConfigs

import (
	"time"
)

const tableName = "im_user_chat_configs"

// pid 主键
const pid = "id"

// fieldUserId 当前所属用户
const fieldUserId = "user_id"

// fieldPeerId 对话对方用户
const fieldPeerId = "peer_id"

// fieldConvId 关联的conv_id
const fieldConvId = "conv_id"

// fieldUnreadCount 该用户在该对话下的未读数
const fieldUnreadCount = "unread_count"

// fieldIsPinned 是否置顶
const fieldIsPinned = "is_pinned"

// fieldIsMuted 是否免打扰
const fieldIsMuted = "is_muted"

// fieldIsDeleted 用户是否在本地删除了该对话
const fieldIsDeleted = "is_deleted"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id          uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                                                                                                      // 主键
	UserId      uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;uniqueIndex:uk_user_conv,priority:1;uniqueIndex:uk_user_peer,priority:1;index:idx_user_list,priority:1" json:"userId"` // 当前所属用户
	PeerId      uint64    `gorm:"column:peer_id;type:bigint unsigned;not null;default:0;uniqueIndex:uk_user_peer,priority:2" json:"peerId"`                                                                    // 对话对方用户
	ConvId      uint64    `gorm:"column:conv_id;type:bigint unsigned;not null;default:0;uniqueIndex:uk_user_conv,priority:2" json:"convId"`                                                                    // 关联的conv_id
	UnreadCount uint      `gorm:"column:unread_count;type:int unsigned;not null;default:0;" json:"unreadCount"`                                                                                                // 该用户在该对话下的未读数
	IsPinned    int       `gorm:"column:is_pinned;type:tinyint(1);not null;default:0;index:idx_user_list,priority:3" json:"isPinned"`                                                                          // 是否置顶
	IsMuted     int       `gorm:"column:is_muted;type:tinyint(1);not null;default:0;" json:"isMuted"`                                                                                                          // 是否免打扰
	IsDeleted   int       `gorm:"column:is_deleted;type:tinyint(1);not null;default:0;index:idx_user_list,priority:2" json:"isDeleted"`                                                                        // 用户是否在本地删除了该对话
	CreatedAt   time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime;index:idx_user_list,priority:4" json:"updatedAt"`
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

func (itself *Entity) TableName() string {
	return tableName
}
