package setupservice

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configPath = "storage/setup.json"
)

type DatabaseConfig struct {
	Type     string `json:"type"`     // mysql 或 sqlite
	Host     string `json:"host"`     // mysql 专用
	Port     string `json:"port"`     // mysql 专用
	DBName   string `json:"dbName"`   // 数据库名/sqlite文件路径
	Username string `json:"username"` // mysql 专用
	Password string `json:"password"` // mysql 专用
}

type SiteConfig struct {
	SiteName string         `json:"siteName"`
	SiteDesc string         `json:"siteDesc"`
	Database DatabaseConfig `json:"database"`
}

// IsInitialized 检查是否已初始化
func IsInitialized() bool {
	_, err := os.Stat(configPath)
	return !os.IsNotExist(err)
}

// SaveConfig 保存配置
func SaveConfig(config SiteConfig) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 保存配置文件
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// InitDatabase 初始化数据库
func InitDatabase(config DatabaseConfig) error {
	// 构建数据库连接字符串
	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	//	config.Username,
	//	config.Password,
	//	config.Host,
	//	config.Port,
	//	config.DBName,
	//)

	//// 测试连接
	//db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	//if err != nil {
	//	return fmt.Errorf("连接数据库失败: %w", err)
	//}

	// 执行数据库迁移
	// TODO: 实现数据库迁移逻辑

	return nil
}
