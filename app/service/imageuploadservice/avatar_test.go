package imageuploadservice

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestValidateDefaultAvatar(t *testing.T) {
	// One opaque WebP pixel encoded with VP8.
	data, err := base64.StdEncoding.DecodeString("UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEAAUAmJaQAA3AA/v89WAAAAA==")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, mime    string
		width, height uint16
		valid         bool
	}{
		{"small", "image/webp", 1, 1, true},
		{"maximum", "image/webp", 256, 256, true},
		{"oversize", "image/webp", 257, 257, false},
		{"rectangular", "image/webp", 256, 128, false},
		{"wrong MIME", "image/png", 1, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			img := bytes.Clone(data)
			img[26], img[27] = byte(tc.width), byte(tc.width>>8)
			img[28], img[29] = byte(tc.height), byte(tc.height>>8)
			if err := validateDefaultAvatar(bytes.NewReader(img), tc.mime); (err == nil) != tc.valid {
				t.Fatalf("validation: %v", err)
			}
		})
	}
	if validateDefaultAvatar(bytes.NewReader([]byte("invalid")), "image/webp") == nil {
		t.Fatal("invalid WebP accepted")
	}
}
