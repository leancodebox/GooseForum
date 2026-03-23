package rolePermissionRs

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/samber/lo"
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

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

func DeleteEntity(entity *Entity) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

func GetRsByRoleIdAndPermission(roleId, permissionId uint64) (entities Entity) {
	builder().
		Where(queryopt.Eq(fieldRoleId, roleId)).
		Where(queryopt.Eq(fieldPermissionId, permissionId)).
		First(&entities)
	return
}

func GetRsByRoleId(roleIds uint64) (entities []*Entity) {
	builder().Where(queryopt.Eq(fieldRoleId, roleIds)).Find(&entities)
	return
}

func GetRsByRoleIds(roleIds []uint64) (entities []*Entity) {
	builder().Where(queryopt.In(fieldRoleId, roleIds)).Find(&entities)
	return
}

func GetPermissionIdsByRoleIds(roleIds []uint64) (permissionIds []uint64) {
	return lo.Map(GetRsByRoleIds(roleIds), func(t *Entity, _ int) uint64 {
		return t.PermissionId
	})
}

func GetRsGroupByRoleIds(roleIds []uint64) map[uint64][]uint64 {
	entityList := GetRsByRoleIds(roleIds)
	return lo.MapValues(lo.GroupBy(entityList, func(e *Entity) uint64 {
		return e.RoleId
	}), func(items []*Entity, _ uint64) []uint64 {
		return lo.Map(items, func(e *Entity, _ int) uint64 {
			return e.PermissionId
		})
	})
}
