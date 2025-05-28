package sqlconnect

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/bundles/logging"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	Init    bool
}

func (itself *Connect) IsSqlite() bool {
	return itself.Config.Connection == "sqlite"
}

func GetConnectByPreferences(preferences preferences.ExclusivePreferences) Connect {
	c := Config{
		Connection:         preferences.Get(`connection`, `sqlite`),
		DbUrl:              preferences.Get(`url`),
		DbPath:             preferences.Get(`path`, `:memory:`),
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
		return Connect{Config: config, Connect: dbIns, Error: err}
	}

	if debug {
		slog.Info("开启debug")
		dbIns = dbIns.Debug()
	}

	// 获取底层的 sqlDB
	sqlDB, err := dbIns.DB()
	if err != nil {
		slog.Error(err.Error())
		return Connect{Config: config, Connect: dbIns, Error: err}
	}
	// 设置最大连接数
	sqlDB.SetMaxOpenConns(config.MaxOpenConnections)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(config.MaxIdleConnections)
	// 设置每个链接的过期时间
	sqlDB.SetConnMaxLifetime(time.Duration(config.MaxLifeSeconds) * time.Second)
	return Connect{Config: config, Connect: dbIns, Error: err}
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

	// 构建 SQLite DSN，启用 WAL 模式和其他优化配置
	dsn := buildSQLiteDSN(dbPath, map[string]string{
		"journal_mode":       "WAL",
		"cache_size":         "-20000",
		"synchronous":        "NORMAL",
		"journal_size_limit": "1048576",
		"wal_autocheckpoint": "1000",
		"page_size":          "8192",
		"busy_timeout":       "5000",
	})

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
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

func buildSQLiteDSN(filepath string, config map[string]string) string {
	// 如果 filepath 已经包含参数，则拆分路径和参数
	basePath := filepath
	existingParams := make(map[string]string)
	if idx := strings.Index(filepath, "?"); idx != -1 {
		basePath = filepath[:idx]
		paramStr := filepath[idx+1:]

		// 解析已有的参数
		for _, param := range strings.Split(paramStr, "&") {
			if strings.HasPrefix(param, "_pragma=") {
				// 提取 _pragma 的值
				pragmaValue := strings.TrimPrefix(param, "_pragma=")
				// 拆分 key 和 value
				if pidx := strings.Index(pragmaValue, "("); pidx != -1 {
					key := pragmaValue[:pidx]
					value := pragmaValue[pidx+1 : len(pragmaValue)-1] // 去掉括号
					existingParams[key] = value
				}
			}
		}
	}

	// 构建新的参数
	var params []string
	for key, value := range config {
		// 如果参数已经存在，跳过
		if _, exists := existingParams[key]; exists {
			continue
		}
		// 添加新的参数
		params = append(params, fmt.Sprintf("_pragma=%s(%s)", key, value))
	}

	// 如果有已有的参数，添加到 params 中
	if len(existingParams) > 0 {
		for key, value := range existingParams {
			params = append(params, fmt.Sprintf("_pragma=%s(%s)", key, value))
		}
	}

	// 拼接路径和参数
	if len(params) > 0 {
		return basePath + "?" + strings.Join(params, "&")
	}
	return basePath
}
