package articleUserAction

import (
	"gorm.io/gorm"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func builder() *gorm.DB {
	return db.Connect().Table(tableName)
}
