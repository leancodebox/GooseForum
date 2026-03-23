package articlesUserStat

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func SaveOrCreateById(entity *Entity) int64 {
	if entity.ArticleId == 0 {
		return create(entity)
	}

	return save(entity)
}

func Get(id any) (entity Entity) {
	builder().First(&entity, id)
	return
}

// IncrementUserReply 增加评论计数
func IncrementUserReply(articleId, userId uint64) error {
	now := time.Now()

	// 1. 更新统计表 (Upsert)
	err := builder().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "article_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"reply_count":   gorm.Expr("reply_count + 1"),
			"last_reply_at": now,
		}),
	}).Create(map[string]any{
		"article_id":    articleId,
		"user_id":       userId,
		"reply_count":   1,
		"last_reply_at": now,
	}).Error

	return err
}

// DecrementUserReply 减少评论计数 (用于删除评论)
func DecrementUserReply(articleId, userId uint64) error {
	// 1. 减少计数
	// 注意：SQLite 不支持直接在 Update 里写 Case When 逻辑太复杂，我们先尝试减 1
	result := builder().
		Where("article_id = ? AND user_id = ? AND reply_count > 0", articleId, userId).
		Update("reply_count", gorm.Expr("reply_count - 1"))

	return result.Error
}

// SyncArticlePosters 重新计算并刷新文章表的头像 JSON
func SyncArticlePosters(articleId uint64) []uint64 {
	var activeUserIDs []uint64
	// 利用我们之前设计的 idx_active_query 索引进行极速查询
	builder().
		Where("article_id = ?", articleId).
		Order("reply_count DESC, last_reply_at DESC").
		Limit(3).
		Pluck("user_id", &activeUserIDs)
	return activeUserIDs
}

func FixStat(articleId, userId uint64, count uint32, lastReplyAt time.Time) error {
	return builder().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "article_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"reply_count":   count,
			"last_reply_at": lastReplyAt,
		}),
	}).Create(map[string]any{
		"article_id":    articleId,
		"user_id":       userId,
		"reply_count":   count,
		"last_reply_at": lastReplyAt,
	}).Error
}

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

//func deleteEntity(entity *Entity) int64 {
//	result := builder().Delete(entity)
//	return result.RowsAffected
//}

//func all() (entities []*Entity) {
//	builder().Find(&entities)
//	return
//}
