package scenario_test

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

func TestPinAndRunScenario(t *testing.T) {
	root, _ := seedPinnedScenario(t)

	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed || result.Blocked || result.Drift {
		t.Fatalf("unexpected scenario result: %#v", result)
	}
	if len(result.Outcomes) != 1 || result.Outcomes[0].Status != "passed" {
		t.Fatalf("unexpected outcomes: %#v", result.Outcomes)
	}
	if _, err := os.Stat(result.ReportPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(result.TextReportPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(result.MarkdownReportPath); err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(result.TextReportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(text), "Scenario: app-flow") || !strings.Contains(string(text), "passed=1") {
		t.Fatalf("unexpected text report:\n%s", string(text))
	}
	markdown, err := os.ReadFile(result.MarkdownReportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(markdown), "# Scenario Report: app-flow") || !strings.Contains(string(markdown), "| `app-web:test` | `passed` |") {
		t.Fatalf("unexpected markdown report:\n%s", string(markdown))
	}
}

func TestScenarioRunAvoidsSameSecondReportCollision(t *testing.T) {
	root, _ := seedPinnedScenario(t)
	now := time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC)

	first, err := scenario.Run(root, "app-flow", now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := scenario.Run(root, "app-flow", now)
	if err != nil {
		t.Fatal(err)
	}

	if first.ReportPath == second.ReportPath {
		t.Fatalf("expected distinct report paths, got %s", first.ReportPath)
	}
	if first.TextReportPath == second.TextReportPath {
		t.Fatalf("expected distinct text report paths, got %s", first.TextReportPath)
	}
	if first.MarkdownReportPath == second.MarkdownReportPath {
		t.Fatalf("expected distinct markdown report paths, got %s", first.MarkdownReportPath)
	}
	if _, err := os.Stat(first.ReportPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(second.ReportPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(first.MarkdownReportPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(second.MarkdownReportPath); err != nil {
		t.Fatal(err)
	}
	if len(first.Outcomes) != 1 || len(second.Outcomes) != 1 {
		t.Fatalf("unexpected outcomes: first=%#v second=%#v", first.Outcomes, second.Outcomes)
	}
	if first.Outcomes[0].StdoutPath == nil || second.Outcomes[0].StdoutPath == nil {
		t.Fatalf("expected stdout logs for both runs: first=%#v second=%#v", first.Outcomes, second.Outcomes)
	}
	if *first.Outcomes[0].StdoutPath == *second.Outcomes[0].StdoutPath {
		t.Fatalf("expected distinct stdout logs, got %s", *first.Outcomes[0].StdoutPath)
	}
}

func TestScenarioRunBlocksOnDrift(t *testing.T) {
	root, checkout := seedPinnedScenario(t)
	if err := os.WriteFile(filepath.Join(checkout, "README.md"), []byte("# changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, checkout, "git", "add", ".")
	run(t, checkout, "git", "commit", "-m", "change")

	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Blocked || !result.Drift || result.Failed {
		t.Fatalf("expected drift-blocked result, got %#v", result)
	}
	if got := result.Outcomes[0]; got.Status != "blocked" || !strings.Contains(got.Reason, "pinned ref drift") {
		t.Fatalf("unexpected drift outcome: %#v", got)
	}
}

func TestScenarioRunBlocksWhenCleanWorktreeRequired(t *testing.T) {
	root, checkout := seedPinnedScenario(t)
	lock, err := scenario.Load(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	lock.Checks[0].RequiresCleanWorktree = true
	lockPath, err := workspace.ScenarioPath(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(lockPath, lock); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(checkout, "README.md"), []byte("# dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Blocked || result.Failed || result.Drift {
		t.Fatalf("expected clean-worktree block, got %#v", result)
	}
	if got := result.Outcomes[0]; got.Status != "blocked" || !strings.Contains(got.Reason, "worktree is not clean") {
		t.Fatalf("unexpected clean-worktree outcome: %#v", got)
	}
}

func TestScenarioRunBlocksOnEmptyCommand(t *testing.T) {
	root, _ := seedPinnedScenario(t)
	lock, err := scenario.Load(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	lock.Checks[0].Run = ""
	lockPath, err := workspace.ScenarioPath(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(lockPath, lock); err != nil {
		t.Fatal(err)
	}
	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Blocked || result.Failed || result.Drift {
		t.Fatalf("expected blocked empty-command result, got %#v", result)
	}
	if got := result.Outcomes[0]; got.Status != "blocked" || !strings.Contains(got.Reason, "missing command") {
		t.Fatalf("unexpected empty-command outcome: %#v", got)
	}
}

func TestScenarioPinRejectsQuotedRunCommand(t *testing.T) {
	root, _, changeID := seedScenarioWorkspace(t)
	repo, err := workspace.LoadRepo(root, "app-web")
	if err != nil {
		t.Fatal(err)
	}
	entrypoint := repo.Entrypoints["test"]
	entrypoint.Run = `bin/test --name "two words"`
	repo.Entrypoints["test"] = entrypoint
	if err := manifest.WriteYAML(filepath.Join(root, "repos", "app-web", "repo.yaml"), repo); err != nil {
		t.Fatal(err)
	}

	_, err = scenario.Pin(root, "app-flow", changeID, time.Date(2026, 4, 19, 12, 5, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "quoted arguments") {
		t.Fatalf("expected quoted command error, got %v", err)
	}
}

func TestScenarioRunBlocksQuotedRunCommand(t *testing.T) {
	root, _ := seedPinnedScenario(t)
	lock, err := scenario.Load(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	lock.Checks[0].Run = `bin/test --name "two words"`
	lockPath, err := workspace.ScenarioPath(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(lockPath, lock); err != nil {
		t.Fatal(err)
	}

	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Blocked || result.Failed || result.Drift {
		t.Fatalf("expected blocked quoted-command result, got %#v", result)
	}
	if got := result.Outcomes[0]; got.Status != "blocked" || !strings.Contains(got.Reason, "quoted arguments") {
		t.Fatalf("unexpected quoted-command outcome: %#v", got)
	}
}

func TestScenarioRunFailsOnCommandFailure(t *testing.T) {
	root, checkout := seedPinnedScenario(t)
	writeTestCommand(t, checkout, `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "nope")
	os.Exit(7)
}
`)
	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Failed || result.Blocked || result.Drift {
		t.Fatalf("expected failed result, got %#v", result)
	}
	outcome := result.Outcomes[0]
	if outcome.Status != "failed" || outcome.StderrPath == nil || outcome.StdoutPath == nil {
		t.Fatalf("unexpected failed outcome: %#v", outcome)
	}
	stderr, err := os.ReadFile(filepath.Join(filepath.Dir(result.ReportPath), *outcome.StderrPath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stderr), "nope") {
		t.Fatalf("expected stderr log to contain command output, got %q", string(stderr))
	}
	text, err := os.ReadFile(result.TextReportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(text), "failed=1") || !strings.Contains(string(text), "stderr:") {
		t.Fatalf("unexpected failure text report:\n%s", string(text))
	}
	markdown, err := os.ReadFile(result.MarkdownReportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(markdown), "## Diagnostics") || !strings.Contains(string(markdown), "nope") {
		t.Fatalf("unexpected failure markdown report:\n%s", string(markdown))
	}
}

func TestScenarioRunFailsOnTimeout(t *testing.T) {
	root, checkout := seedPinnedScenario(t)
	writeTestCommand(t, checkout, `package main

import "time"

func main() {
	time.Sleep(2 * time.Second)
}
`)
	lock, err := scenario.Load(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	lock.Checks[0].TimeoutSeconds = 1
	lockPath, err := workspace.ScenarioPath(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(lockPath, lock); err != nil {
		t.Fatal(err)
	}
	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Failed || result.Blocked || result.Drift {
		t.Fatalf("expected timeout failure, got %#v", result)
	}
	if got := result.Outcomes[0]; got.Status != "failed" || !strings.Contains(got.Reason, "timed out") {
		t.Fatalf("unexpected timeout outcome: %#v", got)
	}
}

func TestScenarioRunBlocksUnsafeCWD(t *testing.T) {
	root, _ := seedPinnedScenario(t)
	lock, err := scenario.Load(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	lock.Checks[0].CWD = ".."
	lockPath, err := workspace.ScenarioPath(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(lockPath, lock); err != nil {
		t.Fatal(err)
	}

	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Blocked || result.Failed {
		t.Fatalf("expected blocked unsafe-cwd result, got %#v", result)
	}
	if got := result.Outcomes[0]; got.Status != "blocked" || !strings.Contains(got.Reason, "escapes repo checkout") {
		t.Fatalf("unexpected unsafe-cwd outcome: %#v", got)
	}
}

func TestScenarioRunBlocksSymlinkCWDOutsideCheckout(t *testing.T) {
	root, checkout := seedPinnedScenario(t)
	outside := t.TempDir()
	link := filepath.Join(checkout, "linked-out")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	lock, err := scenario.Load(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	lock.Checks[0].CWD = "linked-out"
	lockPath, err := workspace.ScenarioPath(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(lockPath, lock); err != nil {
		t.Fatal(err)
	}

	result, err := scenario.Run(root, "app-flow", time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Blocked || result.Failed {
		t.Fatalf("expected blocked symlink-cwd result, got %#v", result)
	}
	if got := result.Outcomes[0]; got.Status != "blocked" || !strings.Contains(got.Reason, "escapes repo checkout") {
		t.Fatalf("unexpected symlink-cwd outcome: %#v", got)
	}
}

func seedPinnedScenario(t *testing.T) (string, string) {
	t.Helper()
	root, checkout, changeID := seedScenarioWorkspace(t)
	lockPath, err := scenario.Pin(root, "app-flow", changeID, time.Date(2026, 4, 19, 12, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatal(err)
	}
	return root, checkout
}

func seedScenarioWorkspace(t *testing.T) (string, string, string) {
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
	changeID, err := workspace.CreateChange(root, "app-flow", "test rollout", "contract", time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return root, checkout, changeID
}

func initGitRepo(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# app-web\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeTestCommand(t, root, `package main

import "fmt"

func main() {
	fmt.Println("ok")
}
`)
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

func writeTestCommand(t *testing.T, root string, source string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "testcmd.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}
