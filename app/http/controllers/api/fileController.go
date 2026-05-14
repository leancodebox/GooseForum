package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
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
		c.Redirect(http.StatusFound, urlconfig.GetDefaultAvatar())
		return
	}
	c.Header("Content-Disposition", "inline")
	c.Data(http.StatusOK, entity.Type, entity.Data)
}

// SaveImgByGinContext handles image uploads with size and content checks.
func SaveImgByGinContext(c *gin.Context) {
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()

	if !postingConfig.UploadControl.AllowAttachments {
		c.JSON(http.StatusForbidden, component.FailData("目前已关闭附件上传功能"))
		return
	}

	userId := c.GetUint64(`userId`)
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailData("用户未登录"))
		return
	}

	userEntity, _ := users.Get(userId)

	if err, code := component.CheckUserPermission(&userEntity, "上传附件"); err != nil {
		c.JSON(code, component.FailData(err.Error()))
		return
	}

	if postingConfig.UploadControl.NewUserUploadCooldownMinutes > 0 {
		cooldownTime := userEntity.CreatedAt.Add(time.Duration(postingConfig.UploadControl.NewUserUploadCooldownMinutes) * time.Minute)
		if time.Now().Before(cooldownTime) {
			c.JSON(http.StatusBadRequest, component.FailData(fmt.Sprintf("新用户注册%d分钟后才能上传，请在 %s 后再试",
				postingConfig.UploadControl.NewUserUploadCooldownMinutes,
				cooldownTime.Format("2006-01-02 15:04:05"))))
			return
		}
	}

	if postingConfig.UploadControl.MaxDailyUploadsPerUser > 0 {
		count := filedata.CountDailyUploads(userId)
		if count >= int64(postingConfig.UploadControl.MaxDailyUploadsPerUser) {
			c.JSON(http.StatusBadRequest, component.FailData(fmt.Sprintf("您今日已上传 %d 个文件，已达到每日限制", count)))
			return
		}
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, component.FailData("文件上传失败，请检查文件是否正确选择"))
		return
	}

	if file.Filename == "" {
		c.JSON(http.StatusBadRequest, component.FailData("文件名不能为空"))
		return
	}

	configMaxSize := int64(postingConfig.UploadControl.MaxAttachmentSizeKb) * 1024
	maxSize := int64(filedata.MaxFileSize)
	if configMaxSize > 0 && configMaxSize < maxSize {
		maxSize = configMaxSize
	}

	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, component.FailData(fmt.Sprintf("文件大小超过限制，最大允许%dKB", maxSize/1024)))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := postingConfig.UploadControl.AuthorizedExtensions
	if len(allowedExts) > 0 {
		if !isAllowedExtension(ext, allowedExts) {
			c.JSON(http.StatusBadRequest, component.FailData(fmt.Sprintf("不支持的文件格式，允许的格式为: %s", strings.Join(allowedExts, ", "))))
			return
		}
	} else {
		if _, err = filedata.CheckImageType(file.Filename); err != nil {
			c.JSON(http.StatusBadRequest, component.FailData("不支持的图片格式，仅支持 JPG、PNG、GIF、WebP、BMP 格式"))
			return
		}
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailData("文件读取失败，请重试"))
		return
	}
	defer src.Close()

	header := make([]byte, 512)
	n, _ := io.ReadFull(src, header)
	if n > 0 {
		if !isValidImageContent(header[:n]) {
			c.JSON(http.StatusBadRequest, component.FailData("文件内容不是有效的图片格式"))
			return
		}
	}

	remainingData, err := io.ReadAll(io.LimitReader(src, maxSize-int64(n)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailData("文件内容读取失败"))
		return
	}

	fileData := append(header[:n], remainingData...)

	if int64(len(fileData)) > maxSize {
		c.JSON(http.StatusBadRequest, component.FailData(fmt.Sprintf("文件大小超过限制，最大允许%dKB", maxSize/1024)))
		return
	}

	folderName := time.Now().Format("2006/01/02")

	entity, err := filedata.SaveFileFromUpload(userId, fileData, file.Filename, folderName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailData(fmt.Sprintf("文件保存失败: %s", err.Error())))
		return
	}

	c.JSON(http.StatusOK, component.SuccessData(map[string]any{
		"url":      entity.GetAccessPath(),
		"filename": file.Filename,
		"size":     len(fileData),
		"message":  "图片上传成功",
	}))
}

// isValidImageContent checks common image file signatures.
func isValidImageContent(data []byte) bool {
	if len(data) < 8 {
		return false
	}

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

func isAllowedExtension(ext string, allowedExts []string) bool {
	for _, allowedExt := range allowedExts {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}
	return false
}
