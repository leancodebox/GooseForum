package db4fileconnect

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
	once = new(sync.Once)
)

var dbConnect sqlconnect.Connect

func Connect() *gorm.DB {
	once.Do(func() {
		dbConfig := preferences.GetExclusivePreferences("db.file")
		dbConnect = sqlconnect.GetConnectByPreferences(dbConfig)
	})
	return dbConnect.Connect
}

func IsSqlite() bool {
	Connect()
	return dbConnect.IsSqlite()
}

// Close 关闭数据库连接
func Close() {
	dbConnect.Close()
}

func BackupSQLiteHandle() {
	dbConnect.BackupSQLiteHandle()
}
