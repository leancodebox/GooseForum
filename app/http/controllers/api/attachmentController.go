package api

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/attachmentservice"
	"github.com/leancodebox/GooseForum/app/service/trustlevelservice"
)

func attachmentFailure(c *gin.Context, err error) {
	var oversized *http.MaxBytesError
	if errors.As(err, &oversized) {
		c.JSON(http.StatusBadRequest, component.FailDataCode(component.MessageAttachmentTooLarge, nil))
		return
	}
	var p *trustlevelservice.PolicyError
	if errors.As(err, &p) {
		c.JSON(p.Status, component.FailDataCode(component.MessageCode(p.Code), p.Params))
		return
	}
	c.JSON(http.StatusInternalServerError, component.FailDataCode(component.MessageOperationFailed, nil))
}
func AttachmentPolicy(c *gin.Context) {
	policy, err := attachmentservice.Policy(component.LoginUserId(c))
	if err != nil {
		attachmentFailure(c, err)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.JSON(200, component.SuccessDataCode(policy, component.MessageOperationSuccess, nil))
}
func UploadAttachment(c *gin.Context) {
	uploader, err := attachmentservice.NewUploader(component.LoginUserId(c))
	if err != nil {
		attachmentFailure(c, err)
		return
	}
	requestLimit := uploader.MaxFileSize() + 1024*1024
	if c.Request.ContentLength > requestLimit {
		c.JSON(400, component.FailDataCode(component.MessageAttachmentTooLarge, nil))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, requestLimit)
	reader, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(400, component.FailDataCode(component.MessageUploadFileMissing, nil))
		return
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			var oversized *http.MaxBytesError
			if errors.As(err, &oversized) {
				attachmentFailure(c, err)
			} else {
				c.JSON(400, component.FailDataCode(component.MessageUploadFileMissing, nil))
			}
			return
		}
		if part.FormName() != "file" {
			continue
		}
		// Use the raw filename: multipart.Part.FileName strips path components.
		_, disposition, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if err != nil || disposition["filename"] == "" {
			c.JSON(400, component.FailDataCode(component.MessageAttachmentInvalidFile, nil))
			return
		}
		result, err := uploader.Upload(c.Request.Context(), disposition["filename"], part)
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				c.JSON(400, component.FailDataCode(component.MessageAttachmentInvalidFile, nil))
			} else {
				attachmentFailure(c, err)
			}
			return
		}
		c.JSON(200, component.SuccessDataCode(result, component.MessageUploadSuccess, nil))
		return
	}
}
func DownloadAttachment(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	c.Header("Vary", "Cookie, Authorization")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(404)
		return
	}
	e, obj, err := attachmentservice.Open(c.Request.Context(), component.LoginUserId(c), id)
	if err != nil {
		c.Status(404)
		return
	}
	defer obj.Body.Close()
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": e.Filename}))
	c.DataFromReader(200, obj.Metadata.Size, "application/octet-stream", obj.Body, nil)
}
func CancelAttachment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(404)
		return
	}
	if err := attachmentservice.Cancel(c.Request.Context(), component.LoginUserId(c), id); err != nil {
		c.Status(404)
		return
	}
	c.JSON(200, component.SuccessDataCode(true, component.MessageOperationSuccess, nil))
}
