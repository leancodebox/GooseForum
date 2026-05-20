package badges

import "github.com/leancodebox/GooseForum/app/bundles/queryopt"

func All() (entities []*Entity) {
	builder().
		Order("sort_order ASC, id ASC").
		Find(&entities)
	return
}

func GetByCodes(codes []string) (entities []*Entity) {
	if len(codes) == 0 {
		return
	}
	builder().Where("code IN ?", codes).Find(&entities)
	return
}

func GetEnabledCustom() (entities []*Entity) {
	builder().
		Where(queryopt.Eq("type", TypeCustom)).
		Where(queryopt.Eq("is_enabled", true)).
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
