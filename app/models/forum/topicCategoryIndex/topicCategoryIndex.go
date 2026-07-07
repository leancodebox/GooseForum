package topicCategoryIndex

import "time"

const tableName = "topic_category_index"

type Entity struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	TopicId    uint64    `gorm:"column:topic_id;type:bigint unsigned;not null;index:idx_topic_id;uniqueIndex:uniq_topic_category,priority:1;index:idx_topic_category_effective,priority:3" json:"topicId"`
	CategoryId uint64    `gorm:"column:category_id;type:bigint unsigned;not null;index;uniqueIndex:uniq_topic_category,priority:2;index:idx_topic_category_effective,priority:2;" json:"categoryId"`
	Effective  int       `gorm:"column:effective;type:int;not null;default:0;index:idx_topic_category_effective,priority:1;" json:"effective"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
