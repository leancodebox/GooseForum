package users

import (
	"github.com/leancodebox/GooseForum/bundles/algorithm"
	"time"
)

const tableName = "users"

// pid
const pid = "id"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

// fieldDeletedAt
const fieldDeletedAt = "deleted_at"

// fieldUsername
const fieldUsername = "username"

// fieldEmail
const fieldEmail = "email"

// fieldPassword
const fieldPassword = "password"

type Entity struct {
	Id        uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                 //
	Username  string     `gorm:"column:username;type:varchar(255);not null;default:'';" json:"username"` //
	Email     string     `gorm:"column:email;type:varchar(255);not null;default:'';" json:"email"`       //
	Password  string     `gorm:"column:password;type:varchar(255);not null;default:'';" json:"password"` //
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime;" json:"createdAt"`                      //
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime;" json:"updatedAt"`                      //
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"`                      //
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
