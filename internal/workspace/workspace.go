package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/fsutil"
	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
)

const (
	WorkspaceFile = "coordination/workspace.yaml"
	ContextsFile  = "coordination/contexts.yaml"
	BindingsFile  = "local/bindings.yaml"
)

var workspaceDirs = []string{
	"coordination/changes",
	"coordination/scenarios",
	"coordination/rules",
	"guidance/rules",
	"guidance/skills",
	"repos",
	"local/reports",
	"runtime",
	"config",
	"bin",
}

var safeIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*(\.[A-Za-z0-9][A-Za-z0-9_-]*)*$`)

func ValidateID(kind string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", kind)
	}
	if value != strings.TrimSpace(value) {
		return fmt.Errorf("%s %q must not have leading or trailing whitespace", kind, value)
	}
	if filepath.IsAbs(value) || strings.ContainsAny(value, `/\`) {
		return fmt.Errorf("%s %q must not contain path separators", kind, value)
	}
	if !safeIDPattern.MatchString(value) {
		return fmt.Errorf("%s %q must match [A-Za-z0-9][A-Za-z0-9._-]* without empty dot segments", kind, value)
	}
	return nil
}

func FindRoot(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(current); err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}
	for {
		if fsutil.Exists(filepath.Join(current, WorkspaceFile)) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("could not find workspace root containing %s", WorkspaceFile)
}

func Init(root string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	for _, rel := range workspaceDirs {
		if err := fsutil.EnsureDir(filepath.Join(absRoot, rel)); err != nil {
			return err
		}
	}
	if err := manifest.WriteYAMLIfMissing(filepath.Join(absRoot, WorkspaceFile), model.WorkspaceDocument{
		Version: 1,
		Workspace: model.WorkspaceMeta{
			ID:    filepath.Base(absRoot),
			Model: "thin-coordination-layer",
		},
		Repos:     []string{},
		Relations: []model.Relation{},
		Rules:     []string{},
	}); err != nil {
		return err
	}
	if err := manifest.WriteYAMLIfMissing(filepath.Join(absRoot, ContextsFile), model.ContextsDocument{
		Version:  1,
		Contexts: map[string]model.Context{},
	}); err != nil {
		return err
	}
	if err := manifest.WriteYAMLIfMissing(filepath.Join(absRoot, BindingsFile), model.BindingsDocument{
		Version:  1,
		Bindings: map[string]model.Binding{},
	}); err != nil {
		return err
	}
	if err := fsutil.WriteTextIfMissing(filepath.Join(absRoot, "guidance/rules/always-on.md"), `# Always-on rules

Keep this file short.

- Start repo-local first.
- Treat repo entrypoints as authoritative.
- Expand cross-repo only by relation, change, or explicit task signal.
- Prefer skills for multi-step procedures instead of growing this file.
`); err != nil {
		return err
	}
	return fsutil.WriteTextIfMissing(filepath.Join(absRoot, "guidance/skills/README.md"), `Place reusable SKILL.md bundles here.

Each skill should live in its own directory:

