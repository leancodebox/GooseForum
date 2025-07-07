package kvstore

import (
	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

// Prohibit manual changes
// 禁止手动更改本文件

func builder() *gorm.DB {
	return db.Connect().Table(tableName)
}

func first(db *gorm.DB) (el Entity) {
	db.First(&el)
	return
}

func List(db *gorm.DB) (el []*Entity) {
	db.Find(&el)
	return
}
