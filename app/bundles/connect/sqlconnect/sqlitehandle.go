package sqlconnect

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/fileopt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"gorm.io/gorm"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func getBackUpDir() string {
	return preferences.Get("db.backupDir")
}

func (itself *Connect) GenerateBackupPath(backupDir string) string {
	if !itself.IsSqlite() || itself.Connect == nil {
		return ""
	}
	sourcePath := itself.Config.DbPath
	// 处理内存数据库的特殊标识
	if sourcePath == ":memory:" {
		return filepath.Join(backupDir, fmt.Sprintf("memory_%s.db", time.Now().Format("20060102_150405")))
	}

	// 提取源文件名（不含扩展名）
	baseName := filepath.Base(sourcePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// 生成带时间戳的备份文件名
	timestamp := time.Now().Format("20060102_150405")
	return filepath.Join(backupDir, fmt.Sprintf("%s_%s.db", timestamp, nameWithoutExt))
}

func (itself *Connect) BackupSQLiteHandle() {
	if !itself.IsSqlite() || itself.Connect == nil {
		return
	}
	if !preferences.GetBool("db.backupSqlite", false) {
		return
	}
	backupDir := getBackUpDir()
	backupPath := itself.GenerateBackupPath(backupDir)
	slog.Info("backupDir DirExistOrCreate", "err", fileopt.DirExistOrCreate(backupDir))
	err := backupSQLite(itself.Connect, backupPath)
	slog.Info("backupSQLite", "err", err)
	keep := max(preferences.GetInt("db.keep", 7), 1)
	cleanOldBackups(itself.Config.DbPath, keep)
}

func backupSQLite(db *gorm.DB, backupPath string) error {

	// 使用 SQLite 备份 API
	result := db.Exec("VACUUM main INTO ?", backupPath)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func cleanOldBackups(sourcePath string, keep int) {
	// 获取同源的所有备份文件
	// 提取源文件名（不含扩展名）
	baseName := filepath.Base(sourcePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	searchFileName := filepath.Join(getBackUpDir(), fmt.Sprintf("*%s*.db", nameWithoutExt))
	files, err := filepath.Glob(searchFileName)
	if err != nil {
		slog.Error("cleanOldBackups err", "err", err)
		return
	}

	// 按修改时间排序（旧文件在前）
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// 删除超量旧备份
	if len(files) > keep {
		for _, f := range files[:len(files)-keep] {
			if err = os.Remove(f); err != nil {
				slog.Error("cleanOldBackups 删除失败", "file", f, "err", err)
			}
		}
	}
}
