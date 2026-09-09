package services

import (
	"net/http"
	"testing"
)

// A 401/403 from isme is the ordinary outcome of an expired or missing token, so
// it must not be logged as a server-side error — that is what filled rainy's
// production log with what looked like a fault. Everything else must stay an
// error, so a real isme outage is not quietly downgraded to a warning.
func TestIsExpectedRejection(t *testing.T) {
	cases := map[int]bool{
		http.StatusUnauthorized:        true,
		http.StatusForbidden:           true,
		http.StatusInternalServerError: false,
		http.StatusBadGateway:          false,
		http.StatusServiceUnavailable:  false,
		http.StatusNotFound:            false,
		http.StatusTooManyRequests:     false,
		http.StatusBadRequest:          false,
	}
	for status, want := range cases {
		if got := isExpectedRejection(operationGetMe, status); got != want {
			t.Errorf("isExpectedRejection(%q, %d) = %v, want %v", operationGetMe, status, got, want)
		}
	}
}

// The refresh endpoint answers an expired, rotated or unknown refresh token with
// 400 and `invalid refresh token` rather than 401. That is a session ending
// normally, so it belongs at warn — rainy's production log carried it as
// `error` (`Failed refresh token from external auth (400)`) purely because the
// classification did not look at which operation was attempted.
func TestIsExpectedRejectionRefreshTokenBadRequest(t *testing.T) {
	if !isExpectedRejection(operationRefreshToken, http.StatusBadRequest) {
		t.Error("isExpectedRejection(refresh token, 400) = false, want true — an expired refresh token is a rejection, not a fault")
	}
}

// The 400 carve-out must stay scoped to refresh. On every other operation a 400
// means this client built a malformed request, which someone has to see.
func TestIsExpectedRejectionBadRequestStaysErrorElsewhere(t *testing.T) {
	others := []authOperation{
		operationGetMe,
		operationRequestLogin,
		operationExchangeCode,
		operationLogout,
	}
	for _, operation := range others {
		if isExpectedRejection(operation, http.StatusBadRequest) {
			t.Errorf("isExpectedRejection(%q, 400) = true, want false — only refresh treats a 400 as a rejection", operation)
		}
	}
}

// The credential rejections are operation-independent: every call gets the
// 401/403 treatment, including refresh.
func TestIsExpectedRejectionCredentialStatusesCoverEveryOperation(t *testing.T) {
	operations := []authOperation{
		operationGetMe,
		operationRequestLogin,
		operationRefreshToken,
		operationExchangeCode,
		operationLogout,
	}
	for _, operation := range operations {
		for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
			if !isExpectedRejection(operation, status) {
				t.Errorf("isExpectedRejection(%q, %d) = false, want true", operation, status)
			}
		}
		if isExpectedRejection(operation, http.StatusInternalServerError) {
			t.Errorf("isExpectedRejection(%q, 500) = true, want false — an isme outage must stay an error", operation)
		}
	}
}
