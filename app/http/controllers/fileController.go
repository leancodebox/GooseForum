package controllers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/resource"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
)

func GetFileByFileName(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}
	filename = strings.TrimPrefix(filename, "/")

	entity, err := filedata.GetFileByName(filename)
	if err != nil {
		c.Data(http.StatusOK, "image/png", resource.GetDefaultAvatar())
		//c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Type", entity.Type)
	c.Header("Content-Disposition", "inline")
	c.Data(http.StatusOK, entity.Type, entity.Data)
}

// SaveImgByGinContext 处理图片上传请求
// 包含文件大小限制、格式验证、内容检查等安全措施
func SaveImgByGinContext(c *gin.Context) {
	// 获取用户ID
	userId := c.GetUint64(`userId`)
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailData("用户未登录"))
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, component.FailData("文件上传失败，请检查文件是否正确选择"))
		return
	}

	// 验证文件名
	if file.Filename == "" {
		c.JSON(http.StatusBadRequest, component.FailData("文件名不能为空"))
		return
	}

	// 预检查文件大小（基于Header中的大小）
	if file.Size > filedata.MaxFileSize {
		c.JSON(http.StatusBadRequest, component.FailData(fmt.Sprintf("文件大小超过限制，最大允许%dMB", filedata.MaxFileSize/(1024*1024))))
		return
	}

	// 预检查文件扩展名
	_, err = filedata.CheckImageType(file.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, component.FailData("不支持的图片格式，仅支持 JPG、PNG、GIF、WebP、BMP 格式"))
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailData("文件读取失败，请重试"))
		return
	}
	defer src.Close()

	// 读取文件内容
	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailData("文件内容读取失败"))
		return
	}

	// 二次验证文件大小（基于实际读取的数据）
	if len(fileData) > filedata.MaxFileSize {
		c.JSON(http.StatusBadRequest, component.FailData(fmt.Sprintf("文件大小超过限制，最大允许%dMB", filedata.MaxFileSize/(1024*1024))))
		return
	}

	// 验证文件内容是否为有效图片（检查文件头）
	if !isValidImageContent(fileData) {
		c.JSON(http.StatusBadRequest, component.FailData("文件内容不是有效的图片格式"))
		return
	}

	// 生成存储路径
	folderName := time.Now().Format("2006/01/02")

	// 保存文件
	entity, err := filedata.SaveFileFromUpload(userId, fileData, file.Filename, folderName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailData(fmt.Sprintf("文件保存失败: %s", err.Error())))
		return
	}

	c.JSON(http.StatusOK, component.SuccessData(map[string]interface{}{
		"url":      entity.GetAccessPath(),
		"filename": file.Filename,
		"size":     len(fileData),
		"message":  "图片上传成功",
	}))
}

// isValidImageContent 验证文件内容是否为有效的图片格式
// 通过检查文件头魔数来判断文件类型
func isValidImageContent(data []byte) bool {
	if len(data) < 8 {
		return false
	}

	// 检查常见图片格式的文件头
	var imageSignatures = [][]byte{
		{0xFF, 0xD8, 0xFF}, // JPEG
		{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG
		{0x47, 0x49, 0x46, 0x38, 0x37, 0x61},             // GIF87a
		{0x47, 0x49, 0x46, 0x38, 0x39, 0x61},             // GIF89a
		{0x52, 0x49, 0x46, 0x46},                         // WebP (RIFF)
		{0x42, 0x4D},                                     // BMP
	}

	for _, signature := range imageSignatures {
		if len(data) >= len(signature) && bytes.HasPrefix(data, signature) {
			// 对于WebP，需要进一步验证
			if bytes.HasPrefix(signature, []byte{0x52, 0x49, 0x46, 0x46}) {
				if len(data) >= 12 && bytes.Equal(data[8:12], []byte("WEBP")) {
					return true
				}
				continue
			}
			return true
		}
	}

	return false
}
