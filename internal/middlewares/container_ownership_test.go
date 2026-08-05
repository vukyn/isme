package middlewares

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// ownershipScanRoot is the tree this invariant scans: everything under internal/,
// reached relative to this package's directory.
const ownershipScanRoot = ".."

// ownershipMinScannedFiles guards against a path typo making the walk vacuous —
// the one way this test could pass while proving nothing. internal/ holds ~100
// non-test .go files, so this floor has room to shrink without going quiet.
const ownershipMinScannedFiles = 80

// ownershipAllowedFiles are the files permitted to release a di container: whoever
// CREATES a sub-container owns it.
//
//   - middlewares/di_container_middleware.go — DiContainerMiddleware creates the
//     request sub-container and releases it with its own defer.
//
// isme has no startup one-off that opens its own sub-container (gardener's
// bootstrapOwner is the platform's example of a legitimate second entry). Note
// di/di_middleware.go DOES retain a sub-container deliberately — but it never
// releases it, so it is not a violation here; see its own comment for why.
var ownershipAllowedFiles = map[string]bool{
	filepath.Join("middlewares", "di_container_middleware.go"): true,
}

// forbiddenReleaseCalls are the ways a borrower could release a container it does
// not own. `ctn` is the name every isme handler binds the request container to.
var forbiddenReleaseCalls = []string{
	"ctn.Delete()",
	"ctn.DeleteWithSubContainers()",
}

// TestNoHandlerReleasesRequestContainer is the exhaustiveness guard for the sweep
// that removed 55 `defer ctn.Delete()` calls from the handlers.
//
// It exists because the runtime Close counter in
// TestDiContainerMiddlewareReleasesRequestContainer cannot see this: that harness
// mounts stub routes, so a leftover defer in a REAL handler never runs inside it and
// the count still reads one. Nothing else catches one either — the compiler is
// happy, and sarulabs/di's second Delete returns nil, so the only symptom is every
// registered Close running twice, and every request-scope Close here is a debug log
// line. A static invariant is the only thing that can prove the sweep was complete
// and, more importantly, that it stays complete.
//
// This was not hypothetical: while the same guard was being written for medioa2, a
// stray `git checkout -- <file>` restored one removed defer from HEAD mid-sweep, and
// this is the check that caught it.
//
// The rule it encodes: whoever CREATES a sub-container releases it. Handlers only
// borrow the one DiContainerMiddleware put in the Fiber locals.
func TestNoHandlerReleasesRequestContainer(t *testing.T) {
	root, err := filepath.Abs(ownershipScanRoot)
	if err != nil {
		t.Fatalf("resolve scan root: %v", err)
	}

	scanned := 0
	violations := make([]string, 0)

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if ownershipAllowedFiles[relative] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		scanned++

		for number, line := range strings.Split(string(content), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				// A comment may legitimately quote the forbidden call while
				// explaining the rule — several do.
				continue
			}
			for _, call := range forbiddenReleaseCalls {
				if strings.Contains(trimmed, call) {
					violations = append(violations, relative+":"+strconv.Itoa(number+1)+": "+trimmed)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}

	if scanned < ownershipMinScannedFiles {
		t.Fatalf("scanned only %d .go files under %s, want at least %d — the invariant is not actually looking at the codebase", scanned, root, ownershipMinScannedFiles)
	}

	if len(violations) > 0 {
		t.Fatalf("%d handler(s) release a di container they do not own:\n\t%s\n\nDiContainerMiddleware creates the request sub-container and releases it on every path; a handler only borrows it. A second Delete returns nil but re-runs every registered Close, so this is invisible at runtime — remove the call.",
			len(violations), strings.Join(violations, "\n\t"))
	}
}
