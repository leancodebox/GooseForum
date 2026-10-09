package api

import (
	"io"

	"github.com/leancodebox/GooseForum/app/service/imageuploadservice"
)

var errInvalidImageContent = imageuploadservice.ErrInvalidImageContent

const maximumImageHeaderSize = imageuploadservice.MaximumImageHeaderSize

func validateUploadedImage(reader io.Reader, contentType string) error {
	return imageuploadservice.ValidateContent(reader, contentType)
}
