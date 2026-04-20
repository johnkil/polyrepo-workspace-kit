package orient

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

func TestListAndGetContexts(t *testing.T) {
	root := seedWorkspace(t)
	writeContexts(t, root, map[string]model.Context{
		"zeta": {Repos: []string{"app-web"}},
		"api":  {Repos: []string{"shared-schema", "app-web"}},
	})

	contexts, err := ListContexts(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{contexts[0].ID, contexts[1].ID}; got[0] != "api" || got[1] != "zeta" {
		t.Fatalf("contexts not sorted: %#v", got)
	}
	if contexts[0].RepoCount != 2 {
		t.Fatalf("unexpected repo count: %#v", contexts[0])
	}

	context, err := GetContext(root, "api")
	if err != nil {
		t.Fatal(err)
	}
	if len(context.Repos) != 2 || context.Repos[0] != "shared-schema" {
		t.Fatalf("unexpected context: %#v", context)
	}
	if _, err := GetContext(root, "missing"); err == nil {
		t.Fatal("expected unknown context error")
	}
}

func TestWorkspaceInfoSummarizesWorkspace(t *testing.T) {
	root := seedWorkspace(t)
	if _, err := workspace.SetBinding(root, "app-web", t.TempDir()); err != nil {
		t.Fatal(err)
	}
	writeContexts(t, root, map[string]model.Context{
		"schema-rollout": {Repos: []string{"shared-schema", "app-web"}},
	})
	if _, err := workspace.CreateChange(root, "schema-rollout", "payload rollout", "contract", time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(filepath.Join(root, "coordination", "scenarios", "schema-rollout", "manifest.lock.yaml"), model.ScenarioLockDocument{
		Version: 1,
		Scenario: model.ScenarioMeta{
			ID:      "schema-rollout",
			Change:  "CHG-2026-04-19-001",
			Context: "schema-rollout",
		},
	}); err != nil {
		t.Fatal(err)
	}

	info, err := WorkspaceInfo(root)
	if err != nil {
		t.Fatal(err)
	}
	if info.RepoCount != 2 || info.BoundRepos != 1 || info.TotalRepos != 2 {
		t.Fatalf("unexpected repo/binding counts: %#v", info)
	}
	if info.ChangeCount != 1 || info.LatestChange != "CHG-2026-04-19-001" {
		t.Fatalf("unexpected change summary: %#v", info)
	}
	if info.ScenarioCount != 1 || info.LatestScenario != "schema-rollout" {
		t.Fatalf("unexpected scenario summary: %#v", info)
	}
	if len(info.RepoKinds) != 2 {
		t.Fatalf("expected repo kind counts, got %#v", info.RepoKinds)
	}
}

func TestWorkspaceInfoSurfacesRepoManifestLoadFailure(t *testing.T) {
	root := seedWorkspace(t)
	if err := os.Remove(filepath.Join(root, "repos", "shared-schema", "repo.yaml")); err != nil {
		t.Fatal(err)
	}

	_, err := WorkspaceInfo(root)
	if err == nil || !strings.Contains(err.Error(), `load repo "shared-schema"`) {
		t.Fatalf("expected repo manifest load error, got %v", err)
	}
}

func TestWorkspaceStatusReportsBindingsAndGitState(t *testing.T) {
	root := seedWorkspace(t)
	checkout := initGitRepo(t, filepath.Join(t.TempDir(), "app-web"))
	if _, err := workspace.SetBinding(root, "app-web", checkout); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(checkout, "README.md"), []byte("# changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(checkout, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := WorkspaceStatus(root, StatusOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Repos) != 2 {
		t.Fatalf("expected two repo statuses, got %#v", report.Repos)
	}
	app := findRepoStatus(t, report, "app-web")
	if app.BindingStatus != "ok" || app.GitStatus != "ok" {
		t.Fatalf("unexpected app status: %#v", app)
	}
	if app.DirtyFiles != 1 || app.UntrackedFiles != 1 || app.Upstream != "none" {
		t.Fatalf("unexpected app git summary: %#v", app)
	}
	shared := findRepoStatus(t, report, "shared-schema")
	if shared.BindingStatus != "missing" {
		t.Fatalf("expected missing binding, got %#v", shared)
	}
}

func TestWorkspaceStatusCanScopeToContext(t *testing.T) {
	root := seedWorkspace(t)
	writeContexts(t, root, map[string]model.Context{
		"app-only": {Repos: []string{"app-web"}},
	})
	report, err := WorkspaceStatus(root, StatusOptions{ContextID: "app-only"})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Repos) != 1 || report.Repos[0].RepoID != "app-web" {
		t.Fatalf("unexpected scoped status: %#v", report.Repos)
	}
}

func TestScenarioStatusReportsDriftAndBlocked(t *testing.T) {
	root := seedWorkspace(t)
	checkout := initGitRepo(t, filepath.Join(t.TempDir(), "app-web"))
	if _, err := workspace.SetBinding(root, "app-web", checkout); err != nil {
		t.Fatal(err)
	}
	writeContexts(t, root, map[string]model.Context{
		"app-flow": {Repos: []string{"app-web"}},
	})
	changeID, err := workspace.CreateChange(root, "app-flow", "test rollout", "contract", time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.Pin(root, "app-flow", changeID, time.Date(2026, 4, 19, 12, 1, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	clean, err := ScenarioStatus(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if clean.Drift || clean.Blocked || clean.Repos[0].ScenarioStatus != "ok" {
		t.Fatalf("expected clean scenario status, got %#v", clean)
	}

	if err := os.WriteFile(filepath.Join(checkout, "README.md"), []byte("# changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, checkout, "git", "add", ".")
	run(t, checkout, "git", "commit", "-m", "change")

	drift, err := ScenarioStatus(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if !drift.Drift || drift.Repos[0].ScenarioStatus != "drift" {
		t.Fatalf("expected drift scenario status, got %#v", drift)
	}

	bindings, err := workspace.LoadBindings(root)
	if err != nil {
		t.Fatal(err)
	}
	bindings.Bindings["app-web"] = model.Binding{Path: filepath.Join(t.TempDir(), "missing")}
	if err := workspace.SaveBindings(root, bindings); err != nil {
		t.Fatal(err)
	}
	blocked, err := ScenarioStatus(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if !blocked.Blocked || blocked.Repos[0].ScenarioStatus != "blocked" {
		t.Fatalf("expected blocked scenario status, got %#v", blocked)
	}
}

func TestDoctorCombinesManifestAndLocalDiagnostics(t *testing.T) {
	missing := Doctor(t.TempDir())
	if len(missing.Errors) == 0 {
		t.Fatal("expected missing workspace error")
	}

	root := seedWorkspace(t)
	report := Doctor(root)
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors for missing binding warning, got %#v", report.Errors)
	}
	if !contains(report.Warnings, "missing local binding") {
		t.Fatalf("expected missing binding warning, got %#v", report.Warnings)
	}

	checkout := initGitRepo(t, filepath.Join(t.TempDir(), "app-web"))
	if _, err := workspace.SetBinding(root, "app-web", checkout); err != nil {
		t.Fatal(err)
	}
	repo, err := workspace.LoadRepo(root, "app-web")
	if err != nil {
		t.Fatal(err)
	}
	entrypoint := repo.Entrypoints["test"]
	entrypoint.CWD = "missing-dir"
	repo.Entrypoints["test"] = entrypoint
	if err := manifest.WriteYAML(filepath.Join(root, "repos", "app-web", "repo.yaml"), repo); err != nil {
		t.Fatal(err)
	}
	withInvalidCWD := Doctor(root)
	if !contains(withInvalidCWD.Errors, "entrypoint test cwd") {
		t.Fatalf("expected invalid cwd error, got %#v", withInvalidCWD.Errors)
	}
}

func seedWorkspace(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "shared-schema", "contract"); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeContexts(t *testing.T, root string, contexts map[string]model.Context) {
	t.Helper()
	if err := manifest.WriteYAML(filepath.Join(root, workspace.ContextsFile), model.ContextsDocument{
		Version:  1,
		Contexts: contexts,
	}); err != nil {
		t.Fatal(err)
	}
}

func findRepoStatus(t *testing.T, report StatusReport, repoID string) RepoStatus {
	t.Helper()
	for _, status := range report.Repos {
		if status.RepoID == repoID {
			return status
		}
	}
	t.Fatalf("missing repo status for %s in %#v", repoID, report.Repos)
	return RepoStatus{}
}

func contains(items []string, needle string) bool {
	for _, item := range items {
		if strings.Contains(item, needle) {
			return true
		}
	}
	return false
}

func initGitRepo(t *testing.T, root string) string {
	t.Helper()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# repo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bin", "test"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	run(t, root, "git", "init")
	run(t, root, "git", "config", "user.email", "test@example.com")
	run(t, root, "git", "config", "user.name", "Test User")
	run(t, root, "git", "add", ".")
	run(t, root, "git", "commit", "-m", "init")
	return root
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, string(out))
	}
}