guidance/skills/<skill-name>/SKILL.md
`)
}

func LoadWorkspace(root string) (model.WorkspaceDocument, error) {
	var doc model.WorkspaceDocument
	err := manifest.LoadYAML(filepath.Join(root, WorkspaceFile), &doc)
	return doc, err
}

func SaveWorkspace(root string, doc model.WorkspaceDocument) error {
	if doc.Version == 0 {
		doc.Version = 1
	}
	return manifest.WriteYAML(filepath.Join(root, WorkspaceFile), doc)
}

func LoadContexts(root string) (model.ContextsDocument, error) {
	var doc model.ContextsDocument
	err := manifest.LoadYAML(filepath.Join(root, ContextsFile), &doc)
	if doc.Contexts == nil {
		doc.Contexts = map[string]model.Context{}
	}
	return doc, err
}

func LoadBindings(root string) (model.BindingsDocument, error) {
	var doc model.BindingsDocument
	err := manifest.LoadYAML(filepath.Join(root, BindingsFile), &doc)
	if doc.Bindings == nil {
		doc.Bindings = map[string]model.Binding{}
	}
	return doc, err
}

func ResolveRepoCheckout(root string, repoID string) (string, error) {
	bindings, err := LoadBindings(root)
	if err != nil {
		return "", err
	}
	binding, ok := bindings.Bindings[repoID]
	if !ok {
		return "", fmt.Errorf("missing local binding for repo %q", repoID)
	}
	if binding.Path == "" {
		return "", fmt.Errorf("local binding for repo %q has empty path", repoID)
	}
	absPath, err := fsutil.Abs(binding.Path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("bound path for repo %q is not accessible: %w", repoID, err)
	}
	return absPath, nil
}

func SaveBindings(root string, doc model.BindingsDocument) error {
	if doc.Version == 0 {
		doc.Version = 1
	}
	if doc.Bindings == nil {
		doc.Bindings = map[string]model.Binding{}
	}
	return manifest.WriteYAML(filepath.Join(root, BindingsFile), doc)
}

func LoadRepo(root string, repoID string) (model.RepoDocument, error) {
	var doc model.RepoDocument
	path, err := RepoManifestPath(root, repoID)
	if err != nil {
		return doc, err
	}
	err = manifest.LoadYAML(path, &doc)
	return doc, err
}

func LoadChange(root string, changeID string) (model.ChangeDocument, error) {
	var doc model.ChangeDocument
	path, err := ChangePath(root, changeID)
	if err != nil {
		return doc, err
	}
	err = manifest.LoadYAML(path, &doc)
	return doc, err
}

func RepoManifestPath(root string, repoID string) (string, error) {
	if err := ValidateID("repo id", repoID); err != nil {
		return "", err
	}
	return filepath.Join(root, "repos", repoID, "repo.yaml"), nil
}

func ChangePath(root string, changeID string) (string, error) {
	if err := ValidateID("change id", changeID); err != nil {
		return "", err
	}
	return filepath.Join(root, "coordination", "changes", changeID+".yaml"), nil
}

func ScenarioPath(root string, scenarioID string) (string, error) {
	if err := ValidateID("scenario id", scenarioID); err != nil {
		return "", err
	}
	return filepath.Join(root, "coordination", "scenarios", scenarioID, "manifest.lock.yaml"), nil
}

func RepoIDs(doc model.WorkspaceDocument) map[string]struct{} {
	out := make(map[string]struct{}, len(doc.Repos))
	for _, repoID := range doc.Repos {
		out[repoID] = struct{}{}
	}
	return out
}

func RegisterRepo(root string, repoID string, kind string) (string, error) {
	if err := ValidateID("repo id", repoID); err != nil {
		return "", err
	}
	if strings.TrimSpace(kind) == "" {
		return "", fmt.Errorf("repo kind is required")
	}
	doc, err := LoadWorkspace(root)
	if err != nil {
		return "", err
	}
	manifestPath, err := RepoManifestPath(root, repoID)
	if err != nil {
		return "", err
	}
	if !fsutil.Exists(manifestPath) {
		if err := manifest.WriteYAML(manifestPath, model.RepoDocument{
			Version: 1,
			Repo: model.RepoMeta{
				ID:   repoID,
				Kind: kind,
			},
			ReadFirst: []string{},
			Entrypoints: map[string]model.Entrypoint{
				"test": {
					Run:            "bin/test",
					CWD:            ".",
					TimeoutSeconds: 600,
					EnvProfile:     "default",
				},
			},
		}); err != nil {
			return "", err
		}
	}
	if !contains(doc.Repos, repoID) {
		doc.Repos = append(doc.Repos, repoID)
		sort.Strings(doc.Repos)
		if err := SaveWorkspace(root, doc); err != nil {
			return "", err
		}
	}
	return manifestPath, nil
}

func SetBinding(root string, repoID string, path string) (string, error) {
	if err := ValidateID("repo id", repoID); err != nil {
		return "", err
	}
	doc, err := LoadWorkspace(root)
	if err != nil {
		return "", err
	}
	if _, ok := RepoIDs(doc)[repoID]; !ok {
		return "", fmt.Errorf("unknown repo %q; register it first", repoID)
	}
	absPath, err := fsutil.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("binding path for repo %q is not accessible: %w", repoID, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("binding path for repo %q is not a directory: %s", repoID, absPath)
	}
	bindings, err := LoadBindings(root)
	if err != nil {
		return "", err
	}
	bindings.Bindings[repoID] = model.Binding{Path: absPath}
	if err := SaveBindings(root, bindings); err != nil {
		return "", err
	}
	return absPath, nil
}

func NextChangeID(root string, now time.Time) string {
	prefix := "CHG-" + now.Format("2006-01-02") + "-"
	entries, _ := filepath.Glob(filepath.Join(root, "coordination", "changes", prefix+"*.yaml"))
	maxSuffix := 0
	for _, entry := range entries {
		name := strings.TrimSuffix(filepath.Base(entry), ".yaml")
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		suffix, err := strconv.Atoi(strings.TrimPrefix(name, prefix))
		if err != nil {
			continue
		}
		if suffix > maxSuffix {
			maxSuffix = suffix
		}
	}
	return fmt.Sprintf("%s%03d", prefix, maxSuffix+1)
}

func CreateChange(root string, contextID string, title string, kind string, now time.Time) (string, error) {
	if strings.TrimSpace(contextID) == "" {
		return "", fmt.Errorf("context is required")
	}
	if strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("change title is required")
	}
	if strings.TrimSpace(kind) == "" {
		kind = "contract"
	}
	contexts, err := LoadContexts(root)
	if err != nil {
		return "", err
	}
	context, ok := contexts.Contexts[contextID]
	if !ok {
		return "", fmt.Errorf("unknown context %q", contextID)
	}
	for attempt := 0; attempt < 1000; attempt++ {
		changeID := NextChangeID(root, now)
		path, err := ChangePath(root, changeID)
		if err != nil {
			return "", err
		}
		err = manifest.WriteYAMLExclusive(path, model.ChangeDocument{
			Version: 1,
			Change: model.Change{
				ID:      changeID,
				Title:   title,
				Kind:    kind,
				Context: contextID,
				Repos:   context.Repos,
			},
		})
		if err == nil {
			return changeID, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("could not allocate unique change id for %s", now.Format("2006-01-02"))
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
