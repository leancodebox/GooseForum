package api

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
)

type directImageUploadInitRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

type directImageUploadCompleteRequest struct {
	Name string `json:"name"`
}

type directImageUploadInitResult struct {
	Mode   string                    `json:"mode"`
	Name   string                    `json:"name,omitempty"`
	Upload *filestorage.DirectUpload `json:"upload,omitempty"`
}

func InitDirectImageUpload(c *gin.Context) {
	var request directImageUploadInitRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageRequestParseFailed, component.MessageParams{"error": err.Error()}))
		return
	}
	if !filestorage.SupportsDirectUpload() {
		c.JSON(http.StatusOK, component.SuccessDataCode(directImageUploadInitResult{Mode: "proxy"}, component.MessageOperationSuccess, nil))
		return
	}
	userId := c.GetUint64("userId")
	policy, failure := resolveImageUploadPolicy(userId)
	if failure != nil {
		c.JSON(failure.Status, failure.Data)
		return
	}
	contentType, failure := policy.Validate(request.Filename, request.Size, request.ContentType)
	if failure != nil {
		c.JSON(failure.Status, failure.Data)
		return
	}
	name := filestorage.NewUploadName(request.Filename, time.Now().Format("2006/01/02"))
	session, err := filestorage.BeginDirectUpload(c.Request.Context(), filestorage.DirectUploadRequest{
		Name: name, ContentType: contentType, Size: request.Size, UserId: userId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageUploadSaveFailed, nil))
		return
	}
	c.JSON(http.StatusOK, component.SuccessDataCode(directImageUploadInitResult{
		Mode: "direct", Name: session.Metadata.Name, Upload: &session.Upload,
	}, component.MessageOperationSuccess, nil))
}

func CompleteDirectImageUpload(c *gin.Context) {
	var request directImageUploadCompleteRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Name == "" {
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageRequestInvalidParams, nil))
		return
	}
	userId := c.GetUint64("userId")
	metadata, err := filestorage.CompleteDirectUpload(c.Request.Context(), filestorage.CompleteDirectUploadRequest{
		Name: request.Name, UserId: userId,
		Validator: func(reader io.Reader, contentType string) error {
			err := validateUploadedImage(reader, contentType)
			if errors.Is(err, errInvalidImageContent) {
				return errors.Join(filestorage.ErrDirectUploadInvalidObject, err)
			}
			return err
		},
	})
	if err != nil {
		writeDirectUploadError(c, err)
		return
	}
	if err := fileusageservice.AddUploadOwner(userId, metadata.Name); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 5*time.Second)
		cleanupErr := filestorage.Delete(cleanupCtx, metadata.Name)
		cancel()
		if cleanupErr != nil {
			slog.Error("delete direct upload after owner usage failure", "fileName", metadata.Name, "err", cleanupErr)
		}
		c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageUploadSaveFailed, nil))
		return
	}
	c.JSON(http.StatusOK, component.SuccessDataCode(map[string]any{
		"url": metadata.GetAccessPath(), "filename": metadata.Name, "size": metadata.Size,
	}, component.MessageUploadSuccess, nil))
}

func AbortDirectImageUpload(c *gin.Context) {
	var request directImageUploadCompleteRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Name == "" {
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageRequestInvalidParams, nil))
		return
	}
	err := filestorage.AbortDirectUpload(c.Request.Context(), filestorage.CompleteDirectUploadRequest{
		Name: request.Name, UserId: c.GetUint64("userId"),
	})
	if err != nil {
		writeDirectUploadError(c, err)
		return
	}
	c.JSON(http.StatusOK, component.SuccessDataCode(true, component.MessageOperationSuccess, nil))
}

func writeDirectUploadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, filestorage.ErrDirectUploadOwnerMismatch):
		c.JSON(http.StatusNotFound, component.FailDataCode(component.MessagePageNotFound, nil))
	case errors.Is(err, filestorage.ErrDirectUploadInvalidObject):
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageUploadInvalidImage, nil))
	case errors.Is(err, filestorage.ErrDirectUploadUnsupported):
		c.JSON(http.StatusConflict, component.FailDataCode(component.MessageOperationFailed, nil))
	default:
		c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageUploadSaveFailed, nil))
	}
}
