package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzValidateReportPath(f *testing.F) {
	for _, seed := range []string{
		"logs/run/stdout.txt",
		"stdout.txt",
		"",
		".",
		"..",
		"../escape.txt",
		"logs/../../escape.txt",
		"/tmp/escape.txt",
		`logs\stdout.txt`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 512 {
			return
		}

		root := t.TempDir()
		reportDir := filepath.Join(root, "local", "reports", "app-flow")
		if err := os.MkdirAll(reportDir, 0o755); err != nil {
			t.Fatal(err)
		}

		report := &Report{}
		validateReportPath(root, reportDir, "local/reports/app-flow/run.yaml", "app-web:test", "stdout_path", &value, report)
		if len(report.Errors) == 0 && value != "" {
			assertAcceptedReportPathWithinRoot(t, root, reportDir, value)
		}
	})
}

func assertAcceptedReportPathWithinRoot(t *testing.T, root string, reportDir string, value string) {
	t.Helper()

	if filepath.IsAbs(value) {
		t.Fatalf("absolute report path %q was accepted", value)
	}
	clean := filepath.Clean(value)
	if clean == "." || clean == ".." || clean == "" || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		t.Fatalf("unsafe report path %q was accepted", value)
	}
	target := filepath.Join(reportDir, clean)
	relToRoot, err := filepath.Rel(root, target)
	if err == nil && (relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator))) {
		t.Fatalf("report path %q resolves outside workspace root %q", value, root)
	}
}
