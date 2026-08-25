package api

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

const maximumImageHeaderSize = 1024 * 1024

var errInvalidImageContent = errors.New("invalid image content")

var imageFormatContentTypes = map[string]string{
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
	"bmp":  "image/bmp",
}

func validateUploadedImage(reader io.Reader, expectedContentType string) error {
	data, err := io.ReadAll(io.LimitReader(reader, maximumImageHeaderSize+1))
	if err != nil {
		return fmt.Errorf("read image header: %w", err)
	}
	if len(data) == 0 || len(data) > maximumImageHeaderSize {
		return fmt.Errorf("%w: image header is empty or exceeds %d bytes", errInvalidImageContent, maximumImageHeaderSize)
	}
	detected := http.DetectContentType(data)
	if !strings.EqualFold(detected, expectedContentType) {
		return fmt.Errorf("%w: detected content type %q does not match %q", errInvalidImageContent, detected, expectedContentType)
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%w: decode image header: %v", errInvalidImageContent, err)
	}
	decodedContentType, ok := imageFormatContentTypes[format]
	if !ok || !strings.EqualFold(decodedContentType, expectedContentType) {
		return fmt.Errorf("%w: decoded image format %q does not match %q", errInvalidImageContent, format, expectedContentType)
	}
	return nil
}
