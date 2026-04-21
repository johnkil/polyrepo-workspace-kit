package relations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/relations"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

func TestSuggestFromPackageJSON(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	writeFile(t, app, "package.json", `{
  "name": "@acme/app-web",
  "dependencies": {
    "@acme/shared-schema": "1.0.0"
  }
}
`)
	writeFile(t, schema, "package.json", `{"name":"@acme/shared-schema"}`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "package.json dependencies", "@acme/shared-schema")
}

func TestSuggestFromGoModAndSkipsExistingRelation(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	writeFile(t, app, "go.mod", `module github.com/acme/app-web

require github.com/acme/shared-schema v0.1.0
`)
	writeFile(t, schema, "go.mod", `module github.com/acme/shared-schema
`)
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	doc.Relations = []model.Relation{{From: "app-web", To: "shared-schema", Kind: "runtime"}}
	if err := workspace.SaveWorkspace(root, doc); err != nil {
		t.Fatal(err)
	}

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Suggestions) != 0 {
		t.Fatalf("expected existing relation to be skipped, got %#v", report.Suggestions)
	}
}

func TestSuggestFromCargoAndGradle(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	tools := t.TempDir()
	if _, err := workspace.RegisterRepo(root, "build-tools", "tooling"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "build-tools", tools); err != nil {
		t.Fatal(err)
	}
	writeFile(t, app, "Cargo.toml", `[package]
name = "app-web"

[dependencies]
shared-schema = "0.1"

[dev-dependencies]
build-tools = "0.1"
`)
	writeFile(t, schema, "Cargo.toml", `[package]
name = "shared-schema"
`)
	writeFile(t, tools, "settings.gradle.kts", `rootProject.name = "build-tools"`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "Cargo.toml dependencies", "shared-schema")
	assertSuggestion(t, report, "app-web", "build-tools", "build", "Cargo.toml dev-dependencies", "build-tools")
}

