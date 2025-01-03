package dbconnect

import (
	"github.com/leancodebox/GooseForum/app/bundles/logging"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"log/slog"
	"sync"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/goose/preferences"

	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	// GORM 的 MYSQL 数据库驱动导入
	"gorm.io/driver/mysql"
)

//func init() {
//	bootstrap.AddDInit(connectDB)
//}

var (
	debug         = setting.IsDebug()
	dbConfig      = preferences.GetExclusivePreferences("db.default")
	isSqlite bool = false
)

var (
	connection         = dbConfig.Get(`connection`, `sqlite`)
	dbUrl              = dbConfig.Get(`url`)
	dbPath             = dbConfig.Get(`path`, `:memory:`)
	maxIdleConnections = dbConfig.GetInt(`maxIdleConnections`, 2)
	maxOpenConnections = dbConfig.GetInt(`maxOpenConnections`, 2)
	maxLifeSeconds     = dbConfig.GetInt(`maxLifeSeconds`, 60)
	once               = new(sync.Once)
)

// DB gorm.DB 对象
var dbIns *gorm.DB

func Connect() *gorm.DB {
	once.Do(func() {
		connectDB()
	})
	return dbIns
}
func IsSqlite() bool {
	connectDB()
	return isSqlite
}

// ConnectDB 初始化模型
func connectDB() {

	var err error
	switch connection {
	case "sqlite":
		slog.Info("use sqlite")
		dbIns, err = connectSqlLiteDB(logging.NewGormLogger())
		break
	case "mysql":
		slog.Info("use mysql")
		dbIns, err = connectMysqlDB(logging.NewGormLogger())
		break
	default:
		slog.Info("use sqlite because unselect db")
		dbIns, err = connectSqlLiteDB(logging.NewGormLogger())
		break
	}

	if err != nil {
		slog.Error("db connect", "err", err)
		panic(err)
	}

	if debug {
		slog.Info("开启debug")
		dbIns = dbIns.Debug()
	}

	// 获取底层的 sqlDB
	sqlDB, _ := dbIns.DB()
	// 设置最大连接数
	sqlDB.SetMaxOpenConns(maxOpenConnections)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(maxIdleConnections)
	// 设置每个链接的过期时间
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifeSeconds) * time.Second)
}

func connectMysqlDB(_logger gormlogger.Interface) (*gorm.DB, error) {
	// 初始化 MySQL 连接信息
	gormConfig := mysql.New(mysql.Config{
		DSN: dbUrl,
	})

	// 准备数据库连接池
	db, err := gorm.Open(gormConfig, &gorm.Config{
		Logger: _logger,
	})
	return db, err
}

func connectSqlLiteDB(_logger gormlogger.Interface) (*gorm.DB, error) {
	isSqlite = true
	if dbPath == ":memory:" {
		// ":memory:"
	} else if err := createFileIfNotExists(dbPath); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=busy_timeout(5000)"), &gorm.Config{Logger: _logger})
	return db, err
}

func createFileIfNotExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
			return err
		}
	}
	return nil
}
