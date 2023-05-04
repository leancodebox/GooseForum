package Users

import (
	"github.com/leancodebox/goose/luckrand"
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

type Users struct {
	Id        uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                             //
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime;type:datetime;" json:"createdAt"`                   //
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime;type:datetime;" json:"updatedAt"`                   //
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;" json:"-"`                                          //
	Username  string     `gorm:"column:username;type:varchar(255);uniqueIndex;not null;default:'';" json:"username"` //
	Email     string     `gorm:"column:email;type:varchar(255);uniqueIndex;not null;default:'';" json:"email"`       //
	Password  string     `gorm:"column:password;type:varchar(255);not null;default:'';" json:"-"`                    //
}

// func (itself *Users) BeforeSave(tx *gorm.DB) (err error) {}
// func (itself *Users) BeforeCreate(tx *gorm.DB) (err error) {}
// func (itself *Users) AfterCreate(tx *gorm.DB) (err error) {}
// func (itself *Users) BeforeUpdate(tx *gorm.DB) (err error) {}
// func (itself *Users) AfterUpdate(tx *gorm.DB) (err error) {}
// func (itself *Users) AfterSave(tx *gorm.DB) (err error) {}
// func (itself *Users) BeforeDelete(tx *gorm.DB) (err error) {}
// func (itself *Users) AfterDelete(tx *gorm.DB) (err error) {}
// func (itself *Users) AfterFind(tx *gorm.DB) (err error) {}

func (itself *Users) TableName() string {
	return tableName
}

func (itself *Users) SetPassword(password string) *Users {
	itself.Password = luckrand.MakePassword(password)
	return itself
}
