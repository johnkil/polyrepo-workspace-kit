package model

import (
	"os"
	"path/filepath"
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
