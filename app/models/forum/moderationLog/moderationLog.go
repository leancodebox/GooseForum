package moderationLog

import "time"

const tableName = "moderation_logs"

const (
	SubjectArticle  = "article"
	SubjectCategory = "category"
	SubjectUser     = "user"
	SubjectSystem   = "system"
)

const (
	ActionArticleBlocked           = "articleBlocked"
	ActionArticleUnblocked         = "articleUnblocked"
	ActionCategoryModeratorAdded   = "categoryModeratorAdded"
	ActionCategoryModeratorRemoved = "categoryModeratorRemoved"
)

const (
	fieldActorUserId = "actor_user_id"
	fieldSubjectType = "subject_type"
	fieldSubjectId   = "subject_id"
)

type Payload struct {
	MessageCode string         `json:"messageCode"`
	Params      map[string]any `json:"params,omitempty"`
}

type Entity struct {
	Id          uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	ActorUserId uint64    `gorm:"column:actor_user_id;type:bigint unsigned;not null;default:0;index:idx_moderation_logs_actor_id,priority:1;" json:"actorUserId"`
	Action      string    `gorm:"column:action;type:varchar(64);not null;default:'';" json:"action"`
	SubjectType string    `gorm:"column:subject_type;type:varchar(32);not null;default:'system';index:idx_moderation_logs_subject,priority:1;" json:"subjectType"`
	SubjectId   uint64    `gorm:"column:subject_id;type:bigint unsigned;not null;default:0;index:idx_moderation_logs_subject,priority:2;" json:"subjectId"`
	Payload     Payload   `gorm:"column:payload;type:json;serializer:json" json:"payload"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;<-:create;index:idx_moderation_logs_id_desc,priority:1;" json:"createdAt"`
}

func (itself *Entity) TableName() string {
	return tableName
}
