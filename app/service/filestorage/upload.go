package filestorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
)

const (
	MaxFileSize = 4 * 1024 * 1024
	AvatarPath  = "avatars"
)

type AvatarUpload struct {
	Filename string
	Data     []byte
}

func SaveFileFromUpload(ctx context.Context, userId uint64, data []byte, filename string, customPath string) (*Metadata, error) {
	if len(data) > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of %dMB", MaxFileSize/(1024*1024))
	}
	contentType, err := filedata.CheckImageType(filename)
	if err != nil {
		return nil, err
	}
	name := NewUploadName(filename, customPath)
	return Put(ctx, PutRequest{
		Name:        name,
		ContentType: contentType,
		Size:        int64(len(data)),
		UserId:      userId,
		Body:        bytes.NewReader(data),
	})
}

func NewUploadName(filename string, customPath string) string {
	return fmt.Sprintf("%s/%s%s", customPath, uuid.New().String(), strings.ToLower(path.Ext(filename)))
}

func SaveAvatar(ctx context.Context, userId uint64, data []byte, filename string) (*Metadata, error) {
	avatarPath := fmt.Sprintf("%s/avatar_%d_%d", AvatarPath, userId, time.Now().Unix())
	return SaveFileFromUpload(ctx, userId, data, filename, avatarPath)
}

func SaveAvatarSet(ctx context.Context, userId uint64, uploads []AvatarUpload) ([]*Metadata, error) {
	if len(uploads) == 0 {
		return nil, errors.New("avatar files are required")
	}
	if len(uploads) > 2 {
		return nil, errors.New("avatar files exceed maximum limit of 2")
	}

	avatarPath := fmt.Sprintf("%s/%d/%d", AvatarPath, userId, time.Now().UnixNano())
	avatarNames := []string{"avatar", "avatar_medium"}
	stored := make([]*Metadata, 0, len(uploads))
	for index, upload := range uploads {
		if len(upload.Data) > MaxFileSize {
			return nil, fmt.Errorf("file size exceeds maximum limit of %dMB", MaxFileSize/(1024*1024))
		}
		contentType, err := filedata.CheckImageType(upload.Filename)
		if err != nil {
			return nil, err
		}
		name := fmt.Sprintf("%s/%s%s", avatarPath, avatarNames[index], strings.ToLower(path.Ext(upload.Filename)))
		metadata, err := Put(ctx, PutRequest{
			Name:        name,
			ContentType: contentType,
			Size:        int64(len(upload.Data)),
			UserId:      userId,
			Body:        bytes.NewReader(upload.Data),
		})
		if err != nil {
			return nil, err
		}
		stored = append(stored, metadata)
	}
	return stored, nil
}
