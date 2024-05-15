package users

import (
	"github.com/leancodebox/GooseForum/bundles/algorithm"
	"time"
)

const tableName = "users"

// pid
const pid = "id"

// fieldUsername
const fieldUsername = "username"

// fieldEmail
const fieldEmail = "email"

// fieldPassword
const fieldPassword = "password"

// fieldMobileAreaCode
const fieldMobileAreaCode = "mobile_area_code"

// fieldMobilePhoneNumber
const fieldMobilePhoneNumber = "mobile_phone_number"

// fieldStatus 状态：0正常 1冻结
const fieldStatus = "status"

// fieldValidate 是否验证通过: 0未通过/未验证 1 验证通过
const fieldValidate = "validate"

// fieldPrestige 声望
const fieldPrestige = "prestige"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

// fieldDeletedAt
const fieldDeletedAt = "deleted_at"

type Entity struct {
	Id                uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                 //
	Username          string     `gorm:"column:username;type:varchar(255);not null;default:'';" json:"username"` //
	Email             string     `gorm:"column:email;type:varchar(255);not null;default:'';" json:"email"`       //
	Password          string     `gorm:"column:password;type:varchar(255);not null;default:'';" json:"password"` //
	MobileAreaCode    string     `gorm:"column:mobile_area_code;type:varchar(16);" json:"mobileAreaCode"`        //
	MobilePhoneNumber string     `gorm:"column:mobile_phone_number;type:varchar(64);" json:"mobilePhoneNumber"`  //
	Status            int8       `gorm:"column:status;type:tinyint;not null;default:0;" json:"status"`           // 状态：0正常 1冻结
	Validate          int8       `gorm:"column:validate;type:tinyint;not null;default:0;" json:"validate"`       // 是否验证通过: 0未通过/未验证 1 验证通过
	Prestige          int64      `gorm:"column:prestige;type:bigint;not null;default:0;" json:"prestige"`        // 声望
	CreatedAt         *time.Time `gorm:"column:created_at;type:datetime;" json:"createdAt"`                      //
	UpdatedAt         *time.Time `gorm:"column:updated_at;type:datetime;" json:"updatedAt"`                      //
	DeletedAt         *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"`                      //
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
