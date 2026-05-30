package articleWatch

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

func ListActiveUserIDsAfter(articleId, afterUserId uint64, excludeUserIds []uint64, limit int) []uint64 {
	if articleId == 0 || limit <= 0 {
		return nil
	}

	userIds := make([]uint64, 0, limit)
	query := builder().
		Select(fieldUserId).
		Where(queryopt.Eq(fieldArticleId, articleId)).
		Where(queryopt.Eq(fieldStatus, 1)).
		Where(fieldUserId+" > ?", afterUserId)
	if len(excludeUserIds) > 0 {
		query = query.Where(fieldUserId+" NOT IN ?", excludeUserIds)
	}
	query.Order(fieldUserId+" ASC").Limit(limit).Pluck(fieldUserId, &userIds)
	return userIds
}
