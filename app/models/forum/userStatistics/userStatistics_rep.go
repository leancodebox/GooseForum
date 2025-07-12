package userStatistics

import "time"

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

func WriteArticle(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET article_count = article_count+1 where user_id = ?", userId)
	return result.RowsAffected
}

func WriteComment(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET reply_count = reply_count+1 where user_id = ?", userId)
	return result.RowsAffected
}

func LikeArticle(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET like_received_count = like_received_count+1 where user_id = ?", userId)
	return result.RowsAffected
}

func GivenLike(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET like_given_count = like_given_count+1 where user_id = ?", userId)
	return result.RowsAffected
}

func CancelLikeArticle(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET like_received_count = like_received_count-1 where user_id = ?", userId)
	return result.RowsAffected
}

func CancelGivenLike(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET like_given_count = like_given_count-1 where user_id = ?", userId)
	return result.RowsAffected
}

// 增加关注
func Following(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET following_count = following_count+1 where user_id = ?", userId)
	return result.RowsAffected
}

// 取消关注
func CancelFollowing(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET following_count = following_count-1 where user_id = ?", userId)
	return result.RowsAffected
}

// 增加粉丝数
func Follower(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET follower_count = follower_count+1 where user_id = ?", userId)
	return result.RowsAffected
}

// 取消粉丝数量
func CancelFollower(userId uint64) int64 {
	result := builder().Exec("UPDATE user_statistics SET follower_count = follower_count-1 where user_id = ?", userId)
	return result.RowsAffected
}

// 更新活跃时间
func UpdateUserActivity(userId uint64, lastActiveTime time.Time) int64 {
	result := builder().Updates(Entity{UserId: userId, LastActiveTime: lastActiveTime})
	return result.RowsAffected
}

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

//func deleteEntity(entity *Entity) int64 {
//	result := builder().Delete(entity)
//	return result.RowsAffected
//}

//func all() (entities []*Entity) {
//	builder().Find(&entities)
//	return
//}
