package rolePermissionRs

import (
	array "github.com/leancodebox/goose/collectionopt"
	"github.com/leancodebox/goose/queryopt"
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

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

//func deleteEntity(entity *Entity) int64 {
//	result := builder().Delete(entity)
//	return result.RowsAffected
//}

func GetRsByRoleIdAndPermission(roleId, permissionId uint64) (entities Entity) {
	builder().
		Where(queryopt.Eq(fieldRoleId, roleId)).
		Where(queryopt.Eq(fieldPermissionId, permissionId)).
		First(&entities)
	return
}

func GetRsByRoleIds(roleIds []uint64) (entities []*Entity) {
	builder().Where(queryopt.In(fieldRoleId, roleIds)).Find(&entities)
	return
}

func GetPermissionIdsByRoleIds(roleIds []uint64) (permissionIds []uint64) {
	return array.ArrayMap(func(t *Entity) uint64 {
		return t.PermissionId
	}, GetRsByRoleIds(roleIds))
}

func GetRsGroupByRoleIds(roleIds []uint64) (result map[uint64][]uint64) {
	entityList := GetRsByRoleIds(roleIds)
	result = make(map[uint64][]uint64)
	for _, entity := range entityList {
		pIds, ok := result[entity.RoleId]
		if ok {
			result[entity.RoleId] = append(pIds, entity.PermissionId)
		} else {
			result[entity.RoleId] = []uint64{entity.PermissionId}
		}
	}
	return
}
