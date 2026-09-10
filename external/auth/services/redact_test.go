package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The bearer token and the session cookie are the two things a debug dump would
// otherwise hand to anyone who can read a log file, so both must be gone and the
// masking must survive a header set with a non-canonical case.
func TestRedactHeadersMasksCredentials(t *testing.T) {
	header := http.Header{}
	header.Set("Authorization", "Bearer real-access-token")
	header.Set("cookie", "session=real-session-value")
	header.Set("Content-Type", "application/json")

	redactHeaders(header)

	if got := header.Get("Authorization"); got != redactedValue {
		t.Errorf("Authorization = %q, want %q", got, redactedValue)
	}
	if got := header.Get("Cookie"); got != redactedValue {
		t.Errorf("Cookie = %q, want %q", got, redactedValue)
	}
	if got := header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want it left alone — redaction must not eat harmless headers", got)
	}
}

// A header that was never present must not be invented: adding an empty
// "[REDACTED]" Authorization line to every dump would be misleading.
func TestRedactHeadersLeavesAbsentHeadersAbsent(t *testing.T) {
	header := http.Header{}
	header.Set("Content-Type", "application/json")

	redactHeaders(header)

	if _, present := header["Authorization"]; present {
		t.Error("Authorization was added to a header set that did not have one")
	}
}

// The refresh-token request body is the highest-value payload on this client:
// the token in it is enough to mint new access tokens.
func TestRedactBodyMasksTopLevelCredentials(t *testing.T) {
	body := redactBody(`{"refresh_token":"real-refresh","app_code":"rainy"}`)

	if strings.Contains(body, "real-refresh") {
		t.Fatalf("refresh_token value survived redaction: %s", body)
	}
	if !strings.Contains(body, "rainy") {
		t.Errorf("non-sensitive app_code was dropped, leaving the dump useless: %s", body)
	}
}

// Login and refresh responses nest both tokens one level down, inside "data".
// A top-level-only redaction would pass this body straight through.
func TestRedactBodyMasksNestedCredentials(t *testing.T) {
	body := redactBody(`{"code":200,"data":{"access_token":"real-access","refresh_token":"real-refresh","user_id":"u1"}}`)

	for _, secret := range []string{"real-access", "real-refresh"} {
		if strings.Contains(body, secret) {
			t.Errorf("nested credential %q survived redaction: %s", secret, body)
		}
	}
	if !strings.Contains(body, "u1") {
		t.Errorf("nested non-sensitive user_id was dropped: %s", body)
	}
}

// app_secret and authorization_code are credentials in the request direction.
func TestRedactBodyMasksRequestDirectionCredentials(t *testing.T) {
	body := redactBody(`{"app_code":"rainy","app_secret":"real-secret","authorization_code":"real-code"}`)

	for _, secret := range []string{"real-secret", "real-code"} {
		if strings.Contains(body, secret) {
			t.Errorf("%q survived redaction: %s", secret, body)
		}
	}
}

// Credentials inside an array element must be reached too.
func TestRedactBodyRecursesThroughArrays(t *testing.T) {
	body := redactBody(`{"sessions":[{"access_token":"real-a"},{"access_token":"real-b"}]}`)

	for _, secret := range []string{"real-a", "real-b"} {
		if strings.Contains(body, secret) {
			t.Errorf("credential %q inside an array survived redaction: %s", secret, body)
		}
	}
}

// Failing closed: a body that is not JSON is replaced wholesale rather than
// passed through, because its shape is unknown and guessing is how a credential
// escapes.
func TestRedactBodyFailsClosedOnNonJSON(t *testing.T) {
	if got := redactBody("Bearer real-token in some plain text"); got != redactedValue {
		t.Errorf("non-JSON body = %q, want it replaced with %q", got, redactedValue)
	}
}

// An empty body stays empty — replacing it would add noise on every GET.
func TestRedactBodyLeavesEmptyBodyEmpty(t *testing.T) {
	if got := redactBody(""); got != "" {
		t.Errorf("empty body = %q, want it left empty", got)
	}
}

// The ordinary error body isme returns must stay readable, since diagnosing a
// 400 is the whole reason the debug client exists.
func TestRedactBodyPreservesErrorPayload(t *testing.T) {
	body := redactBody(`{"code":400,"message":"invalid refresh token"}`)

	var decoded map[string]any
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("redacted error body is no longer valid JSON: %v", err)
	}
	if decoded["message"] != "invalid refresh token" {
		t.Errorf("message = %v, want the original error text preserved", decoded["message"])
	}
}

// captureLogger collects everything resty writes so a test can assert on the
// actual debug dump rather than on the redaction helpers in isolation.
type captureLogger struct{ lines []string }

func (c *captureLogger) Errorf(format string, v ...any) { c.record(format, v...) }
func (c *captureLogger) Warnf(format string, v ...any)  { c.record(format, v...) }
func (c *captureLogger) Debugf(format string, v ...any) { c.record(format, v...) }
func (c *captureLogger) record(format string, v ...any) {
	c.lines = append(c.lines, fmt.Sprintf(format, v...))
}
func (c *captureLogger) all() string { return strings.Join(c.lines, "\n") }

// TestRestWithDebugDumpCarriesNoCredentials is the test that matters, and the
// one the unit tests above do NOT give you: it drives a real request through the
// real client and asserts on what resty actually logged.
//
// The helpers being correct proves nothing if restWithDebug forgets to install
// them — deleting the OnRequestLog/OnResponseLog wiring leaves every unit test
// above passing. This one fails.
func TestRestWithDebugDumpCarriesNoCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"data":{"access_token":"secret-access","refresh_token":"secret-refresh"}}`))
	}))
	defer server.Close()

	capture := &captureLogger{}
	client := (&service{endpoint: server.URL}).
		restWithDebug(context.Background(), 0, 0, time.Second).
		SetLogger(capture)

	_, err := client.R().
		SetHeader("Authorization", "Bearer secret-bearer").
		SetBody(map[string]string{"refresh_token": "secret-sent"}).
		Post("/auth/refresh")
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	dump := capture.all()
	if dump == "" {
		t.Fatal("nothing was logged — the debug client is not dumping, so this test proves nothing")
	}
	for _, secret := range []string{"secret-bearer", "secret-sent", "secret-access", "secret-refresh"} {
		if strings.Contains(dump, secret) {
			t.Errorf("credential %q reached the debug log:\n%s", secret, dump)
		}
	}
}

// TestRestWithDebugEmitsNoCurlCommand guards the deliberate omission of
// EnableGenerateCurlOnDebug. resty writes the curl string into the log BEFORE
// the OnRequestLog callback redacts anything, so a curl line would carry the raw
// Authorization header and be replayable — no hook can scrub it. Re-enabling it
// must break a test, not just contradict a comment.
func TestRestWithDebugEmitsNoCurlCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	capture := &captureLogger{}
	client := (&service{endpoint: server.URL}).
		restWithDebug(context.Background(), 0, 0, time.Second).
		SetLogger(capture)

	if _, err := client.R().SetHeader("Authorization", "Bearer secret-bearer").Get("/auth/me"); err != nil {
		t.Fatalf("request: %v", err)
	}

	if strings.Contains(capture.all(), "REQUEST(CURL)") {
		t.Errorf("a curl command was generated — it bypasses redaction and is replayable:\n%s", capture.all())
	}
}
