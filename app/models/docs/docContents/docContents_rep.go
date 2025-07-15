package docContents

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

//func all() (entities []*Entity) {
//	builder().Find(&entities)
//	return
//}

// GetContentList 获取内容列表
func GetContentList(page, pageSize int, versionId uint64, keyword string, status int) ([]*Entity, int64, error) {
	var contents []*Entity
	var total int64

	query := builder()

	// 版本ID过滤
	if versionId > 0 {
		query = query.Where("version_id = ?", versionId)
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("title LIKE ? OR slug LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态过滤
	if status >= 0 {
		query = query.Where("is_published = ?", status)
	}

	// 获取总数
	if err := query.Model(&Entity{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort_order ASC, created_at DESC").Find(&contents).Error; err != nil {
		return nil, 0, err
	}

	return contents, total, nil
}

// ExistsBySlugAndVersionId 检查slug在版本内是否已存在
func ExistsBySlugAndVersionId(slug string, versionId uint64) bool {
	var count int64
	builder().Model(&Entity{}).Where("slug = ? AND version_id = ?", slug, versionId).Count(&count)
	return count > 0
}

// ExistsBySlugAndVersionIdExcludeId 检查slug在版本内是否被其他内容使用
func ExistsBySlugAndVersionIdExcludeId(slug string, versionId uint64, excludeId uint64) bool {
	var count int64
	builder().Model(&Entity{}).Where("slug = ? AND version_id = ? AND id != ?", slug, versionId, excludeId).Count(&count)
	return count > 0
}

// GetByIdString 通过字符串ID获取内容
func GetByIdString(id string) (entity Entity) {
	builder().First(&entity, id)
	return
}

// Delete 删除内容
func Delete(id uint64) int64 {
	result := builder().Delete(&Entity{}, id)
	return result.RowsAffected
}

// UpdatePublishStatus 更新发布状态
func UpdatePublishStatus(id uint64, isPublished int8) int64 {
	result := builder().Model(&Entity{}).Where("id = ?", id).Update("is_published", isPublished)
	return result.RowsAffected
}

// GetByVersionId 获取版本下的所有内容
func GetByVersionId(versionId uint64) []*Entity {
	var contents []*Entity
	builder().Where("version_id = ?", versionId).Order("sort_order ASC, created_at DESC").Find(&contents)
	return contents
}

func GetBySlug(versionId uint64, slug string) (entity Entity) {
	builder().Where(queryopt.Eq(fieldVersionId, versionId)).Where(queryopt.Eq(fieldSlug, slug)).First(&entity)
	return
}
