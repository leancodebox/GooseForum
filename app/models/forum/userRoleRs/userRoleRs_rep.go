package userRoleRs

import (
	array "github.com/leancodebox/GooseForum/app/bundles/goose/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
)

func create(entity *Entity) error {
	result := builder().Create(entity)
	return result.Error
}

func save(entity *Entity) error {
	result := builder().Save(entity)
	return result.Error
}

func SaveOrCreateById(entity *Entity) error {
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

func GetByUserId(userId uint64) (entities []*Entity) {
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldEffective, 1)).Find(&entities)
	return
}

func GetByUserIdAndRoleId(userId, roleId uint64) (entities Entity) {
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldRoleId, roleId)).First(&entities)
	return
}

func GetRoleIdsByUserId(userId uint64) []uint64 {
	return array.Map(GetByUserId(userId), func(t *Entity) uint64 {
		return t.RoleId
	})
}

func GetByUserIds(userIds []uint64) (entities []*Entity) {
	builder().Where(queryopt.In(fieldUserId, userIds)).Where(queryopt.Eq(fieldEffective, 1)).Find(&entities)
	return
}

func GetRoleGroupByUserIds(userIds []uint64) map[uint64][]role.Entity {
	rs := GetByUserIds(userIds)
	if len(rs) == 0 {
		return map[uint64][]role.Entity{}
	}
	roleIds := array.Map(rs, func(t *Entity) uint64 {
		return t.RoleId
	})
	roleEntityList := role.GetByRoleIds(roleIds)
	roleMap := array.Slice2Map(roleEntityList, func(v *role.Entity) uint64 {
		return v.Id
	})
	res := make(map[uint64][]role.Entity, len(userIds))
	for _, item := range rs {
		if roleItem, ok := roleMap[item.RoleId]; ok {
			userRoleList, urOk := res[item.UserId]
			if urOk {
				userRoleList = append(userRoleList, *roleItem)
			} else {
				userRoleList = []role.Entity{*roleItem}
			}
			res[item.UserId] = userRoleList
		}
	}
	return res
}
