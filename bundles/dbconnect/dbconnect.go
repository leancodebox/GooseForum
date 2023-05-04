package dbconnect

import (
	"fmt"
	"github.com/leancodebox/GooseForum/bundles/logger"
	"github.com/leancodebox/goose/preferences"

	"github.com/glebarez/sqlite"

	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	// GORM 的 MYSQL 数据库驱动导入
	"gorm.io/driver/mysql"
)

var (
	connection         = preferences.Get(`db.connection`, `sqlite`)
	debug              = preferences.GetBool(`app.debug`, `sqlite`)
	dbUrl              = preferences.Get(`db.url`)
	dbPath             = preferences.Get(`db.path`, `:memory:`)
	maxIdleConnections = preferences.GetInt(`db.maxIdleConnections`, 2)
	maxOpenConnections = preferences.GetInt(`db.maxOpenConnections`, 2)
	maxLifeSeconds     = preferences.GetInt(`db.maxLifeSeconds`, 60)
)

func init() {
	connectDB()
}

// DB gorm.DB 对象
var dbIns *gorm.DB

func Std() *gorm.DB {
	return dbIns
}

// NewMysql dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
// NewMysql
func NewMysql(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db, err
}

// ConnectDB 初始化模型
func connectDB() {

	var err error
	switch connection {
	case "sqlite":
		logger.Info("use sqlite")
		dbIns, err = connectSqlLiteDB(logger.NewGormLogger())
		break
	case "mysql":
		logger.Info("use mysql")
		dbIns, err = connectMysqlDB(logger.NewGormLogger())
		break
	default:
		logger.Info("use sqlite because unselect db")
		dbIns, err = connectSqlLiteDB(logger.NewGormLogger())
		break
	}

	if err != nil {
		log.Println(err)
		panic(err)
	}

	if debug {
		fmt.Println("开启debug")
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
