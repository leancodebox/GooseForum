package component

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"path/filepath"
)

// HandleFileUpload 处理文件上传的通用逻辑
func HandleFileUpload(c *gin.Context, fieldName string) ([]byte, string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return nil, "", fmt.Errorf("获取上传文件失败: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return nil, "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		return nil, "", fmt.Errorf("读取文件失败: %w", err)
	}

	return fileData, file.Filename, nil
}

func FilePath(filename string) string {
	return filepath.Join("/file/img", filename)
}
