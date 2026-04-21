package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
	"github.com/johnkil/polyrepo-workspace-kit/internal/telemetry"
	vscodeworkspace "github.com/johnkil/polyrepo-workspace-kit/internal/vscode"
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

func TestVersionCommand(t *testing.T) {
	out, err := executeCLI("version")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"wkit dev", "commit:", "date:", "dirty:", "builtBy:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in version output:\n%s", want, out)
		}
	}

	out, err = executeCLI("--version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "wkit dev") || !strings.Contains(out, "builtBy=source") {
		t.Fatalf("unexpected --version output:\n%s", out)
	}
}

func TestDemoCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("wkit demo requires POSIX sh in v0.x")
	}
	out, err := executeCLI("demo", "minimal")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"demo: minimal",
		"workspace:",
		"markdown-report:",
		"handoff-command:",
		"Results: passed=2 failed=0 blocked=0 skipped=0",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in demo output:\n%s", want, out)
		}
	}
}

func TestInitCommandScaffoldsWorkspace(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	app := t.TempDir()
	schema := t.TempDir()

	out, err := executeCLI(
		"init", root,
		"--repo", "app-web="+app,
		"--repo", "shared-schema="+schema,
		"--repo-kind", "shared-schema=contract",
		"--relation", "app-web:shared-schema:contract",
		"--context", "schema-rollout",
		"--change-title", "payload rollout",
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"initialized workspace at " + root,
		"registered app-web",
		"bound app-web to " + app,
		"registered shared-schema",
		"bound shared-schema to " + schema,
		"relation: app-web -> shared-schema kind=contract",
		"context: schema-rollout",
		"change: CHG-",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in init output:\n%s", want, out)
		}
	}

	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Relations) != 1 || doc.Relations[0].From != "app-web" || doc.Relations[0].To != "shared-schema" {
		t.Fatalf("unexpected relations: %#v", doc.Relations)
	}
	contexts, err := workspace.LoadContexts(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(contexts.Contexts["schema-rollout"].Repos, ","); got != "app-web,shared-schema" {
		t.Fatalf("unexpected context repos: %s", got)
	}
}

func TestInitCommandRejectsUnsupportedRelationKindBeforeWriting(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	app := t.TempDir()
	schema := t.TempDir()

	_, err := executeCLI(
		"init", root,
		"--repo", "app-web="+app,
		"--repo", "shared-schema="+schema,
		"--relation", "app-web:shared-schema:bogus",
	)
	if err == nil || !strings.Contains(err.Error(), `unsupported kind "bogus"`) {
		t.Fatalf("expected unsupported relation kind error, got %v", err)
	}
	if _, statErr := os.Stat(root); !os.IsNotExist(statErr) {
		t.Fatalf("init should reject invalid relation before writing workspace, stat err=%v", statErr)
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

func TestRelationsSuggestCommand(t *testing.T) {
	root := seedCLIWorkspace(t)
	app := t.TempDir()
	schema := t.TempDir()
	if _, err := workspace.SetBinding(root, "app-web", app); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "shared-schema", schema); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "package.json"), []byte(`{
  "name": "@acme/app-web",
  "dependencies": {
    "@acme/shared-schema": "1.0.0"
  }
}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(schema, "package.json"), []byte(`{"name":"@acme/shared-schema"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := executeCLI("--workspace", root, "relations", "suggest")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"suggestions:",
		`- app-web -> shared-schema kind=runtime source="package.json dependencies" evidence="@acme/shared-schema"`,
		"- --relation app-web:shared-schema:runtime",
		"suggestions are read-only",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in relations suggest output:\n%s", want, out)
		}
	}
}

