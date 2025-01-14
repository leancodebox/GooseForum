package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
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
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Type", entity.Type)
	c.Header("Content-Disposition", "inline")
	c.Data(http.StatusOK, entity.Type, entity.Data)
}

func SaveFileByGinContext(c *gin.Context) {

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}

	// 检查文件类型
	contentType, err := filedata.CheckImageType(file.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if file.Size > 2*1024*1024 { // 限制文件大小为2MB
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File size too large"})
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File open failed"})
		return
	}
	defer src.Close()

	// 直接读取文件内容
	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// 生成文件名和路径
	fileExt := path.Ext(file.Filename)
	fileId := uuid.New().String()
	fileName := fileId + fileExt
	folderName := time.Now().Format("2006/01/02")
	filePath := filepath.Join(folderName, fileName)

	// 保存到数据库
	entity, err := filedata.SaveFile(filePath, contentType, fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"name":    entity.Name,
	})
}
