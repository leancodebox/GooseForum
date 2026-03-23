package articleBookmark

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
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
	if entity.Id == 0 {
		return create(entity)
	}

	return save(entity)
}

func Get(id any) (entity Entity) {
	builder().First(&entity, id)
	return
}

func GetByArticleId(userId, articleId any) (entity Entity) {
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldArticleId, articleId)).First(&entity)
	return
}

func GetLikeReceivedCount(userId uint64) int64 {
	var count int64
	builder().
		Joins("left join articles on articles.id = article_like.article_id").
		Where("articles.user_id = ?", userId).
		Where("article_like.status = ?", 1).
		Count(&count)
	return count
}

// GetUserBookmarkedArticleIds 获取用户收藏的文章ID列表
func GetUserBookmarkedArticleIds(userId uint64, page, pageSize int) ([]uint64, int64) {
	var articleIds []uint64
	var total int64

	// 计算总数
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Count(&total)

	// 获取分页数据
	offset := (page - 1) * pageSize
	builder().
		Where(queryopt.Eq(fieldUserId, userId)).
		Where(queryopt.Eq(fieldStatus, 1)).
		Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Pluck(fieldArticleId, &articleIds)

	return articleIds, total
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
