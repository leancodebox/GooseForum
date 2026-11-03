package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/attachmentservice"
	"github.com/leancodebox/GooseForum/app/service/trustlevelservice"
	"mime"
	"net/http"
	"strconv"
)

func attachmentFailure(c *gin.Context, err error) {
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, attachmentservice.HardMaxSize+1024*1024)
	defer func() {
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
	}()
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, component.FailDataCode(component.MessageUploadFileMissing, nil))
		return
	}
	// multipart.FileHeader.Filename has already been reduced to a basename by
	// net/http. Validate the original name so path components are not silently accepted.
	_, disposition, err := mime.ParseMediaType(file.Header.Get("Content-Disposition"))
	if err != nil {
		c.JSON(400, component.FailDataCode("attachment.invalidFile", nil))
		return
	}
	filename := disposition["filename"]
	src, err := file.Open()
	if err != nil {
		attachmentFailure(c, err)
		return
	}
	defer src.Close()
	result, err := attachmentservice.Upload(c.Request.Context(), component.LoginUserId(c), filename, src)
	if err != nil {
		attachmentFailure(c, err)
		return
	}
	c.JSON(200, component.SuccessDataCode(result, component.MessageUploadSuccess, nil))
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
