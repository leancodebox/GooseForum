package accessGroups

import "github.com/leancodebox/GooseForum/app/bundles/queryopt"

func GetBySystemKeys(systemKeys []string) ([]Entity, error) {
	if len(systemKeys) == 0 {
		return []Entity{}, nil
	}
	var entities []Entity
	err := builder().
		Where(queryopt.In("system_key", systemKeys)).
		Where(queryopt.Eq("status", StatusEnabled)).
		Order(queryopt.Asc("id")).
		Find(&entities).Error
	return entities, err
}

func All() ([]Entity, error) {
	var entities []Entity
	err := builder().Order(queryopt.Asc("id")).Find(&entities).Error
	return entities, err
}

func Get(id uint64) (Entity, error) {
	var entity Entity
	if id == 0 {
		return entity, nil
	}
	err := builder().First(&entity, id).Error
	return entity, err
}

func Save(entity *Entity) error {
	return builder().Save(entity).Error
}

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

func Delete(entity *Entity) error {
	return builder().Delete(entity).Error
}
