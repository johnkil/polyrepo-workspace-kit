package demo_test

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/demo"
)

func TestRunMinimalDemo(t *testing.T) {
	skipWindows(t)
	result, err := demo.Run(demo.KindMinimal, time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed || result.Blocked || result.Drift {
		t.Fatalf("unexpected minimal demo result: %#v", result)
	}
	if result.ChangeID != "CHG-2026-04-19-001" {
		t.Fatalf("unexpected change id: %s", result.ChangeID)
	}
	assertExists(t, result.ReportPath)
	assertExists(t, result.TextReportPath)
	assertExists(t, result.MarkdownReportPath)
	if !strings.Contains(result.MarkdownReport, "Results: passed=2 failed=0 blocked=0 skipped=0") {
		t.Fatalf("unexpected markdown report:\n%s", result.MarkdownReport)
	}
	if !strings.Contains(result.HandoffMarkdown, "# Handoff: payload field rollout") {
		t.Fatalf("unexpected handoff markdown:\n%s", result.HandoffMarkdown)
	}
}

func TestRunFailureDemo(t *testing.T) {
	skipWindows(t)
	result, err := demo.Run(demo.KindFailure, time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Failed || !result.Blocked || !result.Drift {
		t.Fatalf("expected failure demo to produce drift and failure, got %#v", result)
	}
	if !strings.Contains(result.MarkdownReport, "Results: passed=0 failed=1 blocked=1 skipped=0") ||
		!strings.Contains(result.MarkdownReport, "payload field customer_id is missing") {
		t.Fatalf("unexpected markdown report:\n%s", result.MarkdownReport)
	}
}

func skipWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("wkit demo requires POSIX sh in v0.x")
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
