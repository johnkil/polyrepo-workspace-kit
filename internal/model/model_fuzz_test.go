package model

import (
	"testing"

	"go.yaml.in/yaml/v3"
)

func FuzzModelYAMLDocuments(f *testing.F) {
	for _, seed := range []string{
		"repo:\n  id: app-web\n  kind: app\nentrypoints:\n  test: bin/test\n",
		"repo:\n  id: app-web\n  kind: app\nentrypoints:\n  test:\n    run: bin/test\n    cwd: .\n",
		"repo:\n  id: app-web\n  kind: app\nentrypoints:\n  test:\n    run: bin/test\n    timeout_second: 30\n",
		"bindings:\n  app-web: /tmp/app-web\n",
		"bindings:\n  app-web:\n    path: /tmp/app-web\n",
		"report:\n  scenario: app-flow\n  report_kind: local-validation-run\nresults:\n  - check: app-web:test\n    status: passed\n",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		if len(data) > 4096 {
			return
		}

		var repo RepoDocument
		if err := yaml.Unmarshal([]byte(data), &repo); err == nil {
			for name, entrypoint := range repo.Entrypoints {
				if entrypoint.Run != "" && entrypoint.CWD == "" {
					t.Fatalf("entrypoint %q decoded with empty cwd", name)
				}
			}
		}

		var bindings BindingsDocument
		_ = yaml.Unmarshal([]byte(data), &bindings)

		var scenario ScenarioLockDocument
		_ = yaml.Unmarshal([]byte(data), &scenario)

		var report ScenarioReportDocument
		_ = yaml.Unmarshal([]byte(data), &report)
	})
}
