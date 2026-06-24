package badges

import "github.com/leancodebox/GooseForum/app/bundles/queryopt"

func All() (entities []*Entity) {
	builder().
		Order("sort_order ASC, id ASC").
		Find(&entities)
	return
}

func GetByCode(code string) (entity Entity) {
	builder().Where(queryopt.Eq("code", code)).First(&entity)
	return
}

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}

func DeleteByCode(code string) error {
	return builder().Where(queryopt.Eq("code", code)).Delete(&Entity{}).Error
}
