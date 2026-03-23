package userActivities

import (
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

func builder() *gorm.DB {
	return db.Connect().Table(tableName)
}

const tableName = "user_activities"

// pid 主键
const pid = "id"

// fieldUserId 发起者 ID，建立索引以便快速查询用户动态
const fieldUserId = "user_id"

// fieldAction 行为类型（如 1:发帖, 2:点赞）
const fieldAction = "action"

// fieldSubjectType 目标类型（Topic, Post, User等）
const fieldSubjectType = "subject_type"

// fieldSubjectId 目标 ID
const fieldSubjectId = "subject_id"

// fieldContentPreview 内容摘要（冗余存储以提高列表查询速度）
const fieldContentPreview = "content_preview"

// fieldCreatedAt 发生时间，建立索引用于排序
const fieldCreatedAt = "created_at"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id             uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                                                // 主键
	UserId         uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;index:idx_user_activity_created;" json:"userId"` // 发起者 ID
	Action         int       `gorm:"column:action;type:int;not null;default:0;" json:"action"`                                              // 行为类型
	SubjectType    string    `gorm:"column:subject_type;type:varchar(32);not null;default:'';" json:"subjectType"`                          // 目标类型
	SubjectId      uint64    `gorm:"column:subject_id;type:bigint unsigned;not null;default:0;" json:"subjectId"`                           // 目标 ID
	ContentPreview string    `gorm:"column:content_preview;type:text;" json:"contentPreview"`                                               // 内容摘要
	CreatedAt      time.Time `gorm:"column:created_at;index;autoCreateTime;index:idx_user_activity_created;" json:"createdAt"`              //
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
