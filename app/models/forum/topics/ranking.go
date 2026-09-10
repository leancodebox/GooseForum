package topics

import "gorm.io/gorm"

// GetForRankingWithDB includes hidden/deleted topics so their scores can settle.
// Old decimal scores are not read during conversion to integer ranking.
func GetForRankingWithDB(db *gorm.DB, id uint64) (Entity, error) {
	var topic Entity
	err := db.Model(&Entity{}).Unscoped().Select("id", "status", "process_status", "deleted_at", "created_at", "published_at", "like_count", "reply_count", "last_posted_at").First(&topic, id).Error
	return topic, err
}
