package trusthistory

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
)

type History struct {
	Id         uint64           `gorm:"primaryKey;autoIncrement;index:idx_trust_history,priority:2" json:"id"`
	UserId     uint64           `gorm:"not null;index:idx_trust_history,priority:1" json:"userId"`
	FromLevel  trustlevel.Level `gorm:"type:smallint;not null" json:"fromLevel"`
	ToLevel    trustlevel.Level `gorm:"type:smallint;not null" json:"toLevel"`
	FromLocked bool             `json:"fromLocked"`
	ToLocked   bool             `json:"toLocked"`
	Source     string           `gorm:"type:varchar(16);not null" json:"source"`
	OperatorId uint64           `json:"operatorId"`
	Reason     string           `gorm:"type:varchar(500);not null" json:"reason"`
	CreatedAt  time.Time        `json:"createdAt"`
}

func (History) TableName() string { return "user_trust_level_history" }
