package reports

import "time"

const tableName = "reports"

const (
	TargetArticle = "article"
	TargetReply   = "reply"
)

const (
	ReasonSpam       = "spam"
	ReasonAbuse      = "abuse"
	ReasonIllegal    = "illegal"
	ReasonIrrelevant = "irrelevant"
	ReasonOther      = "other"
)

const (
	StatusOpen     = "open"
	StatusResolved = "resolved"
	StatusRejected = "rejected"
)

const (
	ResolutionBanned  = "banned"
	ResolutionIgnored = "ignored"
)

const (
	fieldTargetType = "target_type"
	fieldTargetId   = "target_id"
	fieldReporterId = "reporter_id"
	fieldStatus     = "status"
)

type Entity struct {
	Id         uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;index:idx_reports_status_id,priority:2,sort:desc;index:idx_reports_status_article_id,priority:3,sort:desc;" json:"id"`
	TargetType string     `gorm:"column:target_type;type:varchar(32);not null;default:'';index:idx_reports_target,priority:1;index:idx_reports_reporter_target_status,priority:2;" json:"targetType"`
	TargetId   uint64     `gorm:"column:target_id;type:bigint unsigned;not null;default:0;index:idx_reports_target,priority:2;index:idx_reports_reporter_target_status,priority:3;" json:"targetId"`
	ArticleId  uint64     `gorm:"column:article_id;type:bigint unsigned;not null;default:0;index:idx_reports_status_article_id,priority:2;" json:"articleId"`
	ReporterId uint64     `gorm:"column:reporter_id;type:bigint unsigned;not null;default:0;index:idx_reports_reporter_target_status,priority:1;" json:"reporterId"`
	Reason     string     `gorm:"column:reason;type:varchar(32);not null;default:'';" json:"reason"`
	Note       string     `gorm:"column:note;type:varchar(300);not null;default:'';" json:"note"`
	Status     string     `gorm:"column:status;type:varchar(32);not null;default:'open';index:idx_reports_status_id,priority:1;index:idx_reports_status_article_id,priority:1;index:idx_reports_reporter_target_status,priority:4;" json:"status"`
	Resolution string     `gorm:"column:resolution;type:varchar(32);not null;default:'';" json:"resolution"`
	HandlerId  uint64     `gorm:"column:handler_id;type:bigint unsigned;not null;default:0;" json:"handlerId"`
	HandledAt  *time.Time `gorm:"column:handled_at;" json:"handledAt"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
