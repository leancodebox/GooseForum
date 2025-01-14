package routes

import (
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/spf13/cast"
)

func fileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")

	// 文件上传接口
	r.POST("/img-upload", func(c *gin.Context) {
		// 判断上传令牌是否合法
		token := c.PostForm("token")
		if token != "your_token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 获取上传的文件
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
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

		// 生成文件名
		fileExt := path.Ext(file.Filename)
		fileNameInt := time.Now().Unix()
		fileNameStr := cast.ToString(fileNameInt)
		fileName := fileNameStr + fileExt
		folderName := time.Now().Format("2006/01/02")
		filePath := filepath.Join(folderName, fileName)

		// 保存到数据库
		entity, err := filedata.SaveFile(filePath, "image", fileData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
			//"id":      entity.Id,
			"name": entity.Name,
		})
	})

	// 文件获取接口
	r.GET("/img/:id", func(c *gin.Context) {
		id := cast.ToUint64(c.Param("id"))
		if id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
			return
		}

		entity, err := filedata.GetFile(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// 设置内容类型
		contentType := "image/jpeg"
		if strings.HasSuffix(entity.Name, ".png") {
			contentType = "image/png"
		}
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", "inline")

		c.Data(http.StatusOK, contentType, entity.Data)
	})

	r.GET("/image/*filepath", func(c *gin.Context) {
		id := c.Param("filepath")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filepath"})
			return
		}
		id = strings.TrimPrefix(id, "/")
		entity, err := filedata.GetFileByName(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// 设置内容类型
		contentType := "image/jpeg"
		if strings.HasSuffix(entity.Name, ".png") {
			contentType = "image/png"
		}
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", "inline")

		c.Data(http.StatusOK, contentType, entity.Data)
	})
}
