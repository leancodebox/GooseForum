package db4fileconnect

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/sqlconnect"
	"log/slog"
	"sync"

	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"

	"gorm.io/gorm"
)

//func init() {
//	bootstrap.AddDInit(connectDB)
//}

var (
	isSqlite bool = false
	once          = new(sync.Once)
)

// DB gorm.DB 对象
var dbIns *gorm.DB

func Connect() *gorm.DB {
	once.Do(func() {
		dbConfig := preferences.GetExclusivePreferences("db.file")
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


func Close() {
	if dbIns == nil {
		return
	}
	db, err := dbIns.DB()
	if err != nil {
		return
	}
	if db == nil {
		return
	}
	if err = db.Close(); err != nil {
		slog.Error("dbClose", "err", err)
	} else {
		slog.Error("dbCloseSuccess")
	}
}
