package dbconnect

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/sqlconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"sync"

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
		dbConfig := preferences.GetExclusivePreferences("db.default")
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
