package imageuploadservice

import (
	"context"
	"errors"
	"image"
	"io"

	"github.com/leancodebox/GooseForum/app/service/filestorage"
)

func ValidateDefaultAvatar(ctx context.Context, key string) error {
	obj, err := filestorage.Open(ctx, key)
	if err != nil {
		return err
	}
	defer obj.Body.Close()
	return validateDefaultAvatar(obj.Body, obj.Metadata.ContentType)
}

func validateDefaultAvatar(reader io.Reader, contentType string) error {
	cfg, format, err := image.DecodeConfig(io.LimitReader(reader, MaximumImageHeaderSize))
	if err != nil || contentType != "image/webp" || format != "webp" || cfg.Width < 1 || cfg.Width > 256 || cfg.Width != cfg.Height {
		return errors.New("default avatar must be a square WebP image up to 256 × 256 pixels")
	}
	return nil
}
