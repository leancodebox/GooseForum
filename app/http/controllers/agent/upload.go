package agent

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/service/imageuploadservice"
)

const uploadMultipartOverhead = 64 * 1024

func (h *Handler) uploadImage(c *gin.Context) {
	if !writeAllowed(c) {
		return
	}
	policy, failure := imageuploadservice.ResolvePolicy(current(c).User)
	if failure != nil {
		uploadError(c, failure)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, policy.MaxSize+uploadMultipartOverhead)
	if err := c.Request.ParseMultipartForm(uploadMultipartOverhead); err != nil {
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			fail(c, 413, "upload_too_large", "Image request exceeds the upload limit")
		} else {
			fail(c, 400, "invalid_request", "A multipart file field is required")
		}
		return
	}
	defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	files := c.Request.MultipartForm.File["file"]
	if len(files) != 1 || len(c.Request.MultipartForm.File) != 1 || len(c.Request.MultipartForm.Value) != 0 {
		fail(c, 400, "invalid_request", "Exactly one file field is required")
		return
	}
	file := files[0]
	contentType, failure := policy.Validate(file.Filename, file.Size, "")
	if failure != nil {
		uploadError(c, failure)
		return
	}
	src, err := file.Open()
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read image")
		return
	}
	defer func() { _ = src.Close() }()
	data, err := io.ReadAll(io.LimitReader(src, policy.MaxSize+1))
	if err != nil {
		fail(c, 500, "internal_error", "Unable to read image")
		return
	}
	if int64(len(data)) > policy.MaxSize {
		fail(c, 413, "upload_too_large", "Image exceeds the upload limit")
		return
	}
	if err := imageuploadservice.ValidateContent(bytes.NewReader(data), contentType); err != nil {
		fail(c, 422, "invalid_image", "Image content does not match a supported image format")
		return
	}
	entity, err := imageuploadservice.Save(c.Request.Context(), current(c).User.Id, data, file.Filename, false)
	if err != nil {
		slog.Error("agent image upload failed", "userId", current(c).User.Id, "requestId", requestID(c), "err", err)
		fail(c, 500, "upload_failed", "Unable to save image")
		return
	}
	success(c, 201, gin.H{"url": entity.GetAccessPath(), "filename": file.Filename, "size": entity.Size})
}

func uploadError(c *gin.Context, failure *imageuploadservice.Failure) {
	status, code := failure.Status, failure.Code
	details := failure.Params
	if details == nil {
		details = map[string]any{}
	}
	switch code {
	case "upload.cooldown", "upload.dailyLimit":
		status = 429
	case "upload.file.tooLarge":
		status = 413
	case "upload.extension.unsupported", "upload.image.unsupported", "upload.image.invalidContent":
		status = 422
	}
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": "Image upload rejected", "details": details}, "requestId": requestID(c)})
}
