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
	}

	return save(entity)
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

// IsFollowing 判断用户是否关注了某个用户
func IsFollowing(userId, followUserId uint64) bool {
	if userId == 0 || followUserId == 0 {
		return false
	}
	var count int64
	builder().Where(queryopt.Eq(fieldUserId, userId)).
		Where(queryopt.Eq(fieldFollowUserId, followUserId)).
		Where(queryopt.Eq(fieldStatus, 1)).
		Count(&count)
	return count > 0
}

// GetAll 用于全量导出/修复数据，支持分页查询
func GetAll(offset, limit int) ([]*Entity, error) {
	var entities []*Entity
	err := builder().Offset(offset).Limit(limit).Order("id ASC").Find(&entities).Error
	return entities, err
}

// GetAllFollowingIds 获取用户所有关注的用户ID列表
func GetAllFollowingIds(userId uint64) []uint64 {
	if userId == 0 {
		return make([]uint64, 0)
	}
	var followUserIds []uint64
	builder().Select(fieldFollowUserId).Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Pluck(fieldFollowUserId, &followUserIds)
	return followUserIds
}

// GetFollowingList 获取用户关注列表
func GetFollowingList(userId uint64, page, pageSize int) []*users.EntityComplete {
	offset := (page - 1) * pageSize
	var userList []*users.EntityComplete

	// 获取关注的用户ID列表
	var followUserIds []uint64
	builder().Select(fieldFollowUserId).Where(queryopt.Eq(fieldUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Offset(offset).Limit(pageSize).Order(fieldCreatedAt+" DESC").Pluck(fieldFollowUserId, &followUserIds)

	// 根据用户ID获取用户信息
	if len(followUserIds) > 0 {
		userList = users.GetByIds(followUserIds)
	}

	return userList
}

// GetFollowerList 获取用户粉丝列表
func GetFollowerList(userId uint64, page, pageSize int) []*users.EntityComplete {
	offset := (page - 1) * pageSize
	var userList []*users.EntityComplete

	// 获取粉丝的用户ID列表
	var followerUserIds []uint64
	builder().Select(fieldUserId).Where(queryopt.Eq(fieldFollowUserId, userId)).Where(queryopt.Eq(fieldStatus, 1)).Offset(offset).Limit(pageSize).Order(fieldCreatedAt+" DESC").Pluck(fieldUserId, &followerUserIds)

	// 根据用户ID获取用户信息
	if len(followerUserIds) > 0 {
		userList = users.GetByIds(followerUserIds)
	}

	return userList
}
