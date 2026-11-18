package middlewares

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// fiberV3GuardScanRoot is the module root, so cmd/, db/ and external/ are held to
// the same rule as internal/.
const fiberV3GuardScanRoot = "../.."

// fiberV2OnlyImports are import paths that belong to the Fiber v2 line. kuery
// ships each of its Fiber helpers twice (a v2 package and a v3 twin), and the
// twins are not interchangeable at runtime:
//   - kuery/ctx and kuery/ctxv3 each declare their own ContextKey type, so a
//     value stored under one is invisible to the other. A file that only calls
//     pkgCtx.GetUserID(ctx) compiles against either and, on the wrong one,
//     silently reads an empty user id (empty created_by/updated_by, broken
//     ownership checks) with no error anywhere.
//   - the rest are written against *fiber.Ctx from fiber/v2, and importing any
//     of them drags fiber/v2 back into the build.
//
// Matched against the exact import path, so kuery/ctxv3 never matches kuery/ctx.
var fiberV2OnlyImports = []string{
	"github.com/gofiber/fiber/v2",
	"github.com/gofiber/contrib/fiberzerolog",
	"github.com/gofiber/template/html/v2",
	"github.com/vukyn/kuery/ctx",
	"github.com/vukyn/kuery/http/fiber",
	"github.com/vukyn/kuery/recover",
	"github.com/vukyn/kuery/rbac",
	"github.com/vukyn/kuery/auth",
	"github.com/vukyn/kuery/apikey",
	"github.com/vukyn/kuery/http/ratelimit",
}

// fiberV3GuardSkipDirs are trees that hold no Go source of this module.
var fiberV3GuardSkipDirs = map[string]bool{
	".git":               true,
	"node_modules":       true,
	"ui":                 true,
	".code-review-graph": true,
}

func isFiberV2OnlyImport(path string) bool {
	for _, banned := range fiberV2OnlyImports {
		if path == banned || strings.HasPrefix(path, banned+"/") {
			return true
		}
	}
	return false
}

func walkModuleGoFiles(t *testing.T, visit func(path string, source []byte)) {
	t.Helper()
	root, err := filepath.Abs(fiberV3GuardScanRoot)
	if err != nil {
		t.Fatalf("resolve scan root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("scan root %s is not the module root: %v", root, err)
	}

	scanned := 0
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && fiberV3GuardSkipDirs[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		scanned++
		visit(path, source)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	if scanned < ownershipMinScannedFiles {
		t.Fatalf("scanned only %d .go files under %s, want at least %d — the guard is not actually looking at the codebase", scanned, root, ownershipMinScannedFiles)
	}
}

// TestNoFiberV2Imports fails the build when any file (tests included) imports
// a Fiber v2 package or a kuery v2 helper instead of its v3 twin.
func TestNoFiberV2Imports(t *testing.T) {
	walkModuleGoFiles(t, func(path string, source []byte) {
		file, err := parser.ParseFile(token.NewFileSet(), path, source, parser.ImportsOnly)
		if err != nil {
			t.Errorf("parse %s: %v", path, err)
			return
		}
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Errorf("%s: unquote import %s: %v", path, spec.Path.Value, err)
				continue
			}
			if isFiberV2OnlyImport(importPath) {
				t.Errorf("%s imports %q, a Fiber v2 package — use its v3 twin (kuery/ctxv3, kuery/http/fiberv3, kuery/recoverv3, kuery/rbacv3, gofiber/fiber/v3)", path, importPath)
			}
		}
	})
}

// TestBindErrorsAreInvalidRequest fails when a c.Bind().Body/Query error is
// handed to pkgHttp.Err unwrapped. A raw bind error maps to a 500, but a
// malformed request is the client's fault and must be a 400.
func TestBindErrorsAreInvalidRequest(t *testing.T) {
	walkModuleGoFiles(t, func(path string, source []byte) {
		lines := strings.Split(string(source), "\n")
		for index, line := range lines {
			if !strings.Contains(line, "Bind().Body(") && !strings.Contains(line, "Bind().Query(") {
				continue
			}
			for next := index + 1; next < len(lines); next++ {
				trimmed := strings.TrimSpace(lines[next])
				if trimmed == "" {
					continue
				}
				if trimmed == "return pkgHttp.Err(c, err)" {
					t.Errorf("%s:%d returns a raw bind error; wrap it as pkgErr.InvalidRequest(err.Error())", path, next+1)
				}
				break
			}
		}
	})
}

func TestFiberV2GuardMatchesExactPaths(t *testing.T) {
	cases := map[string]bool{
		"github.com/vukyn/kuery/ctx":                  true,
		"github.com/vukyn/kuery/ctxv3":                false,
		"github.com/vukyn/kuery/http/fiber":           true,
		"github.com/vukyn/kuery/http/fiberv3":         false,
		"github.com/vukyn/kuery/rbacv3":               false,
		"github.com/gofiber/fiber/v2/middleware/cors": true,
		"github.com/gofiber/fiber/v3/middleware/cors": false,
		"github.com/gofiber/template/html/v3":         false,
	}
	for importPath, want := range cases {
		if got := isFiberV2OnlyImport(importPath); got != want {
			t.Errorf("isFiberV2OnlyImport(%q) = %v, want %v", importPath, got, want)
		}
	}
}
