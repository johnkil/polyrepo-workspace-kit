package model

import (
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestEntrypointRejectsUnknownMappingFields(t *testing.T) {
	var doc RepoDocument
	err := yaml.Unmarshal([]byte(`
repo:
  id: app-web
  kind: app
read_first: []
entrypoints:
  test:
    run: bin/test
    timeout_second: 30
`), &doc)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestBindingRejectsUnknownMappingFields(t *testing.T) {
	var doc BindingsDocument
	err := yaml.Unmarshal([]byte(`
bindings:
  app-web:
    path: /tmp/app-web
    extra: nope
`), &doc)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}
