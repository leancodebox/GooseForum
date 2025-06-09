package articleCategory

import (
	"fmt"
)

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	fmt.Println(result.Error)
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
	if id == 0 {
		return entity
	}
	builder().First(&entity, id)
	return
}

func SaveAll(entities *[]Entity) int64 {
	result := builder().Save(entities)
	return result.RowsAffected
}

func DeleteEntity(entity *Entity) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

func All() (entities []*Entity) {
	builder().Find(&entities)
	return
}

func Count() int64 {
	var total int64
	builder().Count(&total)
	return total
}