func TestSuggestGradleProjectDependencyKindFollowsConfiguration(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	tools := t.TempDir()
	if _, err := workspace.RegisterRepo(root, "build-tools", "tooling"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "build-tools", tools); err != nil {
		t.Fatal(err)
	}
	writeFile(t, app, "build.gradle.kts", `dependencies {
  implementation(project(":shared-schema"))
  testImplementation(project(":build-tools"))
}`)
	writeFile(t, schema, "settings.gradle.kts", `rootProject.name = "shared-schema"`)
	writeFile(t, tools, "settings.gradle.kts", `rootProject.name = "build-tools"`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle.kts", "shared-schema")
	assertSuggestion(t, report, "app-web", "build-tools", "build", "build.gradle.kts", "build-tools")
}

func TestSuggestGradleGroovyProjectDependencyKindFollowsConfiguration(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	tools := t.TempDir()
	if _, err := workspace.RegisterRepo(root, "build-tools", "tooling"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "build-tools", tools); err != nil {
		t.Fatal(err)
	}
	writeFile(t, app, "build.gradle", `dependencies {
  implementation project(':shared-schema')
  testImplementation project(':build-tools')
}`)
	writeFile(t, schema, "settings.gradle", `rootProject.name = 'shared-schema'`)
	writeFile(t, tools, "settings.gradle", `rootProject.name = 'build-tools'`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle", "shared-schema")
	assertSuggestion(t, report, "app-web", "build-tools", "build", "build.gradle", "build-tools")
}

func TestSuggestGradleNamedProjectPathSyntax(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	writeFile(t, app, "build.gradle", `dependencies {
  implementation project(path: ':shared-schema')
}`)
	writeFile(t, app, "build.gradle.kts", `dependencies {
  implementation(project(path = ":shared-schema"))
}`)
	writeFile(t, schema, "settings.gradle", `rootProject.name = 'shared-schema'`)
	writeFile(t, schema, "settings.gradle.kts", `rootProject.name = "shared-schema"`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle", "shared-schema")
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle.kts", "shared-schema")
}

func TestSuggestGradleNamedProjectPathSyntaxWithExtraArgs(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	writeFile(t, app, "build.gradle", `dependencies {
  implementation project(path: ':shared-schema', configuration: 'default')
}`)
	writeFile(t, app, "build.gradle.kts", `dependencies {
  implementation(project(path = ":shared-schema", configuration = "default"))
}`)
	writeFile(t, schema, "settings.gradle", `rootProject.name = 'shared-schema'`)
	writeFile(t, schema, "settings.gradle.kts", `rootProject.name = "shared-schema"`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle", "shared-schema")
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle.kts", "shared-schema")
}

func TestSuggestGradleNamedExternalDependencySyntax(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	writeFile(t, app, "build.gradle", `dependencies {
  implementation group: 'com.acme', name: 'shared-schema', version: '1.0.0'
}`)
	writeFile(t, app, "build.gradle.kts", `dependencies {
  implementation(group = "com.acme", name = "shared-schema", version = "1.0.0")
}`)
	writeFile(t, schema, "settings.gradle", `rootProject.name = 'shared-schema'`)
	writeFile(t, schema, "settings.gradle.kts", `rootProject.name = "shared-schema"`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle", "shared-schema")
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle.kts", "shared-schema")
}

func TestSuggestGradleOneLineDependencyBlocks(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	tools := t.TempDir()
	if _, err := workspace.RegisterRepo(root, "build-tools", "tooling"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "build-tools", tools); err != nil {
		t.Fatal(err)
	}
	writeFile(t, app, "build.gradle", `dependencies { implementation group: 'com.acme', name: 'shared-schema', version: '1.0.0' }
dependencies { compileOnly project(':build-tools') }`)
	writeFile(t, app, "build.gradle.kts", `dependencies { implementation("com.acme:shared-schema:1.0.0") }
dependencies { testImplementation(project(path = ":build-tools", configuration = "default")) }`)
	writeFile(t, schema, "settings.gradle", `rootProject.name = 'shared-schema'`)
	writeFile(t, schema, "settings.gradle.kts", `rootProject.name = "shared-schema"`)
	writeFile(t, tools, "settings.gradle", `rootProject.name = 'build-tools'`)
	writeFile(t, tools, "settings.gradle.kts", `rootProject.name = "build-tools"`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle", "shared-schema")
	assertSuggestion(t, report, "app-web", "build-tools", "build", "build.gradle", "build-tools")
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle.kts", "com.acme:shared-schema:1.0.0")
	assertSuggestion(t, report, "app-web", "build-tools", "build", "build.gradle.kts", "build-tools")
}

func TestSuggestGradleCompileOnlyProjectDependencyIsBuildKind(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	writeFile(t, app, "build.gradle", `dependencies {
  compileOnly project(':shared-schema')
}`)
	writeFile(t, schema, "settings.gradle", `rootProject.name = 'shared-schema'`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "build", "build.gradle", "shared-schema")
}

func TestSuggestGradleIgnoresNonDependencyProjectReferences(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	writeFile(t, app, "build.gradle", `def schemaProject = project(':shared-schema')`)
	writeFile(t, schema, "settings.gradle", `rootProject.name = 'shared-schema'`)

	report, err := relations.Suggest(root, relations.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Suggestions) != 0 {
		t.Fatalf("expected no suggestions for non-dependency project references, got %#v", report.Suggestions)
	}
}

func TestSuggestContextFilterAndMissingBinding(t *testing.T) {
	root, app, schema := seedWorkspace(t)
	if err := manifest.WriteYAML(filepath.Join(root, workspace.ContextsFile), model.ContextsDocument{
		Version: 1,
		Contexts: map[string]model.Context{
			"schema-rollout": {Repos: []string{"app-web", "shared-schema"}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "docs-site", "docs"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, app, "build.gradle.kts", `dependencies {
  implementation(project(":shared-schema"))
}`)
	writeFile(t, schema, "settings.gradle.kts", `rootProject.name = "shared-schema"`)

	report, err := relations.Suggest(root, relations.Options{ContextID: "schema-rollout"})
	if err != nil {
		t.Fatal(err)
	}
	assertSuggestion(t, report, "app-web", "shared-schema", "runtime", "build.gradle.kts", "shared-schema")
	for _, skipped := range report.Skipped {
		if skipped.Repo == "docs-site" {
			t.Fatalf("context filter should not inspect docs-site, skipped=%#v", report.Skipped)
		}
	}
}

func seedWorkspace(t *testing.T) (string, string, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace")
	app := t.TempDir()
	schema := t.TempDir()
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "shared-schema", "contract"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "app-web", app); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "shared-schema", schema); err != nil {
		t.Fatal(err)
	}
	return root, app, schema
}

func writeFile(t *testing.T, root string, rel string, data string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertSuggestion(t *testing.T, report relations.Report, from string, to string, kind string, source string, evidence string) {
	t.Helper()
	for _, suggestion := range report.Suggestions {
		if suggestion.From == from &&
			suggestion.To == to &&
			suggestion.Kind == kind &&
			suggestion.Source == source &&
			suggestion.Evidence == evidence {
			return
		}
	}
	var got []string
	for _, suggestion := range report.Suggestions {
		got = append(got, suggestion.From+"->"+suggestion.To+" "+suggestion.Kind+" "+suggestion.Source+" "+suggestion.Evidence)
	}
	t.Fatalf("missing suggestion %s->%s %s %s %s; got %s", from, to, kind, source, evidence, strings.Join(got, "; "))
}
