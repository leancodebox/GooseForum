package docProjects

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

// GetProjectList 分页查询项目列表
func GetProjectList(page, pageSize int, keyword string, status *int8, isPublic *int8) ([]Entity, int64, error) {
	var entities []Entity
	var total int64

	query := builder()

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态筛选
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 公开性筛选
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&entities).Error

	return entities, total, err
}

// ExistsBySlug 检查slug是否存在
func ExistsBySlug(slug string) bool {
	var count int64
	builder().Where("slug = ? ", slug).Count(&count)
	return count > 0
}

// ExistsBySlugExcludeId 检查slug是否存在（排除指定ID）
func ExistsBySlugExcludeId(slug string, excludeId uint64) bool {
	var count int64
	builder().Where("slug = ? AND id != ? ", slug, excludeId).Count(&count)
	return count > 0
}

// SoftDelete 软删除
func SoftDelete(entity *Entity) int64 {
	result := builder().Delete(entity)
	return result.RowsAffected
}

// GetBySlug 根据slug获取项目
func GetBySlug(slug string) (entity Entity) {
	builder().Where("slug = ? ", slug).First(&entity)
	return
}

// GetAllActive 获取所有活跃项目
func GetAllActive() (entities []Entity) {
	builder().Where("status = ?", 2).Find(&entities)
	return
}

// GetByOwnerId 根据所有者ID获取项目列表
func GetByOwnerId(ownerId uint64) (entities []Entity) {
	builder().Where("owner_id = ? ", ownerId).Find(&entities)
	return
}
