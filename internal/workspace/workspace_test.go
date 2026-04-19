package workspace_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

func TestInitRegisterAndBind(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "coordination", "workspace.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "local", "reports")); err != nil {
		t.Fatal(err)
	}

	manifestPath, err := workspace.RegisterRepo(root, "app-web", "app")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatal(err)
	}

	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(doc.Repos); got != 1 {
		t.Fatalf("expected 1 repo, got %d", got)
	}

	checkout := t.TempDir()
	boundPath, err := workspace.SetBinding(root, "app-web", checkout)
	if err != nil {
		t.Fatal(err)
	}
	if boundPath != checkout {
		t.Fatalf("expected %q, got %q", checkout, boundPath)
	}

	bindings, err := workspace.LoadBindings(root)
	if err != nil {
		t.Fatal(err)
	}
	if bindings.Bindings["app-web"].Path != checkout {
		t.Fatalf("binding not saved")
	}
}

func TestSetBindingRejectsMissingPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}

	if _, err := workspace.SetBinding(root, "app-web", filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected missing binding path to be rejected")
	}
}

func TestSetBindingRejectsFilePath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(filePath, []byte("# app\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := workspace.SetBinding(root, "app-web", filePath); err == nil {
		t.Fatal("expected file binding path to be rejected")
	}
}

func TestCreateChangeFromContext(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "shared-schema", "contract"); err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(filepath.Join(root, workspace.ContextsFile), model.ContextsDocument{
		Version: 1,
		Contexts: map[string]model.Context{
			"schema-rollout": {Repos: []string{"shared-schema", "app-web"}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	changeID, err := workspace.CreateChange(root, "schema-rollout", "payload rollout", "contract", time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if changeID != "CHG-2026-04-19-001" {
		t.Fatalf("unexpected change id: %s", changeID)
	}
	change, err := workspace.LoadChange(root, changeID)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(change.Change.Repos); got != 2 {
		t.Fatalf("expected 2 repos, got %d", got)
	}
}

func TestRejectsIDsWithPathTraversal(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}

	if _, err := workspace.RegisterRepo(root, "../../escape", "app"); err == nil {
		t.Fatal("expected repo id with path traversal to be rejected")
	}
	if _, err := workspace.ChangePath(root, "../CHG-2026-04-19-001"); err == nil {
		t.Fatal("expected change id with path traversal to be rejected")
	}
	if _, err := workspace.ScenarioPath(root, "../app-flow"); err == nil {
		t.Fatal("expected scenario id with path traversal to be rejected")
	}
}

func TestCreateChangeUsesMaxSuffixWithoutOverwriting(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(filepath.Join(root, workspace.ContextsFile), model.ContextsDocument{
		Version: 1,
		Contexts: map[string]model.Context{
			"app-flow": {Repos: []string{"app-web"}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	firstID, err := workspace.CreateChange(root, "app-flow", "first", "contract", time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if firstID != "CHG-2026-04-19-001" {
		t.Fatalf("unexpected first id: %s", firstID)
	}
	thirdPath, err := workspace.ChangePath(root, "CHG-2026-04-19-003")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(thirdPath, model.ChangeDocument{
		Version: 1,
		Change: model.Change{
			ID:      "CHG-2026-04-19-003",
			Title:   "do not overwrite",
			Kind:    "contract",
			Context: "app-flow",
			Repos:   []string{"app-web"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	nextID, err := workspace.CreateChange(root, "app-flow", "next", "contract", time.Date(2026, 4, 19, 12, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if nextID != "CHG-2026-04-19-004" {
		t.Fatalf("expected max suffix + 1, got %s", nextID)
	}
	data, err := os.ReadFile(thirdPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "do not overwrite") {
		t.Fatalf("existing change was overwritten:\n%s", string(data))
	}
}
