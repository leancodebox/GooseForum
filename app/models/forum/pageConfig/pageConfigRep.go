package pageConfig

import (
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/spf13/cast"
)

func create(entity *Entity) int64 {
	result := builder().Create(entity)
	return result.RowsAffected
}

func save(entity *Entity) int64 {
	result := builder().Save(entity)
	return result.RowsAffected
}

func CreateOrSave(entity *Entity) int64 {
	if entity.Id == 0 {
		return create(entity)
	}

	return save(entity)
}

func GetByPageType(pageType string) (entity Entity) {
	builder().Where(queryopt.Eq(filedPageType, pageType)).First(&entity)
	return
}

func GetConfigByPageType[T any](pageType string, defaultValue T) T {
	var entity Entity
	builder().Where(queryopt.Eq(filedPageType, pageType)).First(&entity)
	if entity.Id > 0 {
		return jsonopt.Decode[T](entity.Config)
	}

	return defaultValue
}

const AppMigrationVersion uint32 = 3

func GetMigrationVersion() uint32 {
	configEntity := GetByPageType(Migration)
	return cast.ToUint32(configEntity.Config)
}

func SyncMigrationVersion(version uint32) {
	configEntity := GetByPageType(Migration)
	configEntity.PageType = Migration
	configEntity.Config = cast.ToString(version)
	CreateOrSave(&configEntity)
}
