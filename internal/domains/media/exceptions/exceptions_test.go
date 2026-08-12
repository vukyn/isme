package exceptions

import (
	"errors"
	"testing"

	pkgErr "github.com/vukyn/kuery/http/errors"
	"github.com/vukyn/kuery/medioa"
)

// statusOf unwraps the mapped error to its status and message, failing the test
// if the mapping did not produce a kuery/http/errors error at all.
func statusOf(t *testing.T, err error) (int, string) {
	t.Helper()
	var mapped pkgErr.Error
	if !errors.As(err, &mapped) {
		t.Fatalf("mapped error is not a pkgErr.Error: %T (%v)", err, err)
	}
	return mapped.Status(), mapped.Error()
}

// A 413 must keep medioa's own wording, because that is the only place the cap
// appears — the cap is medioa's UPLOAD_MAX_FILE_SIZE_MB and isme holds no copy of
// it, so a fixed message leaves the user unable to tell how large an avatar may
// be.
func TestMapMediaErrorTooLargeForwardsUpstreamMessage(t *testing.T) {
	const upstream = "file size too large (max: 50MB)"

	status, message := statusOf(t, MapMediaError(&medioa.APIError{
		StatusCode: 413,
		Code:       413,
		Message:    upstream,
	}))

	if status != 413 {
		t.Errorf("status = %d, want 413", status)
	}
	if message != upstream {
		t.Errorf("message = %q, want %q (the cap must survive the mapping)", message, upstream)
	}
}

// A bare 413 with no envelope message falls back to isme's own wording rather
// than an empty message.
func TestMapMediaErrorTooLargeFallsBackWhenUpstreamIsSilent(t *testing.T) {
	status, message := statusOf(t, MapMediaError(&medioa.APIError{StatusCode: 413}))

	if status != 413 {
		t.Errorf("status = %d, want 413", status)
	}
	if message != "uploaded file is too large" {
		t.Errorf("message = %q, want the isme fallback", message)
	}
}

// The forwarding is scoped to the client-actionable statuses. A rejected key is a
// server-config problem the signed-in user cannot fix, so it must still collapse
// to the operator-facing 502 with none of medioa's text.
func TestMapMediaErrorStillSanitizesKeyRejection(t *testing.T) {
	for _, statusCode := range []int{401, 403} {
		status, message := statusOf(t, MapMediaError(&medioa.APIError{
			StatusCode: statusCode,
			Message:    "invalid api key mk_abcdef",
		}))

		if status != 502 {
			t.Errorf("upstream %d: status = %d, want 502", statusCode, status)
		}
		if message != "media service rejected the key" {
			t.Errorf("upstream %d: message = %q, want the sanitized operator message", statusCode, message)
		}
	}
}

func TestMapMediaErrorNotFound(t *testing.T) {
	status, _ := statusOf(t, MapMediaError(&medioa.APIError{StatusCode: 404}))
	if status != 404 {
		t.Errorf("status = %d, want 404", status)
	}
}

func TestMapMediaErrorNilIsNil(t *testing.T) {
	if err := MapMediaError(nil); err != nil {
		t.Errorf("MapMediaError(nil) = %v, want nil", err)
	}
}
