package docOperationLogs

import (
	"time"
)

const tableName = "doc_operation_logs"

// pid 日志ID
const pid = "id"

// fieldEntityType 实体类型
const fieldEntityType = "entity_type"

// fieldEntityId 实体ID
const fieldEntityId = "entity_id"

// fieldOperationType 操作类型(create/update/delete)
const fieldOperationType = "operation_type"

// fieldFieldName 字段名称
const fieldFieldName = "field_name"

// fieldOldValue 旧值
const fieldOldValue = "old_value"

// fieldNewValue 新值
const fieldNewValue = "new_value"

// fieldMetadata 扩展元数据
const fieldMetadata = "metadata"

// fieldUserId 操作用户ID
const fieldUserId = "user_id"

// fieldIpAddress 操作IP地址
const fieldIpAddress = "ip_address"

// fieldUserAgent 用户代理
const fieldUserAgent = "user_agent"

// fieldCreatedAt 操作时间
const fieldCreatedAt = "created_at"

type Entity struct {
	Id            uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                           // 日志ID
	EntityType    string    `gorm:"column:entity_type;type:varchar(50);not null;default:'';" json:"entityType"`       // 实体类型
	EntityId      uint64    `gorm:"column:entity_id;type:bigint unsigned;not null;default:0;" json:"entityId"`        // 实体ID
	OperationType string    `gorm:"column:operation_type;type:varchar(20);not null;default:'';" json:"operationType"` // 操作类型(create/update/delete)
	FieldName     string    `gorm:"column:field_name;type:varchar(100);" json:"fieldName"`                            // 字段名称
	OldValue      string    `gorm:"column:old_value;type:text;" json:"oldValue"`                                      // 旧值
	NewValue      string    `gorm:"column:new_value;type:text;" json:"newValue"`                                      // 新值
	Metadata      string    `gorm:"column:metadata;type:json;" json:"metadata"`                                       // 扩展元数据
	UserId        uint64    `gorm:"column:user_id;type:bigint unsigned;not null;default:0;" json:"userId"`            // 操作用户ID
	IpAddress     string    `gorm:"column:ip_address;type:varchar(45);" json:"ipAddress"`                             // 操作IP地址
	UserAgent     string    `gorm:"column:user_agent;type:text;" json:"userAgent"`                                    // 用户代理
	CreatedAt     time.Time `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`               // 创建时间
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
