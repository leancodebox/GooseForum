package imageuploadservice

import (
	"bufio"
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

const MaximumImageHeaderSize = 1024 * 1024

var ErrInvalidImageContent = errors.New("invalid image content")

var imageFormatContentTypes = map[string]string{
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
	"bmp":  "image/bmp",
}

func ValidateContent(reader io.Reader, expectedContentType string) error {
	limited := bufio.NewReader(io.LimitReader(reader, MaximumImageHeaderSize))
	header, err := limited.Peek(512)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, bufio.ErrBufferFull) {
		return fmt.Errorf("read image header: %w", err)
	}
	if len(header) == 0 {
		return fmt.Errorf("%w: image is empty", ErrInvalidImageContent)
	}
	detected := http.DetectContentType(header)
	if !strings.EqualFold(detected, expectedContentType) {
		return fmt.Errorf("%w: detected content type %q does not match %q", ErrInvalidImageContent, detected, expectedContentType)
	}
	_, format, err := image.DecodeConfig(limited)
	if err != nil {
		return fmt.Errorf("%w: decode image header: %v", ErrInvalidImageContent, err)
	}
	decodedContentType, ok := imageFormatContentTypes[format]
	if !ok || !strings.EqualFold(decodedContentType, expectedContentType) {
		return fmt.Errorf("%w: decoded image format %q does not match %q", ErrInvalidImageContent, format, expectedContentType)
	}
	return nil
}
