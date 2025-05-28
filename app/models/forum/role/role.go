package role

import (
	"time"
)

const tableName = "role"

// pid
const pid = "id"

// fieldRoleName
const fieldRoleName = "role_name"

// fieldEffective
const fieldEffective = "effective"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

// fieldDeletedAt
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id        uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`             //
	RoleName  string     `gorm:"column:role_name;type:varchar(255);" json:"roleName"`                //
	Effective int        `gorm:"column:effective;type:int;not null;default:0;" json:"effective"`     //
	CreatedAt time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"` //
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"` //
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
