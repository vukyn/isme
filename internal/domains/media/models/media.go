package models

import (
	"errors"
	"io"
)

// MaxAvatarSize caps an avatar upload at 2MB (the mock states "max 2MB"). The
// cap is enforced server-side; the UI also pre-checks so the user gets instant
// feedback before the round trip.
const MaxAvatarSize int64 = 2 * 1024 * 1024

// allowedAvatarContentTypes is the MIME allowlist for avatar uploads. Anything
// outside this set is rejected before the file reaches medioa.
var allowedAvatarContentTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
}

// UploadRequest is the parsed multipart upload the isme backend proxies to
// medioa. File content is streamed, never buffered into the model. Size is the
// declared multipart size used for the 2MB cap check.
type UploadRequest struct {
	File        io.Reader
	FileName    string
	ContentType string
	Ext         string
	Size        int64
}

func (r UploadRequest) Validate() error {
	if r.File == nil {
		return errors.New("file is required")
	}
	if r.FileName == "" {
		return errors.New("file name is required")
	}
	if !allowedAvatarContentTypes[r.ContentType] {
		return errors.New("file must be a PNG, JPEG, or WebP image")
	}
	if r.Size > MaxAvatarSize {
		return errors.New("file is over 2MB")
	}
	return nil
}

// UploadResponse is what the isme UI receives after a successful proxied
// upload: a stable, browser-servable URL plus the medioa file id and metadata.
type UploadResponse struct {
	URL      string `json:"url"`
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}
