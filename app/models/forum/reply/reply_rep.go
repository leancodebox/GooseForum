package reply

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
)

func Create(entity *Entity) error {
	result := builder().Create(entity)
	return result.Error
}

func Save(entity *Entity) error {
	result := builder().Save(entity)
	return result.Error
}

func SaveNoUpdate(entity *Entity) error {
	result := builder().Omit(fieldUpdatedAt).Save(entity)
	return result.Error
}

func Get(id any) (entity Entity) {
	builder().Where(pid, id).First(&entity)
	return
}

func GetByIds(ids []uint64) (entities []*Entity) {
	if len(ids) == 0 {
		return
	}
	builder().Where(queryopt.In(pid, ids)).Find(&entities)
	return
}

// GetAll 用于全量导出/修复数据，支持分页查询
func GetAll(offset, limit int) ([]*Entity, error) {
	var entities []*Entity
	err := builder().Offset(offset).Limit(limit).Order("id ASC").Find(&entities).Error
	return entities, err
}

func QueryById(startId uint64, limit int) (entities []*Entity) {
	builder().Where(queryopt.Gt(pid, startId)).Limit(limit).Order(queryopt.Asc(pid)).Find(&entities)
	return
}

// GetCountGroupByDay 按天统计回复数
func GetCountGroupByDay() ([]map[string]any, error) {
	var results []map[string]any
	err := builder().Select("DATE(created_at) as date, count(*) as count").Group("date").Order("date ASC").Find(&results).Error
	return results, err
}

func GetCount() int64 {
	var count int64
	builder().Count(&count)
	return count
}

func GetMaxId() uint64 {
	var entity Entity
	builder().Order(queryopt.Desc(pid)).Limit(1).First(&entity)
	return entity.Id
}

//func save(entity *Entity) int64 {
//	result := builder().Save(entity)
//	return result.RowsAffected
//}

//func saveAll(entities []*Entity) int64 {
//	result := builder().Save(entities)
//	return result.RowsAffected
//}

func DeleteEntity(entity *Entity) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

//func all() (entities []*Entity) {
//	builder().Find(&entities)
//	return
//}

func GetByMaxIdPage(articleId uint64, id uint64, pageSize int) (entities []Entity) {
	builder().Where(queryopt.Eq(fieldArticleId, articleId)).Where(queryopt.Gt(pid, id)).Limit(pageSize).Find(&entities)
	return
}

func GetByArticleId(articleId uint64) (entities []*Entity) {
	builder().Where(queryopt.Eq(fieldArticleId, articleId)).Limit(20).Order(queryopt.Asc(pid)).Find(&entities)
	return
}

func GetAllByArticleId(articleId uint64) (entities []*Entity) {
	builder().Where(queryopt.Eq(fieldArticleId, articleId)).Find(&entities)
	return
}

func GetByArticleIdAsc(articleId uint64, limit int) (entities []*Entity) {
	builder().Where(queryopt.Eq(fieldArticleId, articleId)).Limit(limit).Order(queryopt.Asc(pid)).Find(&entities)
	return
}

func GetByArticleIdAfter(articleId uint64, id uint64, limit int) (entities []*Entity) {
	builder().
		Where(queryopt.Eq(fieldArticleId, articleId)).
		Where(queryopt.Gt(pid, id)).
		Limit(limit).
		Order(queryopt.Asc(pid)).
		Find(&entities)
	return
}

func GetByArticleIdBefore(articleId uint64, id uint64, limit int) (entities []*Entity) {
	builder().
		Where(queryopt.Eq(fieldArticleId, articleId)).
		Where(queryopt.Lt(pid, id)).
		Limit(limit).
		Order(queryopt.Desc(pid)).
		Find(&entities)
	for i, j := 0, len(entities)-1; i < j; i, j = i+1, j-1 {
		entities[i], entities[j] = entities[j], entities[i]
	}
	return
}

func CountByArticleId(articleId uint64) int64 {
	var count int64
	builder().Where(queryopt.Eq(fieldArticleId, articleId)).Count(&count)
	return count
}

func GetUserCount(userId uint64) int64 {
	var count int64
	builder().Where(queryopt.Eq(fieldUserId, userId)).Where("deleted_at IS NULL").Count(&count)
	return count
}
