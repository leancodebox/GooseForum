package posts

import "time"

// Count all submitted replies, including deleted/moderated ones, to prevent quota refunds.
func DailyReplyCount(userID uint64) (int64, error) {
	var count int64
	err := builder().Where("user_id = ? AND post_no > 1 AND created_at > ?", userID, time.Now().Format(time.DateOnly)).Count(&count).Error
	return count, err
}
