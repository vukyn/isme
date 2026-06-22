package models

import (
	"slices"
	"testing"
)

func TestValidateRedirectURLList(t *testing.T) {
	strSlice := func(s ...string) []string { return s }

	t.Run("valid list is returned cleaned", func(t *testing.T) {
		got, err := ValidateRedirectURLList([]string{
			"  https://a.example.com/callback ",
			"https://b.example.com/cb",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := strSlice("https://a.example.com/callback", "https://b.example.com/cb")
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty input yields non-nil empty slice", func(t *testing.T) {
		got, err := ValidateRedirectURLList(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("expected non-nil empty slice, got nil")
		}
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %v", got)
		}
	})

	t.Run("empty entry rejected", func(t *testing.T) {
		if _, err := ValidateRedirectURLList([]string{"https://a.example.com", "   "}); err == nil {
			t.Fatal("expected error for blank entry, got nil")
		}
	})

	t.Run("invalid url rejected", func(t *testing.T) {
		for _, bad := range []string{"not-a-url", "ftp-relative/path", "https://"} {
			if _, err := ValidateRedirectURLList([]string{bad}); err == nil {
				t.Errorf("expected error for %q, got nil", bad)
			}
		}
	})

	t.Run("duplicates deduped preserving order", func(t *testing.T) {
		got, err := ValidateRedirectURLList([]string{
			"https://a.example.com/cb",
			"https://b.example.com/cb",
			"https://a.example.com/cb",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := strSlice("https://a.example.com/cb", "https://b.example.com/cb")
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("over cap rejected", func(t *testing.T) {
		_, err := ValidateRedirectURLList([]string{
			"https://1.example.com/cb",
			"https://2.example.com/cb",
			"https://3.example.com/cb",
			"https://4.example.com/cb",
		})
		if err == nil {
			t.Fatal("expected error for 4 URLs (cap is 3), got nil")
		}
	})

	t.Run("exactly at cap accepted", func(t *testing.T) {
		got, err := ValidateRedirectURLList([]string{
			"https://1.example.com/cb",
			"https://2.example.com/cb",
			"https://3.example.com/cb",
		})
		if err != nil {
			t.Fatalf("unexpected error at cap: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("expected 3 URLs, got %d", len(got))
		}
	})
}

func TestRegisterRequestValidateRedirectURLs(t *testing.T) {
	base := RegisterRequest{
		AppCode:     "code",
		AppName:     "Name",
		RedirectURL: "https://primary.example.com/cb",
		CtxInfo:     "authen",
		Icon:        "box",
		Color:       "violet",
	}

	t.Run("absent redirect_urls is valid", func(t *testing.T) {
		req := base
		if err := req.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid redirect_urls accepted", func(t *testing.T) {
		req := base
		req.RedirectURLs = []string{"https://extra.example.com/cb"}
		if err := req.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid redirect_urls rejected", func(t *testing.T) {
		req := base
		req.RedirectURLs = []string{"not-a-url"}
		if err := req.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("over cap rejected", func(t *testing.T) {
		req := base
		req.RedirectURLs = []string{
			"https://1.example.com/cb",
			"https://2.example.com/cb",
			"https://3.example.com/cb",
			"https://4.example.com/cb",
		}
		if err := req.Validate(); err == nil {
			t.Fatal("expected error for over-cap list, got nil")
		}
	})
}

func TestUpdateAppearanceRequestValidateRedirectURLs(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	urlsPtr := func(u ...string) *[]string { s := u; return &s }

	t.Run("nil redirect_urls counts as unchanged (other field present)", func(t *testing.T) {
		req := UpdateAppearanceRequest{AppName: strPtr("New")}
		if err := req.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.RedirectURLs != nil {
			t.Error("expected RedirectURLs to remain nil")
		}
	})

	t.Run("redirect_urls alone satisfies at-least-one-field", func(t *testing.T) {
		req := UpdateAppearanceRequest{RedirectURLs: urlsPtr("https://a.example.com/cb")}
		if err := req.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty slice clears (valid)", func(t *testing.T) {
		req := UpdateAppearanceRequest{RedirectURLs: urlsPtr()}
		if err := req.Validate(); err != nil {
			t.Fatalf("unexpected error for clear, got %v", err)
		}
	})

	t.Run("invalid entry rejected", func(t *testing.T) {
		req := UpdateAppearanceRequest{RedirectURLs: urlsPtr("nope")}
		if err := req.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("over cap rejected", func(t *testing.T) {
		req := UpdateAppearanceRequest{RedirectURLs: urlsPtr(
			"https://1.example.com/cb",
			"https://2.example.com/cb",
			"https://3.example.com/cb",
			"https://4.example.com/cb",
		)}
		if err := req.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("no fields at all rejected", func(t *testing.T) {
		req := UpdateAppearanceRequest{}
		if err := req.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