func TestTelemetryCommands(t *testing.T) {
	root := seedCLIWorkspace(t)

	out, err := executeCLI("--workspace", root, "telemetry", "enable")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"telemetry: enabled", "enabled: true", "event_count: 0"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in telemetry enable output:\n%s", want, out)
		}
	}
	if err := telemetry.RecordIfEnabled(root, telemetry.Event{
		Timestamp:  "2026-04-21T12:00:00Z",
		Command:    "wkit status",
		ExitCode:   0,
		DurationMS: 10,
	}); err != nil {
		t.Fatal(err)
	}

	out, err = executeCLI("--workspace", root, "telemetry", "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "event_count: 1") {
		t.Fatalf("expected event count in telemetry status:\n%s", out)
	}
	out, err = executeCLI("--workspace", root, "telemetry", "export")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"command":"wkit status"`) {
		t.Fatalf("expected JSONL export, got:\n%s", out)
	}
	out, err = executeCLI("--workspace", root, "telemetry", "disable")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "telemetry: disabled") || !strings.Contains(out, "enabled: false") {
		t.Fatalf("expected disabled telemetry output:\n%s", out)
	}
}

func TestExecuteRootRecordsTelemetryWhenEnabled(t *testing.T) {
	root := seedCLIWorkspace(t)
	if _, err := telemetry.Enable(root, time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	record := &cliRunRecord{}
	cmd := newRootCommandWithRecorder(record)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--workspace", root, "info"})

	code := executeRoot(cmd, record, time.Now().Add(-time.Second))
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; output:\n%s", code, out.String())
	}
	data, err := telemetry.Export(root)
	if err != nil {
		t.Fatal(err)
	}
	events := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(events) != 1 {
		t.Fatalf("expected exactly one telemetry event, got %d:\n%s", len(events), string(data))
	}
	var event telemetry.Event
	if err := json.Unmarshal([]byte(events[0]), &event); err != nil {
		t.Fatal(err)
	}
	if event.Command != "wkit info" || event.Workspace != root || event.ExitCode != 0 {
		t.Fatalf("unexpected telemetry event: %#v", event)
	}
	if event.DurationMS < 0 {
		t.Fatalf("duration should not be negative: %#v", event)
	}
	if !containsString(event.Args, "--workspace") || !containsString(event.Args, root) {
		t.Fatalf("expected workspace flag in telemetry args: %#v", event.Args)
	}
}

func TestExecuteRootRecordsTelemetryForArgValidationFailure(t *testing.T) {
	root := seedCLIWorkspace(t)
	if _, err := telemetry.Enable(root, time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	record := &cliRunRecord{}
	cmd := newRootCommandWithRecorder(record)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--workspace", root, "change", "show"})

	code := executeRoot(cmd, record, time.Now().Add(-time.Second))
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d; output:\n%s", code, out.String())
	}
	data, err := telemetry.Export(root)
	if err != nil {
		t.Fatal(err)
	}
	events := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(events) != 1 {
		t.Fatalf("expected exactly one telemetry event, got %d:\n%s", len(events), string(data))
	}
	var event telemetry.Event
	if err := json.Unmarshal([]byte(events[0]), &event); err != nil {
		t.Fatal(err)
	}
	if event.Command != "wkit change show" || event.Workspace != root || event.ExitCode != 1 {
		t.Fatalf("unexpected telemetry event: %#v", event)
	}
	if event.DurationMS < 0 {
		t.Fatalf("duration should not be negative: %#v", event)
	}
	if !containsString(event.Args, "--workspace") || !containsString(event.Args, root) {
		t.Fatalf("expected workspace flag in telemetry args: %#v", event.Args)
	}
}

func TestExecuteRootRecordsTelemetryForUnknownCommand(t *testing.T) {
	root := seedCLIWorkspace(t)
	if _, err := telemetry.Enable(root, time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	args := []string{"--workspace", root, "nope"}
	record := &cliRunRecord{RawArgs: args}
	cmd := newRootCommandWithRecorder(record)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)

	code := executeRoot(cmd, record, time.Now().Add(-time.Second))
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d; output:\n%s", code, out.String())
	}
	data, err := telemetry.Export(root)
	if err != nil {
		t.Fatal(err)
	}
	events := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(events) != 1 {
		t.Fatalf("expected exactly one telemetry event, got %d:\n%s", len(events), string(data))
	}
	var event telemetry.Event
	if err := json.Unmarshal([]byte(events[0]), &event); err != nil {
		t.Fatal(err)
	}
	if event.Command != "wkit" || event.Workspace != root || event.ExitCode != 1 {
		t.Fatalf("unexpected telemetry event: %#v", event)
	}
	if event.DurationMS < 0 {
		t.Fatalf("duration should not be negative: %#v", event)
	}
	for _, want := range args {
		if !containsString(event.Args, want) {
			t.Fatalf("expected %q in telemetry args: %#v", want, event.Args)
		}
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

func TestHandoffCommand(t *testing.T) {
	root := seedCLIWorkspace(t)
	writeCLIContexts(t, root, map[string]model.Context{
		"app-flow": {Repos: []string{"app-web"}},
	})
	changeID, err := workspace.CreateChange(root, "app-flow", "payload rollout", "contract", time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	out, err := executeCLI("--workspace", root, "handoff", changeID)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# Handoff: payload rollout", "- change: `" + changeID + "`", "No scenario lock was found"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in handoff output:\n%s", want, out)
		}
	}
}

func TestVSCodeCommandsPlanDiffAndApply(t *testing.T) {
	root := seedCLIWorkspace(t)
	app := filepath.Join(t.TempDir(), "app-web")
	schema := filepath.Join(t.TempDir(), "shared-schema")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(schema, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "app-web", app); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "shared-schema", schema); err != nil {
		t.Fatal(err)
	}

	out, err := executeCLI("--workspace", root, "vscode", "plan")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "target: vscode") || !strings.Contains(out, "[new] workspace") {
		t.Fatalf("unexpected vscode plan output:\n%s", out)
	}

	out, err = executeCLI("--workspace", root, "vscode", "diff")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "+  \"folders\": [") || !strings.Contains(out, "+        \"label\": \"app-web: test\"") {
		t.Fatalf("unexpected vscode diff output:\n%s", out)
	}

	out, err = executeCLI("--workspace", root, "vscode", "apply", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	targetPath, err := vscodeworkspace.TargetPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "written: "+targetPath) {
		t.Fatalf("unexpected vscode apply output:\n%s", out)
	}

	out, err = executeCLI("--workspace", root, "vscode", "plan")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "[unchanged] workspace") {
		t.Fatalf("expected unchanged vscode plan output:\n%s", out)
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

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
