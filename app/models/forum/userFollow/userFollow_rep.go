package userFollow

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
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

func GetByUserId(userId, followUserId uint64) (entity Entity) {
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldFollowUserId, followUserId)).First(&entity)
	return
}

// GetFollowingList 获取用户关注列表
func GetFollowingList(userId uint64, page, pageSize int) ([]*users.Entity, int64) {
	offset := (page - 1) * pageSize
	var userList []*users.Entity
	var total int64
	
	// 获取总数
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Count(&total)
	
	// 获取关注的用户ID列表
	var followUserIds []uint64
	builder().Select(fieldFollowUserId).Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Offset(offset).Limit(pageSize).Order(fieldCreatedAt + " DESC").Pluck(fieldFollowUserId, &followUserIds)
	
	// 根据用户ID获取用户信息
	if len(followUserIds) > 0 {
		userList = users.GetByIds(followUserIds)
	}
	
	return userList, total
}

// GetFollowerList 获取用户粉丝列表
func GetFollowerList(userId uint64, page, pageSize int) ([]*users.Entity, int64) {
	offset := (page - 1) * pageSize
	var userList []*users.Entity
	var total int64
	
	// 获取总数
	builder().Where(queryopt.Eq(fieldFollowUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Count(&total)
	
	// 获取粉丝的用户ID列表
	var followerUserIds []uint64
	builder().Select(fieldUserId).Where(queryopt.Eq(fieldFollowUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Offset(offset).Limit(pageSize).Order(fieldCreatedAt + " DESC").Pluck(fieldUserId, &followerUserIds)
	
	// 根据用户ID获取用户信息
	if len(followerUserIds) > 0 {
		userList = users.GetByIds(followerUserIds)
	}
	
	return userList, total
}
