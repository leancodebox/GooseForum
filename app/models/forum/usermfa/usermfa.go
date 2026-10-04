package usermfa

import "time"

type Factor struct {
	UserID    uint64    `gorm:"primaryKey;autoIncrement:false"`
	Secret    string    `gorm:"type:text;not null"`
	EnabledAt time.Time `gorm:"not null"`
	LastStep  int64     `gorm:"not null;default:0"`
}

func (Factor) TableName() string { return "user_mfa" }

type RecoveryCode struct {
	ID     uint64 `gorm:"primaryKey"`
	UserID uint64 `gorm:"not null;index"`
	Hash   string `gorm:"type:varchar(64);not null;uniqueIndex"`
	UsedAt *time.Time
}

func (RecoveryCode) TableName() string { return "user_mfa_recovery_codes" }
