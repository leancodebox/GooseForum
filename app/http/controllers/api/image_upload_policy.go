package api

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
)

type imageUploadPolicy struct {
	MaxSize     int64
	AllowedExts []string
}

type imageUploadFailure struct {
	Status int
	Data   component.ResultStruct
}

func resolveImageUploadPolicy(userId uint64) (*imageUploadPolicy, *imageUploadFailure) {
	if userId == 0 {
		return nil, uploadFailure(http.StatusUnauthorized, component.MessageAuthRequired, nil)
	}
	postingConfig := hotdataserve.GetPostingSettingsConfigCache()
	userEntity, _ := users.Get(userId)
	isRoleUser := userEntity.RoleId > 0
	if !isRoleUser && !postingConfig.UploadControl.AllowAttachments {
		return nil, uploadFailure(http.StatusForbidden, component.MessageUploadAttachmentDisabled, nil)
	}
	if status, err := component.CheckUserPermission(&userEntity, component.PermissionActionUploadAttachment); err != nil {
		return nil, &imageUploadFailure{Status: status, Data: component.FailDataError(err)}
	}
	if !isRoleUser && postingConfig.UploadControl.NewUserUploadCooldownMinutes > 0 {
		cooldownTime := userEntity.CreatedAt.Add(time.Duration(postingConfig.UploadControl.NewUserUploadCooldownMinutes) * time.Minute)
		if time.Now().Before(cooldownTime) {
			return nil, uploadFailure(http.StatusBadRequest, component.MessageUploadCooldown, component.MessageParams{
				"minutes":     postingConfig.UploadControl.NewUserUploadCooldownMinutes,
				"availableAt": cooldownTime.Format("2006-01-02 15:04:05"),
			})
		}
	}
	if !isRoleUser && postingConfig.UploadControl.MaxDailyUploadsPerUser > 0 {
		count := filedata.CountDailyUploads(userId)
		if count >= int64(postingConfig.UploadControl.MaxDailyUploadsPerUser) {
			return nil, uploadFailure(http.StatusBadRequest, component.MessageUploadDailyLimit, component.MessageParams{"count": count})
		}
	}
	maxSize := int64(filestorage.MaxFileSize)
	configMaxSize := int64(postingConfig.UploadControl.MaxAttachmentSizeKb) * 1024
	if !isRoleUser && configMaxSize > 0 && configMaxSize < maxSize {
		maxSize = configMaxSize
	}
	return &imageUploadPolicy{MaxSize: maxSize, AllowedExts: postingConfig.UploadControl.AuthorizedExtensions}, nil
}

func (policy imageUploadPolicy) Validate(filename string, size int64, reportedContentType string) (string, *imageUploadFailure) {
	if strings.TrimSpace(filename) == "" {
		return "", uploadFailure(http.StatusBadRequest, component.MessageUploadFilenameRequired, nil)
	}
	if size <= 0 {
		return "", uploadFailure(http.StatusBadRequest, component.MessageUploadInvalidImage, nil)
	}
	if size > policy.MaxSize {
		return "", uploadFailure(http.StatusBadRequest, component.MessageUploadFileTooLarge, component.MessageParams{"maxSizeKb": policy.MaxSize / 1024})
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if len(policy.AllowedExts) > 0 && !isAllowedExtension(ext, policy.AllowedExts) {
		return "", uploadFailure(http.StatusBadRequest, component.MessageUploadUnsupportedExt, component.MessageParams{"extensions": strings.Join(policy.AllowedExts, ", ")})
	}
	contentType, err := filedata.CheckImageType(filename)
	if err != nil {
		return "", uploadFailure(http.StatusBadRequest, component.MessageUploadUnsupportedImage, nil)
	}
	if reportedContentType != "" {
		reported, _, err := mime.ParseMediaType(reportedContentType)
		if err != nil || !strings.EqualFold(reported, contentType) {
			return "", uploadFailure(http.StatusBadRequest, component.MessageUploadInvalidImage, nil)
		}
	}
	return contentType, nil
}

func uploadFailure(status int, code component.MessageCode, params component.MessageParams) *imageUploadFailure {
	return &imageUploadFailure{Status: status, Data: component.FailDataCode(code, params)}
}
