package imageuploadservice

import (
	"context"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
)

type Policy struct {
	MaxSize     int64
	AllowedExts []string
}

type Failure struct {
	Status int
	Code   string
	Params map[string]any
}

func failure(status int, code string, params map[string]any) *Failure {
	return &Failure{Status: status, Code: code, Params: params}
}

// ResolvePolicy assumes the caller has checked current account write permissions.
func ResolvePolicy(user users.EntityComplete) (*Policy, *Failure) {
	config := hotdataserve.GetPostingSettingsConfigCache().UploadControl
	isRoleUser := user.RoleId > 0
	if !isRoleUser && !config.AllowAttachments {
		return nil, failure(http.StatusForbidden, "upload.attachment.disabled", nil)
	}
	if !isRoleUser && config.NewUserUploadCooldownMinutes > 0 {
		available := user.CreatedAt.Add(time.Duration(config.NewUserUploadCooldownMinutes) * time.Minute)
		if time.Now().Before(available) {
			return nil, failure(http.StatusBadRequest, "upload.cooldown", map[string]any{"minutes": config.NewUserUploadCooldownMinutes, "availableAt": available.Format("2006-01-02 15:04:05")})
		}
	}
	if !isRoleUser && config.MaxDailyUploadsPerUser > 0 {
		count := filedata.CountDailyUploads(user.Id)
		if count >= int64(config.MaxDailyUploadsPerUser) {
			return nil, failure(http.StatusBadRequest, "upload.dailyLimit", map[string]any{"count": count})
		}
	}
	maxSize := int64(filestorage.MaxFileSize)
	configMax := int64(config.MaxAttachmentSizeKb) * 1024
	if !isRoleUser && configMax > 0 && configMax < maxSize {
		maxSize = configMax
	}
	return &Policy{MaxSize: maxSize, AllowedExts: config.AuthorizedExtensions}, nil
}

func (p Policy) Validate(filename string, size int64, reportedContentType string) (string, *Failure) {
	if strings.TrimSpace(filename) == "" {
		return "", failure(400, "upload.filename.required", nil)
	}
	if size <= 0 {
		return "", failure(400, "upload.image.invalidContent", nil)
	}
	if size > p.MaxSize {
		return "", failure(400, "upload.file.tooLarge", map[string]any{"maxSizeKb": p.MaxSize / 1024})
	}
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := len(p.AllowedExts) == 0
	for _, value := range p.AllowedExts {
		if strings.EqualFold(value, ext) {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", failure(400, "upload.extension.unsupported", map[string]any{"extensions": strings.Join(p.AllowedExts, ", ")})
	}
	contentType, err := filedata.CheckImageType(filename)
	if err != nil {
		return "", failure(400, "upload.image.unsupported", nil)
	}
	if reportedContentType != "" {
		reported, _, err := mime.ParseMediaType(reportedContentType)
		if err != nil || !strings.EqualFold(reported, contentType) {
			return "", failure(400, "upload.image.invalidContent", nil)
		}
	}
	return contentType, nil
}

func Save(ctx context.Context, userID uint64, data []byte, filename string, adminUpload bool) (*filestorage.Metadata, error) {
	entity, err := filestorage.SaveFileFromUpload(ctx, userID, data, filename, time.Now().Format("2006/01/02"))
	if err != nil {
		return nil, err
	}
	if adminUpload {
		fileusageservice.AddAdminUpload(userID, entity.Name)
		return entity, nil
	}
	if err := fileusageservice.AddUploadOwner(userID, entity.Name); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		cleanupErr := filestorage.Delete(cleanupCtx, entity.Name)
		cancel()
		if cleanupErr != nil {
			slog.Error("delete upload after owner usage failure", "fileName", entity.Name, "err", cleanupErr)
		}
		return nil, err
	}
	return entity, nil
}
