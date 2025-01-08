package db4cacheconnect

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/sqlconnect"
	"sync"

	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"

	"gorm.io/gorm"
)

//func init() {
//	bootstrap.AddDInit(connectDB)
//}

var (
	dbConfig      = preferences.GetExclusivePreferences("db.cache")
	isSqlite bool = false
	once          = new(sync.Once)
)

// DB gorm.DB 对象
var dbIns *gorm.DB

func Connect() *gorm.DB {
	once.Do(func() {
		res := sqlconnect.GetConnectByPreferences(dbConfig)
		dbIns = res.Connect
		isSqlite = res.IsSqlite()
	})
	return dbIns
}
func IsSqlite() bool {
	Connect()
	return isSqlite
}
