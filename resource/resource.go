package resource

import (
	"embed"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"io/fs"
	"os"
	"path/filepath"
)

// isDevelopment 判断是否为开发模式
func isDevelopment() bool {
	return !setting.IsProduction()
}

//go:embed static/*
var staticFS embed.FS

// GetStaticFS 返回静态文件的文件系统
func GetStaticFS() (fs.FS, error) {
	if isDevelopment() {
		return os.DirFS(filepath.Join("resource", "static")), nil
	}
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	return static, nil
}
