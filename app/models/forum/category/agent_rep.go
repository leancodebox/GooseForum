package category

type AgentCategory struct {
	Id               uint64
	Name, Desc, Slug string
}

func ListAgentCategories(categoryIDs []uint64, globalManage bool) ([]AgentCategory, error) {
	q := builder().Model(&Entity{}).Order("sort ASC, id ASC")
	if !globalManage {
		q = q.Where("id IN ?", categoryIDs)
	}
	var rows []AgentCategory
	err := q.Find(&rows).Error
	return rows, err
}
