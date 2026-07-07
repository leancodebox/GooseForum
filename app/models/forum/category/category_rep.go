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

func Count() int64 {
	var count int64
	builder().Count(&count)
	return count
}

func All() (entities []*Entity) {
	builder().Order(queryopt.Asc("sort")).Order(queryopt.Asc("id")).Find(&entities)
	return
}

func DeleteEntity(entity *Entity) int64 {
	return builder().Delete(entity).RowsAffected
}
