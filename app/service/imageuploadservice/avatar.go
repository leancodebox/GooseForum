package imageuploadservice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"strings"

	"github.com/leancodebox/GooseForum/app/service/filestorage"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
	"golang.org/x/image/draw"
	"gorm.io/gorm"
)

const avatarPrefix = "avatars/default/v1/"

// DefaultAvatar returns a shared thumbnail for a previously verified administrator upload.
func DefaultAvatar(ctx context.Context, source string) (string, error) {
	if strings.HasPrefix(source, avatarPrefix) {
		return source, nil
	}
	key := fmt.Sprintf("%s%x.png", avatarPrefix, sha256.Sum256([]byte(source)))
	meta, err := filestorage.Stat(ctx, key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		obj, err := filestorage.Open(ctx, source)
		if err != nil {
			return "", err
		}
		defer obj.Body.Close()
		data, err := io.ReadAll(io.LimitReader(obj.Body, filestorage.MaxFileSize+1))
		if err != nil {
			return "", err
		}
		thumbnail, err := avatarThumbnail(data)
		if err != nil {
			return "", err
		}
		meta, err = filestorage.Put(ctx, filestorage.PutRequest{Name: key, ContentType: "image/png", Size: int64(len(thumbnail)), Data: thumbnail, UserId: obj.Metadata.UserId})
		if err != nil {
			// Another save may have finished creating the same immutable image.
			existing, lookupErr := filestorage.Stat(ctx, key)
			if lookupErr != nil {
				return "", err
			}
			meta = existing
		}
	} else if err != nil {
		return "", err
	}
	if err := fileusageservice.AddAdminUpload(meta.UserId, key); err != nil {
		return "", err
	}
	return key, nil
}

func avatarThumbnail(data []byte) ([]byte, error) {
	if len(data) > filestorage.MaxFileSize {
		return nil, errors.New("avatar exceeds upload limit")
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	// Bound decoded memory before asking the format decoder to allocate it.
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 16_000_000/cfg.Height {
		return nil, errors.New("avatar exceeds 16 megapixels")
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := src.Bounds()
	side := min(bounds.Dx(), bounds.Dy())
	origin := bounds.Min.Add(image.Pt((bounds.Dx()-side)/2, (bounds.Dy()-side)/2))
	size := min(side, 256)
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, image.Rectangle{Min: origin, Max: origin.Add(image.Pt(side, side))}, draw.Src, nil)
	var out bytes.Buffer
	if err := png.Encode(&out, dst); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
