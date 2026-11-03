package attachments

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInvalid = errors.New("invalid attachment reference")

func builder() *gorm.DB                   { return dbconnect.Connect().Model(&Entity{}) }
func Create(e *Entity) error              { return builder().Create(e).Error }
func Get(id uint64) (e Entity, err error) { err = builder().First(&e, id).Error; return }
func GetMany(ids []uint64) (rows []Entity, err error) {
	if len(ids) == 0 {
		return
	}
	err = builder().Where("id IN ?", ids).Find(&rows).Error
	return
}

// Claims are permanent: removing a file from content never makes it reusable by another post.
func BindWithDB(tx *gorm.DB, userID, postID uint64, files []Info) error {
	if len(files) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(files))
	seen := make(map[uint64]bool, len(files))
	for _, f := range files {
		if seen[f.ID] {
			return ErrInvalid
		}
		seen[f.ID] = true
		ids = append(ids, f.ID)
	}
	var rows []Entity
	if err := tx.Table((Entity{}).TableName()).Model(&Entity{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id").Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) != len(ids) {
		return ErrInvalid
	}
	unbound := make([]uint64, 0, len(rows))
	for _, row := range rows {
		if row.UserID != userID || (row.PostID != 0 && row.PostID != postID) {
			return ErrInvalid
		}
		if row.PostID == postID {
			continue
		}
		unbound = append(unbound, row.ID)
	}
	if len(unbound) == 0 {
		return nil
	}
	result := tx.Table((Entity{}).TableName()).Model(&Entity{}).Where("id IN ? AND user_id = ? AND post_id = 0", unbound, userID).Update("post_id", postID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(len(unbound)) {
		return ErrInvalid
	}
	return nil
}
func UnboundBefore(before time.Time) (rows []Entity, err error) {
	err = builder().Where("post_id = 0 AND created_at < ?", before).Limit(100).Find(&rows).Error
	return
}

// Claim deletion first, so a concurrent content save either claims the row or fails atomically.
func DeleteUnbound(id, userID uint64) (bool, error) {
	r := builder().Where("id = ? AND user_id = ? AND post_id = 0", id, userID).Delete(&Entity{})
	return r.RowsAffected == 1, r.Error
}
