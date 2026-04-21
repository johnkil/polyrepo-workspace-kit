package demo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/handoff"
	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

const (
	KindMinimal = "minimal"
	KindFailure = "failure"
)

type Result struct {
	Kind               string
	WorkspaceRoot      string
	ReposRoot          string
	ChangeID           string
	ScenarioID         string
	ReportPath         string
	TextReportPath     string
	MarkdownReportPath string
	MarkdownReport     string
	HandoffMarkdown    string
	Failed             bool
	Blocked            bool
	Drift              bool
}

func Run(kind string, now time.Time) (Result, error) {
	if runtime.GOOS == "windows" {
		return Result{}, fmt.Errorf("wkit demo requires POSIX sh in v0.x")
	}
	if strings.TrimSpace(kind) == "" {
		kind = KindMinimal
	}
	if kind != KindMinimal && kind != KindFailure {
		return Result{}, fmt.Errorf("unknown demo %q; expected %q or %q", kind, KindMinimal, KindFailure)
	}
	tmpRoot, err := os.MkdirTemp("", "wkit-demo-"+kind+"-*")
	if err != nil {
		return Result{}, err
	}
	root := filepath.Join(tmpRoot, "workspace")
	reposRoot := filepath.Join(tmpRoot, "repos")
	if err := os.MkdirAll(reposRoot, 0o755); err != nil {
		return Result{}, err
	}
	if err := workspace.Init(root); err != nil {
		return Result{}, err
	}
	if err := configureWorkspace(root); err != nil {
		return Result{}, err
	}
	appRepo := filepath.Join(reposRoot, "app-web")
	schemaRepo := filepath.Join(reposRoot, "shared-schema")
	if err := createDemoRepo(appRepo, "# app-web\n", "app-web ok\n"); err != nil {
		return Result{}, err
	}
	if err := createDemoRepo(schemaRepo, "# shared-schema\n", "shared-schema ok\n"); err != nil {
		return Result{}, err
	}
	if _, err := workspace.SetBinding(root, "app-web", appRepo); err != nil {
		return Result{}, err
	}
	if _, err := workspace.SetBinding(root, "shared-schema", schemaRepo); err != nil {
		return Result{}, err
	}

	changeID, err := workspace.CreateChange(root, "schema-rollout", demoTitle(kind), "contract", now)
	if err != nil {
		return Result{}, err
	}
	if _, err := scenario.Pin(root, "schema-rollout", changeID, now.Add(time.Minute)); err != nil {
		return Result{}, err
	}
	if kind == KindFailure {
		if err := simulateFailure(schemaRepo, appRepo); err != nil {
			return Result{}, err
		}
	}
	run, err := scenario.Run(root, "schema-rollout", now.Add(2*time.Minute))
	if err != nil {
		return Result{}, err
	}
	markdown, err := os.ReadFile(run.MarkdownReportPath)
	if err != nil {
		return Result{}, err
	}
	handoffMarkdown, err := handoff.Markdown(root, changeID, handoff.Options{ScenarioID: "schema-rollout"})
	if err != nil {
		return Result{}, err
	}
	return Result{
		Kind:               kind,
		WorkspaceRoot:      root,
		ReposRoot:          reposRoot,
		ChangeID:           changeID,
		ScenarioID:         "schema-rollout",
		ReportPath:         run.ReportPath,
		TextReportPath:     run.TextReportPath,
		MarkdownReportPath: run.MarkdownReportPath,
		MarkdownReport:     string(markdown),
		HandoffMarkdown:    handoffMarkdown,
		Failed:             run.Failed,
		Blocked:            run.Blocked,
		Drift:              run.Drift,
	}, nil
}

func configureWorkspace(root string) error {
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return err
	}
	doc.Workspace.ID = "demo-workspace"
	doc.Workspace.Model = "thin-coordination-layer"
	doc.Relations = []model.Relation{{From: "app-web", To: "shared-schema", Kind: "contract"}}
	doc.Rules = []string{"contract-rollout-order"}
	if err := workspace.SaveWorkspace(root, doc); err != nil {
		return err
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		return err
	}
	if _, err := workspace.RegisterRepo(root, "shared-schema", "contract"); err != nil {
		return err
	}
	for _, repoID := range []string{"app-web", "shared-schema"} {
		repo, err := workspace.LoadRepo(root, repoID)
		if err != nil {
			return err
		}
		entrypoint := repo.Entrypoints["test"]
		entrypoint.Run = "sh bin/test"
		entrypoint.TimeoutSeconds = 30
		entrypoint.EnvProfile = "default"
		repo.Entrypoints["test"] = entrypoint
		if err := manifest.WriteYAML(filepath.Join(root, "repos", repoID, "repo.yaml"), repo); err != nil {
			return err
		}
	}
	if err := manifest.WriteYAML(filepath.Join(root, workspace.ContextsFile), model.ContextsDocument{
		Version: 1,
		Contexts: map[string]model.Context{
			"schema-rollout": {Repos: []string{"shared-schema", "app-web"}},
		},
	}); err != nil {
		return err
	}
	return manifest.WriteYAML(filepath.Join(root, "coordination", "rules", "contract-rollout-order.yaml"), model.RuleDocument{
		Version: 1,
		Rule: model.Rule{
			ID:   "contract-rollout-order",
			Kind: "rollout-order",
			AppliesTo: model.RuleAppliesTo{
				RelationKind: "contract",
			},
			Policy: model.RulePolicy{
				Order: "provider-before-consumer",
			},
		},
	})
}

func createDemoRepo(root string, readme string, testOutput string) error {
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(readme), 0o644); err != nil {
		return err
	}
	testScript := "#!/usr/bin/env sh\nset -eu\nprintf %s " + shellSingleQuote(testOutput) + "\n"
	if err := os.WriteFile(filepath.Join(root, "bin", "test"), []byte(testScript), 0o755); err != nil {
		return err
	}
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
		{"add", "."},
		{"commit", "-m", "init"},
	} {
		if err := run(root, "git", args...); err != nil {
			return err
		}
	}
	return nil
}

func simulateFailure(schemaRepo string, appRepo string) error {
	readme := filepath.Join(schemaRepo, "README.md")
	current, err := os.ReadFile(readme)
	if err != nil {
		return err
	}
	if err := os.WriteFile(readme, append(current, []byte("\n## payload v4\n\nRemoved legacy payload field.\n")...), 0o644); err != nil {
		return err
	}
	if err := run(schemaRepo, "git", "add", "README.md"); err != nil {
		return err
	}
	if err := run(schemaRepo, "git", "commit", "-m", "simulate schema drift"); err != nil {
		return err
	}
	failingScript := `#!/usr/bin/env sh
set -eu
echo "checking app-web against pinned schema"
echo "app-web contract check failed: payload field customer_id is missing" >&2
echo "hint: regenerate the client after reconciling shared-schema drift" >&2
exit 7
`
	return os.WriteFile(filepath.Join(appRepo, "bin", "test"), []byte(failingScript), 0o755)
}

func demoTitle(kind string) string {
	if kind == KindFailure {
		return "payload field rollout with drift"
	}
	return "payload field rollout"
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v failed: %w\n%s", name, args, err, string(out))
	}
	return nil
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
