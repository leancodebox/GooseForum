package moderators

import "github.com/leancodebox/GooseForum/app/bundles/queryopt"

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}

func Get(id uint64) (entity Entity) {
	if id == 0 {
		return entity
	}
	builder().First(&entity, id)
	return
}

func GetByUserScope(userId uint64, scopeType string, scopeId uint64) (entity Entity) {
	if userId == 0 || scopeType == "" {
		return entity
	}
	builder().
		Where(queryopt.Eq(fieldUserId, userId)).
		Where(queryopt.Eq(fieldScopeType, scopeType)).
		Where(queryopt.Eq(fieldScopeId, scopeId)).
		First(&entity)
	return
}

func GetByCategoryIds(categoryIds []uint64) (entities []*Entity) {
	if len(categoryIds) == 0 {
		return
	}
	builder().
		Where(queryopt.Eq(fieldScopeType, ScopeCategory)).
		Where(queryopt.In(fieldScopeId, categoryIds)).
		Where(queryopt.Eq(fieldStatus, StatusEnabled)).
		Order("id ASC").
		Find(&entities)
	return
}

func AllEnabled() (entities []*Entity) {
	builder().
		Where(queryopt.Eq(fieldStatus, StatusEnabled)).
		Order("id ASC").
		Find(&entities)
	return
}

func Delete(entity *Entity) error {
	return builder().Delete(entity).Error
}
