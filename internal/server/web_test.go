package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/vukyn/isme/internal/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v3"
)

const webTestAPIBaseURL = "https://sso.example.test"

// newWebTestApp mounts exactly the web routes Start mounts, over a fake UI FS
// shaped like a real build: index.html with the config placeholder, a root
// favicon, and one hashed bundle file.
func newWebTestApp(t *testing.T) *fiber.App {
	t.Helper()
	uiFS := fstest.MapFS{
		"index.html":               {Data: []byte(`<!doctype html><html><head><script>window.__APP_CONFIG__ = { API_BASE_URL: "{{.APIBaseURL}}" };</script></head><body><div id="root"></div></body></html>`)},
		"favicon.svg":              {Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)},
		"assets/index-abc12345.js": {Data: []byte(`console.log("bundle");`)},
	}
	cfg := &config.Config{}
	cfg.Vite.BaseURL = webTestAPIBaseURL
	server := &Server{cfg: cfg}

	app := fiber.New(fiber.Config{
		Views: html.NewFileSystem(http.FS(uiFS), ".html"),
	})
	if err := mountWeb(app, uiFS, server.renderHomePage); err != nil {
		t.Fatalf("mountWeb: %v", err)
	}
	return app
}

func webTestGet(t *testing.T, app *fiber.App, path string, headers map[string]string) (*http.Response, string) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read GET %s body: %v", path, err)
	}
	_ = response.Body.Close()
	return response, string(body)
}

func TestMountWeb(t *testing.T) {
	app := newWebTestApp(t)

	t.Run("asset is served without a zero Last-Modified", func(t *testing.T) {
		response, body := webTestGet(t, app, "/assets/index-abc12345.js", nil)
		if response.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", response.StatusCode)
		}
		if contentType := response.Header.Get(fiber.HeaderContentType); !strings.Contains(contentType, "javascript") {
			t.Errorf("Content-Type = %q, want javascript", contentType)
		}
		if body != `console.log("bundle");` {
			t.Errorf("body = %q, want the bundle", body)
		}
		if lastModified := response.Header.Get(fiber.HeaderLastModified); lastModified != "" {
			t.Errorf("Last-Modified = %q, want none (the embed FS has a zero modtime)", lastModified)
		}
	})

	t.Run("asset ignores If-Modified-Since", func(t *testing.T) {
		response, _ := webTestGet(t, app, "/assets/index-abc12345.js", map[string]string{
			fiber.HeaderIfModifiedSince: "Wed, 21 Oct 2099 07:28:00 GMT",
		})
		if response.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200 (a 304 here would pin a stale file)", response.StatusCode)
		}
	})

	t.Run("missing asset is a plain 404, not the SPA", func(t *testing.T) {
		response, body := webTestGet(t, app, "/assets/nope-00000000.js", nil)
		if response.StatusCode != fiber.StatusNotFound {
			t.Fatalf("status = %d, want 404", response.StatusCode)
		}
		if strings.Contains(body, "__APP_CONFIG__") {
			t.Errorf("missing asset answered with index.html: %q", body)
		}
	})

	t.Run("favicon is the svg, not the SPA", func(t *testing.T) {
		response, body := webTestGet(t, app, "/favicon.svg", nil)
		if response.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", response.StatusCode)
		}
		if contentType := response.Header.Get(fiber.HeaderContentType); !strings.HasPrefix(contentType, "image/svg+xml") {
			t.Errorf("Content-Type = %q, want image/svg+xml", contentType)
		}
		if !strings.HasPrefix(body, "<svg") {
			t.Errorf("body = %q, want the svg", body)
		}
		if lastModified := response.Header.Get(fiber.HeaderLastModified); lastModified != "" {
			t.Errorf("Last-Modified = %q, want none", lastModified)
		}
	})

	for _, path := range []string{"/", "/sso/login", "/accept-invite?token=x", "/users/invite"} {
		t.Run("SPA renders "+path, func(t *testing.T) {
			response, body := webTestGet(t, app, path, nil)
			if response.StatusCode != fiber.StatusOK {
				t.Fatalf("status = %d, want 200", response.StatusCode)
			}
			if contentType := response.Header.Get(fiber.HeaderContentType); !strings.HasPrefix(contentType, "text/html") {
				t.Errorf("Content-Type = %q, want text/html", contentType)
			}
			if strings.Contains(body, "{{.APIBaseURL}}") {
				t.Errorf("template placeholder was not rendered: %q", body)
			}
			// html/template JS-escapes "/" inside a <script>; the browser reads
			// the escaped form back as the same string.
			if !strings.Contains(strings.ReplaceAll(body, `\/`, "/"), webTestAPIBaseURL) {
				t.Errorf("rendered page lacks the API base URL %q: %q", webTestAPIBaseURL, body)
			}
		})
	}
}
