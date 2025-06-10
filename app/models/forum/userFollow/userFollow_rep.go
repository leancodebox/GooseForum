package userFollow

import "github.com/leancodebox/GooseForum/app/bundles/queryopt"

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func SaveOrCreateById(entity *Entity) int64 {
	if entity.UserId == 0 {
		return create(entity)
	} else {
		return save(entity)
	}
}

func Get(id any) (entity Entity) {
	builder().First(&entity, id)
	return
}

func GetFollowingCount(userId uint64) int64 {
	var count int64
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Count(&count)
	return count
}

func GetFollowerCount(userId uint64) int64 {
	var count int64
	builder().Where(queryopt.Eq(fieldFollowUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Count(&count)
	return count
}
