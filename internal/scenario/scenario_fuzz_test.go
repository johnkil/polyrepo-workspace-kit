package scenario

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzResolveCheckCWD(f *testing.F) {
	for _, seed := range []string{
		"",
		".",
		"sub",
		"sub/dir",
		"file.txt",
		"..",
		"../escape",
		"sub/../../escape",
		"/tmp",
		"linked-out",
		`sub\dir`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, cwd string) {
		if len(cwd) > 512 {
			return
		}

		checkout := filepath.Join(t.TempDir(), "checkout")
		if err := os.MkdirAll(filepath.Join(checkout, "sub", "dir"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(checkout, "file.txt"), []byte("not a directory\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		outside := t.TempDir()
		_ = os.Symlink(outside, filepath.Join(checkout, "linked-out"))

		dir, err := resolveCheckCWD(checkout, cwd)
		if err != nil {
			return
		}
		assertResolvedCWDWithinCheckout(t, checkout, dir)
	})
}

func assertResolvedCWDWithinCheckout(t *testing.T, checkout string, dir string) {
	t.Helper()

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("resolved cwd %q is not stat-able: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("resolved cwd %q is not a directory", dir)
	}

	realCheckout, err := filepath.EvalSymlinks(checkout)
	if err != nil {
		t.Fatalf("could not resolve checkout %q: %v", checkout, err)
	}
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("could not resolve cwd %q: %v", dir, err)
	}
	rel, err := filepath.Rel(realCheckout, realDir)
	if err != nil {
		t.Fatalf("could not compute relative cwd: %v", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		t.Fatalf("resolved cwd %q escapes checkout %q", dir, checkout)
	}
}
