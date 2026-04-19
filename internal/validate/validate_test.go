package validate_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/validate"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

func TestWorkspaceValidationOK(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "app-web", t.TempDir()); err != nil {
		t.Fatal(err)
	}

	report := validate.Workspace(root)
	if len(report.Errors) != 0 {
		t.Fatalf("unexpected errors: %#v", report.Errors)
	}
}

func TestWorkspaceValidationCatchesMissingRepoManifest(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	doc.Repos = append(doc.Repos, "missing-repo")
	if err := workspace.SaveWorkspace(root, doc); err != nil {
		t.Fatal(err)
	}

	report := validate.Workspace(root)
	if len(report.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
}

func TestWorkspaceValidationCatchesOrphanRuleFileMismatch(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(filepath.Join(root, "coordination", "rules", "orphan.yaml"), model.RuleDocument{
		Version: 1,
		Rule: model.Rule{
			ID:   "different",
			Kind: "rollout-order",
			Policy: model.RulePolicy{
				Order: "provider-before-consumer",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	report := validate.Workspace(root)
	assertHasError(t, report, "rule.id")
}

func TestWorkspaceValidationCatchesQuotedEntrypointRun(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
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
	entrypoint.Run = `bin/test --name "two words"`
	repo.Entrypoints["test"] = entrypoint
	if err := manifest.WriteYAML(filepath.Join(root, "repos", "app-web", "repo.yaml"), repo); err != nil {
		t.Fatal(err)
	}

	report := validate.Workspace(root)
	assertHasError(t, report, "quoted arguments")
}

func TestWorkspaceValidationCatchesInvalidScenarioReport(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo-workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(root, "local", "reports", "app-flow", "20260419T121000Z.yaml")
	if err := manifest.WriteYAML(badPath, model.ScenarioReportDocument{
		Version: 1,
		Report: model.ScenarioReportMeta{
			Scenario:    "different-flow",
			GeneratedAt: "2026-04-19T12:10:00Z",
			ReportKind:  "local-validation-run",
		},
		Results: []model.ScenarioRunOutcome{
			{Check: "app-web:test", Status: "weird"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	report := validate.Workspace(root)
	if len(report.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
}

func TestWorkspaceValidationCatchesInvalidScenarioRefs(t *testing.T) {
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
	scenarioPath, err := workspace.ScenarioPath(root, "app-flow")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.WriteYAML(scenarioPath, model.ScenarioLockDocument{
		Version: 1,
		Scenario: model.ScenarioMeta{
			ID:      "other-flow",
			Change:  "CHG-2026-04-19-999",
			Context: "app-flow",
		},
		Repos: []model.ScenarioRepo{
			{Repo: "app-web"},
		},
		Checks: []model.ScenarioCheck{
			{ID: "app-web:test", Repo: "app-web", CWD: ".", Run: "bin/test", Status: "planned"},
			{ID: "app-web:test", Repo: "missing-repo", CWD: "..", Run: "bin/test", Status: "planned"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	report := validate.Workspace(root)
	assertHasError(t, report, "scenario.id")
	assertHasError(t, report, "references missing change")
	assertHasError(t, report, "duplicate check id")
	assertHasError(t, report, "path escapes repo checkout")
}

func assertHasError(t *testing.T, report validate.Report, want string) {
	t.Helper()
	for _, item := range report.Errors {
		if strings.Contains(item, want) {
			return
		}
	}
	t.Fatalf("expected error containing %q, got %#v", want, report.Errors)
}
