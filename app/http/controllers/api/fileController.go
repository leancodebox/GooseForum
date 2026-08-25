package api

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/httputil"
	"github.com/leancodebox/GooseForum/app/service/fileaccessservice"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
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

	metadata, err := filestorage.Stat(c.Request.Context(), filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "File not found",
			"messageCode": component.MessagePageNotFound,
		})
		return
	}
	decision, err := fileaccessservice.Resolve(component.LoginUserId(c), filename, metadata.UserId)
	if err != nil || !decision.Allowed {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "File not found",
			"messageCode": component.MessagePageNotFound,
		})
		return
	}
	object, err := filestorage.Open(c.Request.Context(), filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found", "messageCode": component.MessagePageNotFound})
		return
	}
	defer func() { _ = object.Body.Close() }()
	c.Header("Content-Disposition", "inline")
	if decision.Public {
		httputil.SetLongPublic(c)
	} else {
		c.Header("Cache-Control", "private, max-age=300")
		c.Header("Vary", "Cookie, Authorization")
	}
	c.DataFromReader(http.StatusOK, object.Metadata.Size, object.Metadata.ContentType, object.Body, nil)
}

// SaveImgByGinContext handles image uploads with size and content checks.
func SaveImgByGinContext(c *gin.Context) {
	saveImgByGinContext(c, false)
}

func SaveAdminImgByGinContext(c *gin.Context) {
	saveImgByGinContext(c, true)
}

func saveImgByGinContext(c *gin.Context, adminUpload bool) {
	userId := c.GetUint64(`userId`)
	policy, failure := resolveImageUploadPolicy(userId)
	if failure != nil {
		c.JSON(failure.Status, failure.Data)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageUploadFileMissing, nil))
		return
	}

	contentType, failure := policy.Validate(file.Filename, file.Size, "")
	if failure != nil {
		c.JSON(failure.Status, failure.Data)
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageUploadReadFailed, nil))
		return
	}
	defer func() { _ = src.Close() }()

	fileData, err := io.ReadAll(io.LimitReader(src, policy.MaxSize+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageUploadContentReadFailed, nil))
		return
	}
	if int64(len(fileData)) > policy.MaxSize {
		c.JSON(http.StatusBadRequest, component.FailDataCode(
			component.MessageUploadFileTooLarge,

			component.MessageParams{"maxSizeKb": policy.MaxSize / 1024}))
		return
	}
	if err := validateUploadedImage(bytes.NewReader(fileData), contentType); err != nil {
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageUploadInvalidImage, nil))
		return
	}

	folderName := time.Now().Format("2006/01/02")

	entity, err := filestorage.SaveFileFromUpload(c.Request.Context(), userId, fileData, file.Filename, folderName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailDataCode(
			component.MessageUploadSaveFailed,

			component.MessageParams{"error": err.Error()}))
		return
	}
	if adminUpload {
		fileusageservice.AddAdminUpload(userId, entity.Name)
	} else if err := fileusageservice.AddUploadOwner(userId, entity.Name); err != nil {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 5*time.Second)
		cleanupErr := filestorage.Delete(cleanupCtx, entity.Name)
		cleanupCancel()
		if cleanupErr != nil {
			slog.Error("delete upload after owner usage failure", "fileName", entity.Name, "err", cleanupErr)
		}
		c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageUploadSaveFailed, nil))
		return
	}

	c.JSON(http.StatusOK, component.SuccessDataCode(map[string]any{
		"url":      entity.GetAccessPath(),
		"filename": file.Filename,
		"size":     len(fileData),
	}, component.MessageUploadSuccess, nil))
}

func isAllowedExtension(ext string, allowedExts []string) bool {
	for _, allowedExt := range allowedExts {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}
	return false
}
