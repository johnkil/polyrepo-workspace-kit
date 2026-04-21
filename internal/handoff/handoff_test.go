package handoff_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/handoff"
	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

func TestMarkdownIncludesChangeScenarioAndLatestReport(t *testing.T) {
	root, checkout, changeID := seedHandoffWorkspace(t)
	if _, err := scenario.Pin(root, "app-flow", changeID, time.Date(2026, 4, 19, 12, 5, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if checkout == "" {
		t.Fatal("checkout path should not be empty")
	}

	out, err := handoff.Markdown(root, changeID, handoff.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# Handoff: payload rollout",
		"- change: `" + changeID + "`",
		"- scenario: `app-flow`",
		"latest report: `local/reports/app-flow/",
		"latest markdown report: `local/reports/app-flow/",
		"| `app-web` |",
		"Results: passed=1 failed=0 blocked=0 skipped=0",
		"| `app-web:test` | `passed` |",
		"derived handoff artifact",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in handoff output:\n%s", want, out)
		}
	}
}

func TestMarkdownWithoutScenarioStillSummarizesChange(t *testing.T) {
	root, _, changeID := seedHandoffWorkspace(t)
	out, err := handoff.Markdown(root, changeID, handoff.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"- scenario: none found for this change",
		"No scenario lock was found for this change yet.",
		"wkit scenario pin <scenario-id> --change <change-id>",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in handoff output:\n%s", want, out)
		}
	}
}

func seedHandoffWorkspace(t *testing.T) (string, string, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace")
	checkout := filepath.Join(t.TempDir(), "app-web")
	initGitRepo(t, checkout)

	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}
	repo, err := workspace.LoadRepo(root, "app-web")
	if err != nil {
		t.Fatal(err)
	}
	entrypoint := repo.Entrypoints["test"]
	entrypoint.Run = "go run ./testcmd.go"
	entrypoint.TimeoutSeconds = 30
	entrypoint.EnvProfile = "default"
	repo.Entrypoints["test"] = entrypoint
	if err := manifest.WriteYAML(filepath.Join(root, "repos", "app-web", "repo.yaml"), repo); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "app-web", checkout); err != nil {
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
	changeID, err := workspace.CreateChange(root, "app-flow", "payload rollout", "contract", time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return root, checkout, changeID
}

func initGitRepo(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# app-web\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := `package main

import "fmt"

func main() {
	fmt.Println("ok")
}
`
	if err := os.WriteFile(filepath.Join(root, "testcmd.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, root, "git", "init")
	run(t, root, "git", "config", "user.email", "test@example.com")
	run(t, root, "git", "config", "user.name", "Test User")
	run(t, root, "git", "add", ".")
	run(t, root, "git", "commit", "-m", "init")
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
