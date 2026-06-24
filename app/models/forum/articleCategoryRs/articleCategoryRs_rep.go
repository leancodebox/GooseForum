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
	}

	return save(entity)
}

// GetByArticleId 获取文章的所有分类关系
func GetByArticleId(articleId uint64) (entities []*Entity) {
	builder().Where("article_id = ?", articleId).Find(&entities)
	return
}

// DeleteByArticleId 删除文章的所有分类关系
func DeleteByArticleId(articleId uint64) int64 {
	result := builder().Where("article_id = ?", articleId).Delete(&Entity{})
	return result.RowsAffected
}

func GetOneByCategoryId(cId uint64) (entity Entity) {
	builder().
		Where(queryopt.Eq(fieldArticleCategoryId, cId)).
		Where(queryopt.Eq(fieldEffective, 1)).First(&entity)
	return
}
