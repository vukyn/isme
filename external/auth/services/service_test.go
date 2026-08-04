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
		if got := isExpectedRejection(status); got != want {
			t.Errorf("isExpectedRejection(%d) = %v, want %v", status, got, want)
		}
	}
}
