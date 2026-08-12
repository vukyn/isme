package exceptions

import (
	"errors"

	pkgBase "github.com/vukyn/kuery/http/base"
	pkgErr "github.com/vukyn/kuery/http/errors"
	"github.com/vukyn/kuery/medioa"
)

// MapMediaError translates a kuery/medioa SDK error into an isme HTTP error.
//
// The medioa API key is an isme server-side concern, so key problems
// (401/403 from medioa) collapse to a 502 with an operator-facing message —
// the signed-in user can't fix the key, but the operator must. Other medioa
// failures map to the closest matching isme status.
func MapMediaError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, medioa.ErrUnauthorized), errors.Is(err, medioa.ErrForbidden):
		// The key is missing/revoked/expired or its owner lost bucket
		// membership — a server config problem, not a client one.
		return pkgErr.Forward(pkgBase.Response{Code: 502, Message: "media service rejected the key"})
	case errors.Is(err, medioa.ErrTooLarge):
		// Forward medioa's wording. It reads "file size too large (max: NMB)", and
		// that cap is medioa's own UPLOAD_MAX_FILE_SIZE_MB — a value isme holds no
		// copy of, so replacing the text left the user with no way to tell how large
		// an avatar is allowed to be. It exposes no server detail worth hiding: no
		// key, no bucket, no path. Falls back to isme's own wording when medioa
		// answers a bare 413 with no envelope message.
		var tooLargeErr *medioa.APIError
		if errors.As(err, &tooLargeErr) && tooLargeErr.Message != "" {
			return pkgErr.Forward(pkgBase.Response{Code: 413, Message: tooLargeErr.Message})
		}
		return pkgErr.Forward(pkgBase.Response{Code: 413, Message: "uploaded file is too large"})
	case errors.Is(err, medioa.ErrNotFound):
		return pkgErr.NotFound("media object not found")
	default:
		// An *APIError with no dedicated sentinel (e.g. medioa 400 with an R2
		// storage detail) — forward medioa's own message so the operator can
		// diagnose, rather than collapsing to an opaque "media upload failed".
		var apiErr *medioa.APIError
		if errors.As(err, &apiErr) && apiErr.Message != "" {
			return pkgErr.Forward(pkgBase.Response{Code: 502, Message: "media upload failed: " + apiErr.Message})
		}
		// Network failures, decode errors — surface as an upstream failure.
		return pkgErr.Forward(pkgBase.Response{Code: 502, Message: "media upload failed"})
	}
}
