package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzValidateIDAndDerivedPaths(f *testing.F) {
	for _, seed := range []string{
		"app-web",
		"shared_schema",
		"team.api",
		"CHG-2026-04-19-001",
		"",
		" app-web",
		"app-web ",
		"../escape",
		"..",
		".hidden",
		"app..web",
		"app/web",
		`app\web`,
		"/abs",
	} {
		f.Add(seed)
	}

	pathChecks := []struct {
		name string
		fn   func(string, string) (string, error)
	}{
		{name: "repo", fn: RepoManifestPath},
		{name: "change", fn: ChangePath},
		{name: "scenario", fn: ScenarioPath},
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 512 {
			return
		}

		root := filepath.Join(os.TempDir(), "wkit-fuzz-root")
		idErr := ValidateID("id", value)
		for _, check := range pathChecks {
			path, pathErr := check.fn(root, value)
			if idErr != nil {
				if pathErr == nil {
					t.Fatalf("%s path accepted invalid id %q as %q", check.name, value, path)
				}
				continue
			}
			if pathErr != nil {
				t.Fatalf("%s path rejected valid id %q: %v", check.name, value, pathErr)
			}
			assertPathWithinRoot(t, root, path)
		}
	})
}

func assertPathWithinRoot(t *testing.T, root string, path string) {
	t.Helper()

	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("could not compute relative path from %q to %q: %v", root, path, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		t.Fatalf("path %q escapes root %q", path, root)
	}
}
