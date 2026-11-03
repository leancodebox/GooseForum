package attachments

import "time"

// Info is immutable display metadata copied into the post alongside its content.
type Info struct {
	ID       uint64 `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	URL      string `json:"url"`
}
type Entity struct {
	ID         uint64 `gorm:"primaryKey"`
	UserID     uint64 `gorm:"not null;index"`
	PostID     uint64 `gorm:"not null;default:0;index"`
	ObjectName string `gorm:"size:256;not null;uniqueIndex"`
	Filename   string `gorm:"size:255;not null"`
	Size       int64  `gorm:"not null"`
	CreatedAt  time.Time
}

func (Entity) TableName() string { return "attachments" }
