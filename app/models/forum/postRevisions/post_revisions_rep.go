package postRevisions

import "gorm.io/gorm"

const RetentionLimit = 10

// PruneWithDB uses a bounded index lookup followed by a portable delete.
// Concurrent inserts may temporarily exceed the limit; the next edit retries.
func PruneWithDB(db *gorm.DB, postID uint64) error {
	var cutoff []uint64
	if err := db.Model(&Entity{}).Where("post_id = ?", postID).
		Order("revision_no DESC").Offset(RetentionLimit-1).Limit(1).
		Pluck("revision_no", &cutoff).Error; err != nil {
		return err
	}
	if len(cutoff) == 0 {
		return nil
	}
	return db.Session(&gorm.Session{SkipDefaultTransaction: true}).Where("post_id = ? AND revision_no < ?", postID, cutoff[0]).Delete(&Entity{}).Error
}

// PruneAllWithDB includes orphaned/soft-deleted posts and never loads bodies.
func PruneAllWithDB(db *gorm.DB) error {
	var cursor uint64
	for {
		var ids []uint64
		if err := db.Model(&Entity{}).Where("post_id > ?", cursor).Distinct("post_id").Order("post_id").Limit(200).Pluck("post_id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		for _, id := range ids {
			if err := PruneWithDB(db, id); err != nil {
				return err
			}
		}
		cursor = ids[len(ids)-1]
	}
}
