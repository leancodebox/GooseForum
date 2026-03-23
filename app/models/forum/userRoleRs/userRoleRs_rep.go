package userRoleRs

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/samber/lo"
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

//func deleteEntity(entity *Entity) int64 {
//	result := builder().Delete(entity)
//	return result.RowsAffected
//}

func GetByUserId(userId uint64) (entities []*Entity) {
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldEffective, 1)).Find(&entities)
	return
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
	roleIds := lo.Map(rs, func(t *Entity, _ int) uint64 {
		return t.RoleId
	})
	roleEntityList := role.GetByRoleIds(roleIds)
	roleMap := lo.KeyBy(roleEntityList, func(v *role.Entity) uint64 {
		return v.Id
	})
	return lo.Reduce(rs, func(res map[uint64][]role.Entity, item *Entity, _ int) map[uint64][]role.Entity {
		if roleItem, ok := roleMap[item.RoleId]; ok {
			res[item.UserId] = append(res[item.UserId], *roleItem)
		}
		return res
	}, make(map[uint64][]role.Entity, len(userIds)))
}
