package authsessions

import "time"

type Token struct {
	Id              uint64     `gorm:"primaryKey;autoIncrement"`
	UserId          uint64     `gorm:"not null;index:idx_auth_session_user"`
	TokenHash       string     `gorm:"type:varchar(64);not null;uniqueIndex"`
	TokenVersion    uint64     `gorm:"not null"`
	AuthMethod      string     `gorm:"type:varchar(32);not null"`
	OAuthProvider   string     `gorm:"type:varchar(64)"`
	AuthTime        time.Time  `gorm:"not null"`
	AuthId          string     `gorm:"type:varchar(64)"`
	Reauthenticated bool       `gorm:"not null;default:false"`
	ClientIP        string     `gorm:"type:varchar(45)"`
	UserAgent       string     `gorm:"type:varchar(512)"`
	CreatedAt       time.Time  `gorm:"not null"`
	LastSeenAt      time.Time  `gorm:"not null"`
	ExpiresAt       time.Time  `gorm:"not null;index"`
	RevokedAt       *time.Time `gorm:"index"`
}

func (Token) TableName() string { return "user_auth_tokens" }

type Log struct {
	Id              uint64    `gorm:"primaryKey;autoIncrement"`
	UserId          uint64    `gorm:"not null;index"`
	UserAuthTokenId uint64    `gorm:"index"`
	Action          string    `gorm:"type:varchar(32);not null"`
	ClientIP        string    `gorm:"type:varchar(45)"`
	UserAgent       string    `gorm:"type:varchar(512)"`
	CreatedAt       time.Time `gorm:"not null;index"`
}

func (Log) TableName() string { return "user_auth_token_logs" }
