package categoryGroupPermissions

import "github.com/leancodebox/GooseForum/app/bundles/queryopt"

func EnabledByGroupID(groupID uint64) ([]Entity, error) {
	if groupID == 0 {
		return []Entity{}, nil
	}
	var entities []Entity
	err := builder().
		Where(queryopt.Eq("access_group_id", groupID)).
		Where(queryopt.Eq("status", StatusEnabled)).
		Order(queryopt.Asc("category_id")).
		Find(&entities).Error
	return entities, err
}

func ByCategoryIDs(categoryIDs []uint64) ([]Entity, error) {
	if len(categoryIDs) == 0 {
		return []Entity{}, nil
	}
	var entities []Entity
	err := builder().Where(queryopt.In("category_id", categoryIDs)).Order(queryopt.Asc("category_id")).Order(queryopt.Asc("access_group_id")).Find(&entities).Error
	return entities, err
}

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}
