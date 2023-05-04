package Comment

import (
	"gorm.io/gorm"

	db "github.com/leancodebox/GooseForum/bundles/dbconnect"
)

// Prohibit manual changes
// 禁止手动更改本文件

func builder() *gorm.DB {
	return db.Std().Table(tableName)
}

func first(db *gorm.DB) (el Comment) {
	db.First(&el)
	return
}

func List(db *gorm.DB) (el []*Comment) {
	db.Find(&el)
	return
}
