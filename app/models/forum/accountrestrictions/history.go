package accountrestrictions

import "time"

type History struct {
	Id        uint64     `gorm:"primaryKey;autoIncrement;index:idx_restriction_history_page,priority:2" json:"id"`
	UserId    uint64     `gorm:"not null;index:idx_restriction_user_id,priority:1;index:idx_restriction_history_page,priority:1" json:"userId"`
	ActorId   uint64     `gorm:"not null" json:"actorId"`
	Status    string     `gorm:"type:varchar(16);not null" json:"status"`
	Until     *time.Time `json:"until"`
	Reason    string     `gorm:"type:varchar(500);not null" json:"reason"`
	Note      string     `gorm:"type:varchar(2000);not null" json:"note"`
	CreatedAt time.Time  `gorm:"autoCreateTime;index:idx_restriction_user_id,priority:2" json:"createdAt"`
}

func (History) TableName() string { return "user_restriction_history" }
