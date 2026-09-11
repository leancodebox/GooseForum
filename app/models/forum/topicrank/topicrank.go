package topicrank

import "time"

const tableName = "topic_rank_schedule"

// Keep settled rows so versions cannot reset after deleting/recreating a task.
type Entity struct {
	TopicID   uint64     `gorm:"column:topic_id;primaryKey;autoIncrement:false"`
	NextRunAt *time.Time `gorm:"column:next_run_at;index:idx_topic_rank_due,priority:1"`
	Version   uint64     `gorm:"column:version;not null;default:1"`
}

func (Entity) TableName() string { return tableName }

type ScheduledTopic struct {
	ID      uint64    `gorm:"column:topic_id"`
	At      time.Time `gorm:"column:next_run_at"`
	Version uint64    `gorm:"column:version"`
}
