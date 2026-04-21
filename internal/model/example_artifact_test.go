package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestMinimalScenarioArtifactDecodes(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "minimal-workspace", "artifacts", "schema-rollout")

	var lock ScenarioLockDocument
	loadArtifactYAML(t, filepath.Join(root, "manifest.lock.yaml"), &lock)
	if lock.Scenario.ID != "schema-rollout" {
		t.Fatalf("unexpected scenario id: %q", lock.Scenario.ID)
	}
	if got := len(lock.Repos); got != 2 {
		t.Fatalf("expected 2 pinned repos, got %d", got)
	}
	if got := len(lock.Checks); got != 2 {
		t.Fatalf("expected 2 checks, got %d", got)
	}

	var report ScenarioReportDocument
	loadArtifactYAML(t, filepath.Join(root, "20260419T121000Z.yaml"), &report)
	assertArtifactPathExists(t, root, ptr("20260419T121000Z.md"))
	if report.Report.Scenario != lock.Scenario.ID {
		t.Fatalf("report scenario %q does not match lock scenario %q", report.Report.Scenario, lock.Scenario.ID)
	}
	if got := len(report.Results); got != len(lock.Checks) {
		t.Fatalf("expected %d report results, got %d", len(lock.Checks), got)
	}
	for _, result := range report.Results {
		assertArtifactPathExists(t, root, result.StdoutPath)
		assertArtifactPathExists(t, root, result.StderrPath)
	}
}

func TestFailureScenarioArtifactDecodes(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "failure-workspace", "artifacts", "schema-rollout")

	var lock ScenarioLockDocument
	loadArtifactYAML(t, filepath.Join(root, "manifest.lock.yaml"), &lock)
	if lock.Scenario.ID != "schema-rollout" {
		t.Fatalf("unexpected scenario id: %q", lock.Scenario.ID)
	}
	if got := len(lock.Repos); got != 2 {
		t.Fatalf("expected 2 pinned repos, got %d", got)
	}
	if got := len(lock.Checks); got != 2 {
		t.Fatalf("expected 2 checks, got %d", got)
	}

	var report ScenarioReportDocument
	loadArtifactYAML(t, filepath.Join(root, "20260419T122000Z.yaml"), &report)
	assertArtifactPathExists(t, root, ptr("20260419T122000Z.md"))
	if report.Report.Scenario != lock.Scenario.ID {
		t.Fatalf("report scenario %q does not match lock scenario %q", report.Report.Scenario, lock.Scenario.ID)
	}

	statuses := map[string]string{}
	for _, result := range report.Results {
		statuses[result.Check] = result.Status
		if result.StdoutPath != nil {
			assertArtifactPathExists(t, root, result.StdoutPath)
		}
		if result.StderrPath != nil {
			assertArtifactPathExists(t, root, result.StderrPath)
		}
	}
	if statuses["shared-schema:test"] != "blocked" {
		t.Fatalf("expected shared-schema:test to be blocked, got %q", statuses["shared-schema:test"])
	}
	if statuses["app-web:test"] != "failed" {
		t.Fatalf("expected app-web:test to fail, got %q", statuses["app-web:test"])
	}

	stderrPath := "logs/20260419T122000Z/app-web-test.stderr.txt"
	stderr, err := os.ReadFile(filepath.Join(root, stderrPath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stderr), "payload field customer_id is missing") {
		t.Fatalf("unexpected stderr log:\n%s", string(stderr))
	}
}

func loadArtifactYAML(t *testing.T, path string, out any) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		t.Fatalf("%s did not decode: %v", path, err)
	}
}

func assertArtifactPathExists(t *testing.T, root string, rel *string) {
	t.Helper()

	if rel == nil || *rel == "" {
		t.Fatal("artifact report path is empty")
	}
	path := filepath.Join(root, *rel)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("artifact report path %q is missing: %v", *rel, err)
	}
}

func ptr(value string) *string {
	return &value
}
