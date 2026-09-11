package sensitiveWord

import "time"

const tableName = "sensitive_words"

type Entity struct {
	Id          uint64    `gorm:"primaryKey;column:id;autoIncrement;not null" json:"id"`
	Word        string    `gorm:"column:word;type:varchar(128);not null;uniqueIndex" json:"word"`
	Action      string    `gorm:"column:action;type:varchar(16);not null;default:'reject'" json:"action"`
	Replacement string    `gorm:"column:replacement;type:varchar(128);not null;default:''" json:"replacement"`
	Enabled     bool      `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;<-:create" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

const (
	ActionReject  = "reject"
	ActionReplace = "replace"
	ActionRecord  = "record"
)

func (e *Entity) TableName() string { return tableName }
