package docVersions

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

// GetByIdString 通过字符串ID获取版本
func GetByIdString(id string) (entity Entity) {
	builder().First(&entity, id)
	return
}

// GetVersionList 获取版本列表
func GetVersionList(page, pageSize int, projectId uint64, keyword string, status int) (entities []Entity, total int64, err error) {
	query := builder()

	// 项目ID过滤
	if projectId > 0 {
		query = query.Where("project_id = ?", projectId)
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR slug LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态过滤
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Order("sort_order ASC, created_at DESC").Find(&entities)

	return entities, total, result.Error
}

// ExistsBySlugAndProjectId 检查项目下slug是否存在
func ExistsBySlugAndProjectId(slug string, projectId uint64) bool {
	var count int64
	builder().Where("slug = ? AND project_id = ?", slug, projectId).Count(&count)
	return count > 0
}

// ExistsBySlugAndProjectIdExcludeId 检查项目下slug是否存在（排除指定ID）
func ExistsBySlugAndProjectIdExcludeId(slug string, projectId uint64, excludeId uint64) bool {
	var count int64
	builder().Where("slug = ? AND project_id = ? AND id != ?", slug, projectId, excludeId).Count(&count)
	return count > 0
}

// ClearDefaultByProjectId 清除项目下的默认版本标记
func ClearDefaultByProjectId(projectId uint64) int64 {
	result := builder().Where("project_id = ?", projectId).Update("is_default", 0)
	return result.RowsAffected
}

// SoftDelete 软删除版本
func SoftDelete(entity *Entity) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

// HasDocuments 检查版本下是否有文档
func HasDocuments(versionId uint64) bool {
	// TODO: 这里需要根据实际的文档表结构来实现
	// 暂时返回false，等文档内容模块开发时再完善
	return false
}

// Create 公开的创建方法
func Create(entity *Entity) int64 {
	return create(entity)
}

// Save 公开的保存方法
func Save(entity *Entity) int64 {
	return save(entity)
}

func GetVersionByProject(projectId uint64) (entities []*Entity) {
	builder().Where(queryopt.Eq(fieldProjectId, projectId)).Find(&entities)
	return
}
