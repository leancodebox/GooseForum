package api

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
)

func TestImageUploadPolicyValidatesFileMetadata(t *testing.T) {
	policy := imageUploadPolicy{MaxSize: 1024, AllowedExts: []string{".jpg", ".png", ".webp"}}
	tests := []struct {
		name        string
		filename    string
		size        int64
		contentType string
		wantType    string
		wantCode    component.MessageCode
	}{
		{name: "valid", filename: "photo.webp", size: 512, contentType: "image/webp", wantType: "image/webp"},
		{name: "content type parameters", filename: "photo.jpg", size: 512, contentType: "image/jpeg; charset=binary", wantType: "image/jpeg"},
		{name: "empty name", size: 1, wantCode: component.MessageUploadFilenameRequired},
		{name: "empty file", filename: "photo.png", wantCode: component.MessageUploadInvalidImage},
		{name: "too large", filename: "photo.png", size: 1025, wantCode: component.MessageUploadFileTooLarge},
		{name: "extension denied", filename: "photo.gif", size: 512, contentType: "image/gif", wantCode: component.MessageUploadUnsupportedExt},
		{name: "unsupported image", filename: "photo.svg", size: 512, contentType: "image/svg+xml", wantCode: component.MessageUploadUnsupportedExt},
		{name: "mime mismatch", filename: "photo.png", size: 512, contentType: "image/jpeg", wantCode: component.MessageUploadInvalidImage},
		{name: "bad mime", filename: "photo.png", size: 512, contentType: "not a mime", wantCode: component.MessageUploadInvalidImage},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contentType, failure := policy.Validate(test.filename, test.size, test.contentType)
			if contentType != test.wantType {
				t.Fatalf("content type = %q, want %q", contentType, test.wantType)
			}
			if test.wantCode == "" {
				if failure != nil {
					t.Fatalf("unexpected failure = %#v", failure)
				}
				return
			}
			if failure == nil || failure.Data.MessageCode != test.wantCode {
				t.Fatalf("failure = %#v, want %q", failure, test.wantCode)
			}
		})
	}
}

func TestValidateUploadedImageChecksDecodedFormat(t *testing.T) {
	imageData := encodeTestPNG(t)
	if err := validateUploadedImage(bytes.NewReader(imageData), "image/png"); err != nil {
		t.Fatalf("valid png: %v", err)
	}
	if err := validateUploadedImage(bytes.NewReader(imageData), "image/jpeg"); err == nil {
		t.Fatal("mime mismatch succeeded")
	}
	if err := validateUploadedImage(strings.NewReader("\x89PNG\r\n\x1a\nnot-an-image"), "image/png"); err == nil {
		t.Fatal("signature-only png succeeded")
	} else if !errors.Is(err, errInvalidImageContent) {
		t.Fatalf("invalid image error = %v", err)
	}
	if err := validateUploadedImage(strings.NewReader(""), "image/png"); err == nil {
		t.Fatal("empty image succeeded")
	}
	readErr := errors.New("temporary read failure")
	if err := validateUploadedImage(errorReader{err: readErr}, "image/png"); !errors.Is(err, readErr) || errors.Is(err, errInvalidImageContent) {
		t.Fatalf("read error = %v", err)
	}
}

type errorReader struct{ err error }

func (reader errorReader) Read([]byte) (int, error) { return 0, reader.err }

var _ io.Reader = errorReader{}

func encodeTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.White)
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buffer.Bytes()
}
