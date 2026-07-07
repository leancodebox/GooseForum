package migrationMapping

import "time"

const tableName = "migration_mapping"

type Entity struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	Scope      string    `gorm:"column:scope;type:varchar(64);not null;default:'';uniqueIndex:uniq_migration_mapping,priority:1;index:idx_migration_mapping_target,priority:1;" json:"scope"`
	SourceType string    `gorm:"column:source_type;type:varchar(64);not null;default:'';uniqueIndex:uniq_migration_mapping,priority:2;" json:"sourceType"`
	SourceId   uint64    `gorm:"column:source_id;type:bigint unsigned;not null;default:0;uniqueIndex:uniq_migration_mapping,priority:3;" json:"sourceId"`
	TargetType string    `gorm:"column:target_type;type:varchar(64);not null;default:'';index:idx_migration_mapping_target,priority:2;" json:"targetType"`
	TargetId   uint64    `gorm:"column:target_id;type:bigint unsigned;not null;default:0;index:idx_migration_mapping_target,priority:3;" json:"targetId"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
