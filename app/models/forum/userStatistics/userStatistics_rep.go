package userStatistics

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
	result := builder().Exec("UPDATE user_statistics SET comment_count = comment_count+1 where user_id = ?", userId)
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
