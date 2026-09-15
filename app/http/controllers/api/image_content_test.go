package api

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestValidateUploadedImageAcceptsImageLargerThanHeaderBudget(t *testing.T) {
	var encoded bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	encoded.Write(make([]byte, maximumImageHeaderSize))

	if err := validateUploadedImage(bytes.NewReader(encoded.Bytes()), "image/png"); err != nil {
		t.Fatalf("validate large image: %v", err)
	}
}

func TestValidateUploadedImageRejectsMismatchedContentType(t *testing.T) {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	if err := validateUploadedImage(bytes.NewReader(encoded.Bytes()), "image/jpeg"); err == nil {
		t.Fatal("mismatched image content type was accepted")
	}
}
