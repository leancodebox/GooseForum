package dbconnect

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/connect/sqlconnect"
	"github.com/leancodebox/GooseForum/app/bundles/goose/fileopt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

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
		dbConfig := preferences.GetExclusivePreferences("db.default")
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

// Close 关闭数据库连接
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
		slog.Info("dbCloseSuccess")
	}
}

func BackupSQLiteHandle() {
	if !IsSqlite() {
		return
	}
	if !preferences.GetBool("db.backupSqlite", false) {
		return
	}
	backupDir := preferences.Get("db.backupDir")
	slog.Info("backupDir DirExistOrCreate", "err", fileopt.DirExistOrCreate(backupDir))
	err := backupSQLite(dbIns, backupDir)
	slog.Info("backupSQLite", "err", err)
	cleanOldBackups(backupDir, 7)
}

func backupSQLite(db *gorm.DB, backupDir string) error {
	// 生成带时间戳的备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("backup_%s.db", timestamp))

	// 使用 SQLite 备份 API
	result := db.Exec("VACUUM main INTO ?", backupPath)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func cleanOldBackups(backupDir string, maxBackups int) {
	files, _ := os.ReadDir(backupDir)
	var backupFiles []os.DirEntry

	// 筛选备份文件（按命名规则）
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "backup_") && strings.HasSuffix(file.Name(), ".db") {
			backupFiles = append(backupFiles, file)
		}
	}

	// 按修改时间排序（旧文件在前）
	sort.Slice(backupFiles, func(i, j int) bool {
		infoI, _ := backupFiles[i].Info()
		infoJ, _ := backupFiles[j].Info()
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// 删除超量的旧备份
	if len(backupFiles) > maxBackups {
		for i := 0; i < len(backupFiles)-maxBackups; i++ {
			os.Remove(filepath.Join(backupDir, backupFiles[i].Name()))
		}
	}
}
