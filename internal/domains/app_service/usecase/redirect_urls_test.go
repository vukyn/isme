package usecase

import (
	"slices"
	"testing"
)

func TestMarshalUnmarshalRedirectURLs(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		urls := []string{"https://a.example.com/cb", "https://b.example.com/cb"}
		raw := marshalRedirectURLs(urls)
		got := unmarshalRedirectURLs(raw)
		if !slices.Equal(got, urls) {
			t.Errorf("round trip = %v, want %v", got, urls)
		}
	})

	t.Run("nil marshals to empty array", func(t *testing.T) {
		if got := marshalRedirectURLs(nil); got != "[]" {
			t.Errorf("marshalRedirectURLs(nil) = %q, want %q", got, "[]")
		}
	})

	t.Run("empty marshals to empty array", func(t *testing.T) {
		if got := marshalRedirectURLs([]string{}); got != "[]" {
			t.Errorf("marshalRedirectURLs([]) = %q, want %q", got, "[]")
		}
	})
}

func TestUnmarshalRedirectURLsTolerant(t *testing.T) {
	cases := map[string]string{
		"empty string": "",
		"garbage":      "not json",
		"json null":    "null",
		"empty array":  "[]",
		"object":       `{"a":1}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			got := unmarshalRedirectURLs(raw)
			if got == nil {
				t.Fatal("expected non-nil slice (JSON-null-slice gotcha), got nil")
			}
			if len(got) != 0 {
				t.Errorf("expected empty slice for %q, got %v", raw, got)
			}
		})
	}
}
