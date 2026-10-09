package api

import (
	"net/http"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
	"github.com/leancodebox/GooseForum/app/service/imageuploadservice"
)

type imageUploadPolicy imageuploadservice.Policy

type imageUploadFailure struct {
	Status int
	Data   component.ResultStruct
}

func resolveImageUploadPolicy(userID uint64) (*imageUploadPolicy, *imageUploadFailure) {
	return resolveImageUploadPolicyFor(userID, false)
}

func resolveAdminImageUploadPolicy(userID uint64) (*imageUploadPolicy, *imageUploadFailure) {
	return resolveImageUploadPolicyFor(userID, true)
}

func resolveImageUploadPolicyFor(userID uint64, adminUpload bool) (*imageUploadPolicy, *imageUploadFailure) {
	if userID == 0 {
		return nil, uploadFailure(http.StatusUnauthorized, component.MessageAuthRequired, nil)
	}
	if adminUpload {
		return &imageUploadPolicy{MaxSize: int64(filestorage.MaxFileSize), AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}}, nil
	}
	user, err := users.Get(userID)
	if err != nil {
		return nil, uploadFailure(http.StatusForbidden, component.MessagePermissionResolveFailed, nil)
	}
	if status, err := component.CheckUserPermission(&user, component.PermissionActionUploadAttachment); err != nil {
		return nil, &imageUploadFailure{Status: status, Data: component.FailDataError(err)}
	}
	policy, failure := imageuploadservice.ResolvePolicy(user)
	if failure != nil {
		return nil, uploadFailure(failure.Status, component.MessageCode(failure.Code), component.MessageParams(failure.Params))
	}
	return (*imageUploadPolicy)(policy), nil
}

func (p imageUploadPolicy) Validate(filename string, size int64, contentType string) (string, *imageUploadFailure) {
	value, failure := imageuploadservice.Policy(p).Validate(filename, size, contentType)
	if failure != nil {
		return "", uploadFailure(failure.Status, component.MessageCode(failure.Code), component.MessageParams(failure.Params))
	}
	return value, nil
}

func uploadFailure(status int, code component.MessageCode, params component.MessageParams) *imageUploadFailure {
	return &imageUploadFailure{Status: status, Data: component.FailDataCode(code, params)}
}
