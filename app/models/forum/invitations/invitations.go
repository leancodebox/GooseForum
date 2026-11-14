package invitations

import "time"

const tableName = "registration_invitations"

// Entity stores only a digest of the registration invitation secret.
type Entity struct {
	Id               uint64     `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	Hash             string     `gorm:"column:hash;type:varchar(64);uniqueIndex;not null;" json:"-"`
	Prefix           string     `gorm:"column:prefix;type:varchar(16);not null;index;" json:"prefix"`
	MaxUses          int        `gorm:"column:max_uses;not null;default:1;" json:"maxUses"`
	UsedCount        int        `gorm:"column:used_count;not null;default:0;" json:"usedCount"`
	CreatedBy        uint64     `gorm:"column:created_by;not null;index;" json:"createdBy"`
	ExpiresAt        *time.Time `gorm:"column:expires_at;index;" json:"expiresAt,omitempty"`
	RevokedAt        *time.Time `gorm:"column:revoked_at;index;" json:"revokedAt,omitempty"`
	ReservationHash  string     `gorm:"column:reservation_hash;type:varchar(64);index;" json:"-"`
	ReservationUntil *time.Time `gorm:"column:reservation_until;index;" json:"-"`
	LastUsedAt       *time.Time `gorm:"column:last_used_at;" json:"lastUsedAt,omitempty"`
	LastUsedUserId   uint64     `gorm:"column:last_used_user_id;" json:"lastUsedUserId,omitempty"`
	EmailTaskId      *uint64    `gorm:"column:email_task_id;index;" json:"emailTaskId,omitempty"`
	EmailTaskStatus  *uint8     `gorm:"column:email_task_status;->;" json:"emailTaskStatus,omitempty"`
	CreatedAt        time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

func (Entity) TableName() string { return tableName }
