package sqlconnect

import (
	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/logging"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var debug = setting.IsDebug()

type Config struct {
	Connection         string
	DbUrl              string
	DbPath             string
	MaxIdleConnections int
	MaxOpenConnections int
	MaxLifeSeconds     int
}

type Connect struct {
	Config  Config
	Connect *gorm.DB
	Error   error
}

func (itself *Connect) IsSqlite() bool {
	return itself.Config.Connection == "sqlite"
}

func GetConnectByPreferences(preferences preferences.ExclusivePreferences) Connect {
	c := Config{
		Connection:         preferences.Get(`connection`, `sqlite`),
		DbUrl:              preferences.Get(`url`),
		DbPath:             preferences.Get(`path`, `storage/database/sqlite4file.db`),
		MaxIdleConnections: preferences.GetInt(`maxIdleConnections`, 2),
		MaxOpenConnections: preferences.GetInt(`maxOpenConnections`, 2),
		MaxLifeSeconds:     preferences.GetInt(`maxLifeSeconds`, 60),
	}
	return GetConnect(c)
}

// GetConnect 初始化模型
func GetConnect(config Config) Connect {
	var dbIns *gorm.DB
	var err error
	switch config.Connection {
	case "sqlite":
		slog.Info("use sqlite")
		dbIns, err = connectSqlLiteDB(config.DbPath)
		break
	case "mysql":
		slog.Info("use mysql")
		dbIns, err = connectMysqlDB(config.DbUrl)
		break
	default:
		slog.Info("use sqlite because unselect db")
		dbIns, err = connectSqlLiteDB(config.DbPath)
		break
	}

	if err != nil {
		slog.Error(err.Error())
		return Connect{config, dbIns, err}
	}

	if debug {
		slog.Info("开启debug")
		dbIns = dbIns.Debug()
	}

	// 获取底层的 sqlDB
	sqlDB, err := dbIns.DB()
	if err != nil {
		slog.Error(err.Error())
		return Connect{config, dbIns, nil}
	}
	// 设置最大连接数
	sqlDB.SetMaxOpenConns(config.MaxOpenConnections)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(config.MaxIdleConnections)
	// 设置每个链接的过期时间
	sqlDB.SetConnMaxLifetime(time.Duration(config.MaxLifeSeconds) * time.Second)
	return Connect{config, dbIns, nil}
}

func connectMysqlDB(dbUrl string) (*gorm.DB, error) {
	// 初始化 MySQL 连接信息
	gormConfig := mysql.New(mysql.Config{
		DSN: dbUrl,
	})

	// 准备数据库连接池
	db, err := gorm.Open(gormConfig, &gorm.Config{
		Logger: logging.NewGormLogger(),
	})
	return db, err
}

func connectSqlLiteDB(dbPath string) (*gorm.DB, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	} else if dbPath == ":memory:" {
		// ":memory:"
	} else if err := createFileIfNotExists(dbPath); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=busy_timeout(5000)"), &gorm.Config{
		Logger: logging.NewGormLogger(),
	})
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
