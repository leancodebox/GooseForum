package userOAuth

import (
	"time"
)

const tableName = "user_o_auth"

// pid id
const pid = "id"

// fieldUserId 关联用户id
const fieldUserId = "user_id"

// fieldProvider 平台标识(github/twitter)
const fieldProvider = "provider"

// fieldProviderUid 第三方用户唯一ID
const fieldProviderUid = "provider_uid"

// fieldAccessToken 访问令牌
const fieldAccessToken = "access_token"

// fieldRefreshToken 刷新令牌
const fieldRefreshToken = "refresh_token"

// fieldTokenExpiry 令牌过期时间
const fieldTokenExpiry = "token_expiry"

// fieldScopes 授权范围
const fieldScopes = "scopes"

// fieldRawUserData 平台返回的原始用户数据
const fieldRawUserData = "raw_user_data"

// fieldCreatedAt 创建时间
const fieldCreatedAt = "created_at"

// fieldUpdatedAt 更新时间
const fieldUpdatedAt = "updated_at"

type Entity struct {
	Id           uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                             // id
	UserId       uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;index" json:"userId"`         // 关联用户id
	Provider     string    `gorm:"column:provider;type:varchar(32);default:0;" json:"provider"`                        // 平台标识(github/twitter)
	ProviderUid  string    `gorm:"column:provider_uid;type:varchar(255);not null;default:'';index" json:"providerUid"` // 第三方用户唯一ID
	AccessToken  string    `gorm:"column:access_token;type:varchar(1024);not null;default:'';" json:"accessToken"`     // 访问令牌
	RefreshToken string    `gorm:"column:refresh_token;type:varchar(1024);not null;default:'';" json:"refreshToken"`   // 刷新令牌
	TokenExpiry  time.Time `gorm:"column:token_expiry;type:datetime;" json:"tokenExpiry"`                              // 令牌过期时间
	Scopes       string    `gorm:"column:scopes;type:text;" json:"scopes"`                                             // 授权范围
	RawUserData  string    `gorm:"column:raw_user_data;type:text;" json:"rawUserData"`                                 // 平台返回的原始用户数据
	CreatedAt    time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`                 //
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
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
