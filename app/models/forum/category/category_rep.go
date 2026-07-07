package category

import "github.com/leancodebox/GooseForum/app/bundles/queryopt"

func SaveOrCreateById(entity *Entity) int64 {
	if entity.Id == 0 {
		return builder().Create(entity).RowsAffected
	}

	return builder().Save(entity).RowsAffected
}

func Get(id uint64) (entity Entity) {
	builder().First(&entity, id)
	return
}

func All() (entities []*Entity) {
	builder().Order(queryopt.Asc("sort")).Order(queryopt.Asc("id")).Find(&entities)
	return
}
