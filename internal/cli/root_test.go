package cli

import (
	"bytes"
	"errors"
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

func TestScenarioExitErrorPrioritizesCommandFailure(t *testing.T) {
	err := scenarioExitError(scenario.RunResult{
		Failed:  true,
		Blocked: true,
		Drift:   true,
	})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.Code != 5 {
		t.Fatalf("expected command failure exit code 5, got %d", exitErr.Code)
	}
}

func TestContextCommands(t *testing.T) {
	root := seedCLIWorkspace(t)
	writeCLIContexts(t, root, map[string]model.Context{
		"zeta": {Repos: []string{"app-web"}},
		"api":  {Repos: []string{"shared-schema", "app-web"}},
	})

	out, err := executeCLI("--workspace", root, "context", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- api repos=2") || strings.Index(out, "- api") > strings.Index(out, "- zeta") {
		t.Fatalf("unexpected context list output:\n%s", out)
	}

	out, err = executeCLI("--workspace", root, "context", "show", "api")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "context: api") || !strings.Contains(out, "- shared-schema") {
		t.Fatalf("unexpected context show output:\n%s", out)
	}

	if _, err := executeCLI("--workspace", root, "context", "show", "missing"); err == nil {
		t.Fatal("expected unknown context error")
	}
}

func TestInfoStatusAndDoctorCommands(t *testing.T) {
	root := seedCLIWorkspace(t)
	checkout := initCLIGitRepo(t, filepath.Join(t.TempDir(), "app-web"))
	if _, err := workspace.SetBinding(root, "app-web", checkout); err != nil {
		t.Fatal(err)
	}
	writeCLIContexts(t, root, map[string]model.Context{
		"app-only": {Repos: []string{"app-web"}},
	})

	out, err := executeCLI("--workspace", root, "info")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "workspace: workspace") || !strings.Contains(out, "bindings: 1/2") || !strings.Contains(out, "wkit status") {
		t.Fatalf("unexpected info output:\n%s", out)
	}
	out, err = executeCLI("--workspace", root, "overview")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "workspace: workspace") {
		t.Fatalf("unexpected overview alias output:\n%s", out)
	}

	out, err = executeCLI("--workspace", root, "status", "--context", "app-only")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "app-web binding=ok git=ok") || strings.Contains(out, "shared-schema") {
		t.Fatalf("unexpected scoped status output:\n%s", out)
	}

	out, err = executeCLI("--workspace", root, "doctor")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "warning:") || !strings.Contains(out, "summary: errors=0") {
		t.Fatalf("expected warning-only doctor output, got:\n%s", out)
	}
}

func TestScenarioStatusCommandReturnsDriftExit(t *testing.T) {
	root := seedCLIWorkspace(t)
	checkout := initCLIGitRepo(t, filepath.Join(t.TempDir(), "app-web"))
	if _, err := workspace.SetBinding(root, "app-web", checkout); err != nil {
		t.Fatal(err)
	}
	writeCLIContexts(t, root, map[string]model.Context{
		"app-flow": {Repos: []string{"app-web"}},
	})
	changeID, err := workspace.CreateChange(root, "app-flow", "test rollout", "contract", time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.Pin(root, "app-flow", changeID, time.Date(2026, 4, 19, 12, 1, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	out, err := executeCLI("--workspace", root, "scenario", "status", "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "[ok] app-web") {
		t.Fatalf("unexpected clean scenario status output:\n%s", out)
	}

	if err := os.WriteFile(filepath.Join(checkout, "README.md"), []byte("# changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, checkout, "git", "add", ".")
	run(t, checkout, "git", "commit", "-m", "change")

	out, err = executeCLI("--workspace", root, "scenario", "status", "app-flow")
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 4 {
		t.Fatalf("expected scenario drift exit code 4, got err=%v output=\n%s", err, out)
	}
	if !strings.Contains(out, "[drift] app-web") {
		t.Fatalf("unexpected drift scenario status output:\n%s", out)
	}
}

func executeCLI(args ...string) (string, error) {
	cmd := newRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func seedCLIWorkspace(t *testing.T) string {
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

func writeCLIContexts(t *testing.T, root string, contexts map[string]model.Context) {
	t.Helper()
	if err := manifest.WriteYAML(filepath.Join(root, workspace.ContextsFile), model.ContextsDocument{
		Version:  1,
		Contexts: contexts,
	}); err != nil {
		t.Fatal(err)
	}
}

func initCLIGitRepo(t *testing.T, root string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# repo\n"), 0o644); err != nil {
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
