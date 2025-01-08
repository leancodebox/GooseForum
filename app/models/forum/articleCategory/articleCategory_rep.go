package articleCategory

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/collectionopt"
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

// GetByIds 根据ID列表获取分类列表
func GetByIds(ids []uint64) (entities []*Entity) {
	if len(ids) == 0 {
		return
	}
	builder().Where("id IN ?", ids).Find(&entities)
	return
}

// GetMapByIds 根据ID列表获取分类Map
func GetMapByIds(ids []uint64) map[uint64]*Entity {
	return collectionopt.Slice2Map(GetByIds(ids), func(v *Entity) uint64 {
		return v.Id
	})
}
