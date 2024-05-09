package users

import (
	"github.com/leancodebox/GooseForum/bundles/algorithm"
	"time"
)

const tableName = "users"
const pid = "id"
const fieldCreatedAt = "created_at"
const fieldUpdatedAt = "updated_at"
const fieldDeletedAt = "deleted_at"
const fieldUsername = "username"
const fieldEmail = "email"
const fieldPassword = "password"

type Entity struct {
	Id        uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                             //
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime;type:datetime;" json:"createdAt"`                   //
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime;type:datetime;" json:"updatedAt"`                   //
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;" json:"-"`                                          //
	Username  string     `gorm:"column:username;type:varchar(255);uniqueIndex;not null;default:'';" json:"username"` //
	Email     string     `gorm:"column:email;type:varchar(255);uniqueIndex;not null;default:'';" json:"email"`       //
	Password  string     `gorm:"column:password;type:varchar(255);not null;default:'';" json:"-"`                    //
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

func (itself *Entity) SetPassword(password string) *Entity {
	itself.Password, _ = algorithm.MakePassword(password)
	return itself
}
