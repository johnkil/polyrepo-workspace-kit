package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/scaffold"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

func TestApplyRegistersBindsContextRelationsAndChange(t *testing.T) {
	app := t.TempDir()
	schema := t.TempDir()
	root := filepath.Join(t.TempDir(), "workspace")

	result, err := scaffold.Apply(scaffold.Options{
		Root: root,
		Repos: []scaffold.RepoSpec{
			{ID: "app-web", Path: app, Kind: "app"},
			{ID: "shared-schema", Path: schema, Kind: "contract"},
		},
		Relations: []scaffold.RelationSpec{
			{From: "app-web", To: "shared-schema", Kind: "contract"},
		},
		ContextID:   "schema-rollout",
		ChangeTitle: "payload rollout",
		ChangeKind:  "contract",
		Now:         time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ContextID != "schema-rollout" || result.ChangeID != "CHG-2026-04-19-001" {
		t.Fatalf("unexpected scaffold result: %#v", result)
	}
	if _, err := os.Stat(result.ChangePath); err != nil {
		t.Fatal(err)
	}

	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Relations) != 1 || doc.Relations[0].Kind != "contract" {
		t.Fatalf("unexpected relations: %#v", doc.Relations)
	}
	contexts, err := workspace.LoadContexts(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := contexts.Contexts["schema-rollout"].Repos; strings.Join(got, ",") != "app-web,shared-schema" {
		t.Fatalf("unexpected context repos: %#v", got)
	}
	bindings, err := workspace.LoadBindings(root)
	if err != nil {
		t.Fatal(err)
	}
	if bindings.Bindings["app-web"].Path != app || bindings.Bindings["shared-schema"].Path != schema {
		t.Fatalf("unexpected bindings: %#v", bindings.Bindings)
	}
}

func TestApplyRejectsUnknownRelationEndpoint(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	_, err := scaffold.Apply(scaffold.Options{
		Root: root,
		Repos: []scaffold.RepoSpec{
			{ID: "app-web", Path: t.TempDir(), Kind: "app"},
		},
		Relations: []scaffold.RelationSpec{
			{From: "app-web", To: "shared-schema", Kind: "contract"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "unknown relation repo") {
		t.Fatalf("expected unknown relation repo error, got %v", err)
	}
}

func TestApplyDoesNotOverwriteDifferentContext(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	app := t.TempDir()
	schema := t.TempDir()
	if _, err := scaffold.Apply(scaffold.Options{
		Root:      root,
		Repos:     []scaffold.RepoSpec{{ID: "app-web", Path: app, Kind: "app"}},
		ContextID: "app-flow",
	}); err != nil {
		t.Fatal(err)
	}

	_, err := scaffold.Apply(scaffold.Options{
		Root:      root,
		Repos:     []scaffold.RepoSpec{{ID: "shared-schema", Path: schema, Kind: "contract"}},
		ContextID: "app-flow",
	})
	if err == nil || !strings.Contains(err.Error(), "already exists with different repos") {
		t.Fatalf("expected context overwrite error, got %v", err)
	}
}

func TestParseSpecs(t *testing.T) {
	repo, err := scaffold.ParseRepoSpec("app-web=/tmp/path=with-equals")
	if err != nil {
		t.Fatal(err)
	}
	if repo.ID != "app-web" || repo.Path != "/tmp/path=with-equals" || repo.Kind != "app" {
		t.Fatalf("unexpected repo spec: %#v", repo)
	}
	id, kind, err := scaffold.ParseRepoKindSpec("shared-schema=contract")
	if err != nil {
		t.Fatal(err)
	}
	if id != "shared-schema" || kind != "contract" {
		t.Fatalf("unexpected repo kind: %s %s", id, kind)
	}
	relation, err := scaffold.ParseRelationSpec("app-web:shared-schema:contract")
	if err != nil {
		t.Fatal(err)
	}
	if relation.From != "app-web" || relation.To != "shared-schema" || relation.Kind != "contract" {
		t.Fatalf("unexpected relation: %#v", relation)
	}
}
