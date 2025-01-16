package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed css/*
var staticFiles embed.FS

// GetFileSystem 返回一个http.FileSystem，用于提供静态文件服务
func GetFileSystem() http.FileSystem {
	fsys, err := fs.Sub(staticFiles, ".")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}


func GetStaticFile() embed.FS {
	return staticFiles
}
