package api

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/httputil"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func GetFileByFileName(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "Invalid filename",
			"messageCode": component.MessageRequestInvalidParams,
		})
		return
	}
	filename = strings.TrimPrefix(filename, "/")

	entity, err := filedata.GetFileByName(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "File not found",
			"messageCode": component.MessagePageNotFound,
		})
		return
	}
	c.Header("Content-Disposition", "inline")
	httputil.SetLongPublic(c)
	c.Data(http.StatusOK, entity.Type, entity.Data)
}

// SaveImgByGinContext handles image uploads with size and content checks.
func SaveImgByGinContext(c *gin.Context) {
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()

	userId := c.GetUint64(`userId`)
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, component.FailDataCode(component.MessageAuthRequired, nil))
		return
	}

	userEntity, _ := users.Get(userId)
	isRoleUser := userEntity.RoleId > 0

	if !isRoleUser && !postingConfig.UploadControl.AllowAttachments {
		c.JSON(http.StatusForbidden, component.FailDataCode(component.MessageUploadAttachmentDisabled, nil))
		return
	}

	if code, err := component.CheckUserPermission(&userEntity, component.PermissionActionUploadAttachment); err != nil {
		c.JSON(code, component.FailDataError(err))
		return
	}

	if !isRoleUser && postingConfig.UploadControl.NewUserUploadCooldownMinutes > 0 {
		cooldownTime := userEntity.CreatedAt.Add(time.Duration(postingConfig.UploadControl.NewUserUploadCooldownMinutes) * time.Minute)
		if time.Now().Before(cooldownTime) {
			minutes := postingConfig.UploadControl.NewUserUploadCooldownMinutes
			availableAt := cooldownTime.Format("2006-01-02 15:04:05")
			c.JSON(http.StatusBadRequest, component.FailDataCode(
				component.MessageUploadCooldown,

				component.MessageParams{"minutes": minutes, "availableAt": availableAt}))
			return
		}
	}

	if !isRoleUser && postingConfig.UploadControl.MaxDailyUploadsPerUser > 0 {
		count := filedata.CountDailyUploads(userId)
		if count >= int64(postingConfig.UploadControl.MaxDailyUploadsPerUser) {
			c.JSON(http.StatusBadRequest, component.FailDataCode(
				component.MessageUploadDailyLimit,

				component.MessageParams{"count": count}))
			return
		}
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageUploadFileMissing, nil))
		return
	}

	if file.Filename == "" {
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageUploadFilenameRequired, nil))
		return
	}

	configMaxSize := int64(postingConfig.UploadControl.MaxAttachmentSizeKb) * 1024
	maxSize := int64(filedata.MaxFileSize)
	if !isRoleUser && configMaxSize > 0 && configMaxSize < maxSize {
		maxSize = configMaxSize
	}

	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, component.FailDataCode(
			component.MessageUploadFileTooLarge,

			component.MessageParams{"maxSizeKb": maxSize / 1024}))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := postingConfig.UploadControl.AuthorizedExtensions
	if len(allowedExts) > 0 {
		if !isAllowedExtension(ext, allowedExts) {
			extensions := strings.Join(allowedExts, ", ")
			c.JSON(http.StatusBadRequest, component.FailDataCode(
				component.MessageUploadUnsupportedExt,

				component.MessageParams{"extensions": extensions}))
			return
		}
	} else {
		if _, err = filedata.CheckImageType(file.Filename); err != nil {
			c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageUploadUnsupportedImage, nil))
			return
		}
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageUploadReadFailed, nil))
		return
	}
	defer func() { _ = src.Close() }()

	header := make([]byte, 512)
	n, _ := io.ReadFull(src, header)
	if n > 0 {
		if !isValidImageContent(header[:n]) {
			c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageUploadInvalidImage, nil))
			return
		}
	}

	remainingData, err := io.ReadAll(io.LimitReader(src, maxSize-int64(n)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageUploadContentReadFailed, nil))
		return
	}

	fileData := make([]byte, n+len(remainingData))
	copy(fileData, header[:n])
	copy(fileData[n:], remainingData)

	if int64(len(fileData)) > maxSize {
		c.JSON(http.StatusBadRequest, component.FailDataCode(
			component.MessageUploadFileTooLarge,

			component.MessageParams{"maxSizeKb": maxSize / 1024}))
		return
	}

	folderName := time.Now().Format("2006/01/02")

	entity, err := filedata.SaveFileFromUpload(userId, fileData, file.Filename, folderName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailDataCode(
			component.MessageUploadSaveFailed,

			component.MessageParams{"error": err.Error()}))
		return
	}

	c.JSON(http.StatusOK, component.SuccessDataCode(map[string]any{
		"url":      entity.GetAccessPath(),
		"filename": file.Filename,
		"size":     len(fileData),
	}, component.MessageUploadSuccess, nil))
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
