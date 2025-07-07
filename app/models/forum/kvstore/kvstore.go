package kvstore

import "time"

const tableName = "kv_store"
const pid = "key"
const fieldValue = "value"
const fieldTTL = "ttl"
const fieldExpiresAt = "expires_at"
const fieldCreatedAt = "created_at"

type Entity struct {
	Key       string     `gorm:"primaryKey;column:key;not null;default:'';" json:"key"` //
	Value     string     `gorm:"column:value;type:text;;" json:"value"`                 //
	TTL       int        `gorm:"column:ttl;not null;default:0" json:"ttl"`
	ExpiresAt *time.Time `gorm:"column:expires_at;" json:"expiresAt"`                                //
	CreatedAt time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"` //
}

// func (itself *dataReps) BeforeSave(tx *gorm.DB) (err error) {}
// func (itself *dataReps) BeforeCreate(tx *gorm.DB) (err error) {}
// func (itself *dataReps) AfterCreate(tx *gorm.DB) (err error) {}
// func (itself *dataReps) BeforeUpdate(tx *gorm.DB) (err error) {}
// func (itself *dataReps) AfterUpdate(tx *gorm.DB) (err error) {}
// func (itself *dataReps) AfterSave(tx *gorm.DB) (err error) {}
// func (itself *dataReps) BeforeDelete(tx *gorm.DB) (err error) {}
// func (itself *dataReps) AfterDelete(tx *gorm.DB) (err error) {}
// func (itself *dataReps) AfterFind(tx *gorm.DB) (err error) {}

func (*Entity) TableName() string {
	return tableName
}
