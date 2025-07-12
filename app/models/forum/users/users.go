package users

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"

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
	bytes, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	r, err := jsonopt.DecodeE[ExternalInformation](bytes)
	*itself = r
	return err
}

type EntityComplete struct {
	// base
	Id          uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                      //
	Username    string     `gorm:"column:username;index;type:varchar(64);not null;default:'';" json:"username"` //
	Email       string     `gorm:"column:email;index;type:varchar(128);not null;default:'';" json:"email"`      //
	Password    string     `gorm:"column:password;type:varchar(128);not null;default:'';" json:"password"`      //
	Status      int8       `gorm:"column:status;type:tinyint;not null;default:0;" json:"status"`                // 状态：0正常 1冻结
	Validate    int8       `gorm:"column:validate;type:tinyint;not null;default:0;" json:"validate"`            // 是否验证通过: 0未通过/未验证 1 验证通过
	ActivatedAt *time.Time `gorm:"column:activated_at;type:datetime;" json:"activatedAt"`                       // 激活时间

	// info
	Nickname            string              `gorm:"column:nickname;type:varchar(64);not null;default:'';" json:"nickname"`                  //
	RoleId              uint64              `gorm:"column:role_id;type:bigint unsigned;not null;default:0;" json:"roleId"`                  //
	Prestige            int64               `gorm:"column:prestige;type:bigint;not null;default:0;" json:"prestige"`                        // 声望
	AvatarUrl           string              `gorm:"column:avatar_url;type:varchar(255);" json:"avatarUrl"`                                  // 头像URL
	Bio                 string              `gorm:"column:bio;type:varchar(500);not null;default:'';" json:"bio"`                           // 个人简介
	Signature           string              `gorm:"column:signature;type:varchar(255);not null;default:'';" json:"signature"`               // 署名
	WebsiteName         string              `gorm:"column:website_name;type:varchar(64);not null;default:'';" json:"websiteName"`           // 个人网站名
	Website             string              `gorm:"column:website;type:varchar(255);not null;default:'';" json:"website"`                   // 个人网站
	ExternalInformation ExternalInformation `gorm:"column:external_information;type:varchar(2048);default:'{}'" json:"externalInformation"` // 外部信息

	// status
	CreatedAt time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"` //
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;index;" json:"deletedAt"` //
}

func (itself *EntityComplete) GetWebAvatarUrl() string {
	if itself.AvatarUrl == "" {
		return urlconfig.GetDefaultAvatar()
	}
	if strings.HasPrefix(itself.AvatarUrl, "/static/pic/") {
		return itself.AvatarUrl
	}
	return strings.ReplaceAll(urlconfig.FilePath(itself.AvatarUrl), "\\", "/")
}

func (itself *EntityComplete) TableName() string {
	return tableName
}

func (itself *EntityComplete) SetPassword(password string) *EntityComplete {
	itself.Password, _ = algorithm.MakePassword(password)
	return itself
}

func (itself *EntityComplete) Activate() error {
	itself.Validate = 1
	activatedAt := time.Now()
	itself.ActivatedAt = &activatedAt
	return Save(itself)
}
