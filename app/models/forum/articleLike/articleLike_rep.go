package articleLike

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
	} else {
		return save(entity)
	}
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
		Where("articles.user_id = ?", userId).Count(&count)
	return count
}

func GetLikeGivenCount(userId uint64) int64 {
	var count int64
	builder().Where(queryopt.Eq(fieldUserId, userId)).Count(&count)
	return count
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
