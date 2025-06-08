package users

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
)

const tableName = "users"

// pid
const pid = "id"

// fieldUsername
const fieldUsername = "username"

// fieldNickname
const fieldNickname = "nickname"

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
	Id                  uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                     //
	Username            string     `gorm:"column:username;type:varchar(255);not null;default:'';" json:"username"`     //
	Nickname            string     `gorm:"column:nickname;type:varchar(255);not null;default:'';" json:"nickname"`     //
	Email               string     `gorm:"column:email;type:varchar(255);not null;default:'';" json:"email"`           //
	Password            string     `gorm:"column:password;type:varchar(255);not null;default:'';" json:"password"`     //
	MobileAreaCode      string     `gorm:"column:mobile_area_code;type:varchar(16);" json:"mobileAreaCode"`            //
	MobilePhoneNumber   string     `gorm:"column:mobile_phone_number;type:varchar(64);" json:"mobilePhoneNumber"`      //
	RoleId              uint64     `gorm:"column:role_id;type:bigint unsigned;not null;default:0;" json:"roleId"`      //
	Status              int8       `gorm:"column:status;type:tinyint;not null;default:0;" json:"status"`               // 状态：0正常 1冻结
	Validate            int8       `gorm:"column:validate;type:tinyint;not null;default:0;" json:"validate"`           // 是否验证通过: 0未通过/未验证 1 验证通过
	ActivatedAt         time.Time  `gorm:"column:activated_at;type:datetime;" json:"activatedAt"`                      // 激活时间
	Prestige            int64      `gorm:"column:prestige;type:bigint;not null;default:0;" json:"prestige"`            // 声望
	AvatarUrl           string     `gorm:"column:avatar_url;type:varchar(255);" json:"avatarUrl"`                      // 头像URL
	Bio                 string     `gorm:"column:bio;type:varchar(500);" json:"bio"`                                   // 个人简介
	Signature           string     `gorm:"column:signature;type:varchar(255);" json:"signature"`                       // 署名
	Website             string     `gorm:"column:website;type:varchar(255);" json:"website"`                           // 个人网站
	WebsiteName         string     `gorm:"column:website_name;type:varchar(64);" json:"websiteName"`                   // 个人网站名
	ExternalInformation string     `gorm:"column:external_information;type:varchar(2048);" json:"externalInformation"` // 外部信息
	CreatedAt           time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`         //
	UpdatedAt           time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	DeletedAt           *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deletedAt"` //
}

func (itself *Entity) GetWebAvatarUrl() string {
	if itself.AvatarUrl == "" {
		return urlconfig.GetDefaultAvatar()
	}
	return strings.ReplaceAll(urlconfig.FilePath(itself.AvatarUrl), "\\", "/")
}

func (itself *Entity) GetShowName() string {
	if itself.Nickname != "" {
		return itself.Nickname
	}
	return itself.Username
}

func (itself *Entity) GetExternalInformation() ExternalInformation {
	return jsonopt.Decode[ExternalInformation](itself.ExternalInformation)
}
func (itself *Entity) SetExternalInformation(info ExternalInformation) {
	itself.ExternalInformation = jsonopt.Encode(info)
}

type ExternalInformationItem struct {
	Link string `json:"link"`
}

type ExternalInformation struct {
	Github   ExternalInformationItem `json:"github"`
	Weibo    ExternalInformationItem `json:"weibo"`
	Bilibili ExternalInformationItem `json:"bilibili"`
	Twitter  ExternalInformationItem `json:"twitter"`
	LinkedIn ExternalInformationItem `json:"linkedIn"`
	Zhihu    ExternalInformationItem `json:"zhihu"`
}

func (itself *ExternalInformation) Value() (driver.Value, error) {
	return jsonopt.EncodeE(itself)
}

func (itself *ExternalInformation) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	r, err := jsonopt.DecodeE[ExternalInformation](bytes)
	*itself = r
	return err
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

func (itself *Entity) Activate() error {
	itself.Validate = 1
	itself.ActivatedAt = time.Now()
	return Save(itself)
}
