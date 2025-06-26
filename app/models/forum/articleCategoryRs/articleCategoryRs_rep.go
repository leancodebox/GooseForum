package articleCategoryRs

import (
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
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

// GetByArticleId 获取文章的所有分类关系
func GetByArticleId(articleId uint64) (entities []*Entity) {
	builder().Where("article_id = ?", articleId).Find(&entities)
	return
}

// GetByArticleIds 批量获取文章的分类关系
func GetByArticleIds(articleIds []uint64) (entities []*Entity) {
	if len(articleIds) == 0 {
		return
	}
	builder().Where("article_id IN ?", articleIds).Find(&entities)
	return
}

// GetByArticleIdsEffective 批量获取文章的分类关系
func GetByArticleIdsEffective(articleIds []uint64) (entities []*Entity) {
	if len(articleIds) == 0 {
		return
	}
	builder().Where("article_id IN (?)", articleIds).Where(queryopt.Eq(fieldEffective, 1)).Find(&entities)
	return
}

// DeleteByArticleId 删除文章的所有分类关系
func DeleteByArticleId(articleId uint64) int64 {
	result := builder().Where("article_id = ?", articleId).Delete(&Entity{})
	return result.RowsAffected
}

// BatchCreate 批量创建分类关系
func BatchCreate(entities []*Entity) int64 {
	if len(entities) == 0 {
		return 0
	}
	result := builder().Create(&entities)
	return result.RowsAffected
}

func GetOneByCategoryId(cId uint64) (entity Entity) {
	builder().
		Where(queryopt.Eq(fieldArticleCategoryId, cId)).
		Where(queryopt.Eq(fieldEffective, 1)).First(&entity)
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

//func all() (entities []*Entity) {
//	builder().Find(&entities)
//	return
//}
