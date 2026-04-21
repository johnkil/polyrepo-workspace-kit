package vscode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/fsutil"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

const (
	WorkspaceRelPath = "local/vscode/workspace.code-workspace"

	StatusNew             = "new"
	StatusUnchanged       = "unchanged"
	StatusBlocked         = "blocked"
	StatusOverwrite       = "overwrite"
	StatusBackupOverwrite = "backup+overwrite"

	OwnershipWKitOwned = "wkit-owned"
	OwnershipUnmarked  = "unmarked"
	OwnershipUnknown   = "unknown"
)

type PlanOptions struct {
	Force  bool
	Backup bool
	Now    time.Time
}

type Target struct {
	Path         string
	Kind         string
	Source       string
	Status       string
	Ownership    string
	BackupPath   string
	Notes        []string
	RenderedText string
	BoundaryRoot string
}

type Plan struct {
	Targets []Target
	Notes   []string
	Summary map[string]int
}

type DiffItem struct {
	Target Target
	Lines  []string
}

type DiffPlan struct {
	Plan  Plan
	Items []DiffItem
}

type ApplyResult struct {
	Plan    Plan
	Written []Target
	Skipped []Target
}

type codeWorkspaceFile struct {
	Folders  []workspaceFolder `json:"folders"`
	Settings map[string]any    `json:"settings,omitempty"`
	Tasks    taskConfig        `json:"tasks"`
}

type workspaceFolder struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
}

type taskConfig struct {
	Version string          `json:"version"`
	Tasks   []workspaceTask `json:"tasks"`
}

type workspaceTask struct {
	Label          string            `json:"label"`
	Type           string            `json:"type"`
	Command        string            `json:"command"`
	Args           []string          `json:"args,omitempty"`
	Options        *taskOptions      `json:"options,omitempty"`
	Group          string            `json:"group,omitempty"`
	ProblemMatcher []string          `json:"problemMatcher"`
	Detail         string            `json:"detail,omitempty"`
	Presentation   *taskPresentation `json:"presentation,omitempty"`
}

type taskOptions struct {
	CWD string `json:"cwd,omitempty"`
}

type taskPresentation struct {
	Reveal string `json:"reveal,omitempty"`
	Panel  string `json:"panel,omitempty"`
}

func BuildPlan(root string, opts PlanOptions) (Plan, error) {
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	target, notes, err := target(root)
	if err != nil {
		return Plan{}, err
	}
	target = evaluateTarget(target, opts)
	summary := map[string]int{target.Status: 1}
	return Plan{
		Targets: []Target{target},
		Notes:   notes,
		Summary: summary,
	}, nil
}

func BuildDiff(root string, opts PlanOptions) (DiffPlan, error) {
	plan, err := BuildPlan(root, opts)
	if err != nil {
		return DiffPlan{}, err
	}
	var items []DiffItem
	for _, target := range plan.Targets {
		if target.Status == StatusUnchanged {
			continue
		}
		lines := DiffTarget(target)
		if len(lines) > 0 {
			items = append(items, DiffItem{Target: target, Lines: lines})
		}
	}
	return DiffPlan{Plan: plan, Items: items}, nil
}

func Apply(root string, opts PlanOptions) (ApplyResult, error) {
	plan, err := BuildPlan(root, opts)
	if err != nil {
		return ApplyResult{}, err
	}
	blocked := BlockedTargets(plan)
	if len(blocked) > 0 {
		return ApplyResult{Plan: plan}, fmt.Errorf("%d blocked VS Code workspace target(s)", len(blocked))
	}
	result := ApplyResult{Plan: plan}
	for _, target := range plan.Targets {
		switch target.Status {
		case StatusUnchanged:
			result.Skipped = append(result.Skipped, target)
		case StatusNew, StatusOverwrite, StatusBackupOverwrite:
			if err := applyTarget(target); err != nil {
				return result, err
			}
			result.Written = append(result.Written, target)
		}
	}
	return result, nil
}

func TargetPath(root string) (string, error) {
	absRoot, err := fsutil.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(absRoot, WorkspaceRelPath), nil
}

func BlockedTargets(plan Plan) []Target {
	var out []Target
	for _, target := range plan.Targets {
		if target.Status == StatusBlocked {
			out = append(out, target)
		}
	}
	return out
}

func DiffTarget(target Target) []string {
	if err := targetPathError(target); err != nil {
		return []string{fmt.Sprintf("# diff unavailable for %s: unsafe target path: %v\n", target.Path, err)}
	}
	current := ""
	fromFile := "/dev/null"
	if fsutil.Exists(target.Path) {
		data, err := os.ReadFile(target.Path)
		if err == nil {
			current = fsutil.NormalizeText(string(data))
			fromFile = target.Path
		}
	}
	return unifiedDiff(fromFile, target.Path, current, fsutil.NormalizeText(target.RenderedText))
}

func target(root string) (Target, []string, error) {
	absRoot, err := fsutil.Abs(root)
	if err != nil {
		return Target{}, nil, err
	}
	text, err := RenderWorkspace(absRoot)
	if err != nil {
		return Target{}, nil, err
	}
	return Target{
			Path:         filepath.Join(absRoot, WorkspaceRelPath),
			Kind:         "workspace",
			Source:       "vscode",
			RenderedText: text,
			BoundaryRoot: absRoot,
		}, []string{
			"VS Code workspace export is a local derived artifact; it does not write .vscode files into bound repositories.",
		}, nil
}

func RenderWorkspace(root string) (string, error) {
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return "", err
	}
	folders := []workspaceFolder{{
		Name: workspaceRootName(doc),
		Path: root,
	}}
	tasks := []workspaceTask{
		wkitTask(root, "wkit: overview", "overview"),
		wkitTask(root, "wkit: validate", "validate"),
		wkitTask(root, "wkit: doctor", "doctor"),
		wkitTask(root, "wkit: status", "status"),
	}
	for _, scenarioID := range scenarioIDs(root) {
		if err := workspace.ValidateID("scenario id", scenarioID); err != nil {
			return "", err
		}
		tasks = append(tasks,
			wkitTask(root, "scenario: "+scenarioID+": status", "scenario", "status", scenarioID),
			wkitTask(root, "scenario: "+scenarioID+": run", "scenario", "run", scenarioID),
		)
	}
	for _, repoID := range doc.Repos {
		if err := workspace.ValidateID("repo id", repoID); err != nil {
			return "", err
		}
		checkout, err := workspace.ResolveRepoCheckout(root, repoID)
		if err != nil {
			return "", err
		}
		repoDoc, err := workspace.LoadRepo(root, repoID)
		if err != nil {
			return "", err
		}
		folders = append(folders, workspaceFolder{Name: repoID, Path: checkout})
		repoTasks, err := repoEntrypointTasks(repoID, repoDoc)
		if err != nil {
			return "", err
		}
		tasks = append(tasks, repoTasks...)
	}

	data, err := json.MarshalIndent(codeWorkspaceFile{
		Folders: folders,
		Settings: map[string]any{
			"workbench.editor.labelFormat": "medium",
			"scm.alwaysShowRepositories":   true,
		},
		Tasks: taskConfig{
			Version: "2.0.0",
			Tasks:   tasks,
		},
	}, "", "  ")
	if err != nil {
		return "", err
	}
	return fsutil.NormalizeText(`// generated by wkit
// canonical-source: coordination/workspace.yaml, repos/*/repo.yaml, local/bindings.yaml, coordination/scenarios/*
// derived-artifact: do not edit by hand

` + string(data)), nil
}

func workspaceRootName(doc model.WorkspaceDocument) string {
	if strings.TrimSpace(doc.Workspace.ID) == "" {
		return "wkit"
	}
	return "wkit: " + doc.Workspace.ID
}

func wkitTask(root string, label string, args ...string) workspaceTask {
	fullArgs := append([]string{"--workspace", root}, args...)
	return workspaceTask{
		Label:          label,
		Type:           "process",
		Command:        "wkit",
		Args:           fullArgs,
		Options:        &taskOptions{CWD: root},
		ProblemMatcher: []string{},
		Detail:         "from wkit workspace",
		Presentation:   &taskPresentation{Reveal: "always", Panel: "dedicated"},
	}
}

func repoEntrypointTasks(repoID string, repoDoc model.RepoDocument) ([]workspaceTask, error) {
	names := make([]string, 0, len(repoDoc.Entrypoints))
	for name := range repoDoc.Entrypoints {
		names = append(names, name)
	}
	sort.Strings(names)
	tasks := make([]workspaceTask, 0, len(names))
	for _, name := range names {
		entrypoint := repoDoc.Entrypoints[name]
		parts, err := parseCommand(entrypoint.Run)
		if err != nil {
			return nil, fmt.Errorf("repo %s entrypoint %s: %w", repoID, name, err)
		}
		cwd, err := workspaceFolderCWD(repoID, entrypoint.CWD)
		if err != nil {
			return nil, fmt.Errorf("repo %s entrypoint %s cwd: %w", repoID, name, err)
		}
		task := workspaceTask{
			Label:          repoID + ": " + name,
			Type:           "process",
			Command:        parts[0],
			Args:           parts[1:],
			Options:        &taskOptions{CWD: cwd},
			ProblemMatcher: []string{},
			Detail:         fmt.Sprintf("from repos/%s/repo.yaml", repoID),
			Presentation:   &taskPresentation{Reveal: "always", Panel: "dedicated"},
		}
		if name == "test" || name == "build" {
			task.Group = name
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func workspaceFolderCWD(repoID string, cwd string) (string, error) {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}
	if filepath.IsAbs(cwd) {
		return "", fmt.Errorf("path must be relative")
	}
	clean := filepath.Clean(cwd)
	if clean == "." {
		return "${workspaceFolder:" + repoID + "}", nil
	}
	if escapesBoundary(clean) {
		return "", fmt.Errorf("path escapes repo checkout")
	}
	return "${workspaceFolder:" + repoID + "}/" + filepath.ToSlash(clean), nil
}

func scenarioIDs(root string) []string {
	scenarioRoot := filepath.Join(root, "coordination", "scenarios")
	entries, err := os.ReadDir(scenarioRoot)
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if fsutil.Exists(filepath.Join(scenarioRoot, entry.Name(), "manifest.lock.yaml")) {
			ids = append(ids, entry.Name())
		}
	}
	sort.Strings(ids)
	return ids
}

func parseCommand(run string) ([]string, error) {
	trimmed := strings.TrimSpace(run)
	if trimmed == "" {
		return nil, fmt.Errorf("empty command")
	}
	if strings.ContainsAny(trimmed, `"'`) {
		return nil, fmt.Errorf("quoted arguments are not supported; use a repo-local wrapper script")
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return parts, nil
}

func evaluateTarget(target Target, opts PlanOptions) Target {
	if err := targetPathError(target); err != nil {
		target.Status = StatusBlocked
		target.Ownership = OwnershipUnknown
		target.Notes = append(target.Notes, "unsafe target path: "+err.Error())
		return target
	}
	target.Ownership = ownership(target)
	exists := fsutil.Exists(target.Path)
	switch {
	case !exists:
		target.Status = StatusNew
	case fsutil.SameText(target.Path, target.RenderedText):
		target.Status = StatusUnchanged
	case opts.Backup:
		target.Status = StatusBackupOverwrite
	case opts.Force:
		target.Status = StatusOverwrite
	default:
		target.Status = StatusBlocked
	}
	if target.Status == StatusBackupOverwrite {
		target.BackupPath = fsutil.BackupPath(target.Path, opts.Now)
	}
	return target
}

func ownership(target Target) string {
	if !fsutil.Exists(target.Path) {
		return OwnershipUnknown
	}
	data, err := os.ReadFile(target.Path)
	if err == nil && strings.Contains(string(data), "generated by wkit") {
		return OwnershipWKitOwned
	}
	return OwnershipUnmarked
}

func applyTarget(target Target) error {
	if err := targetPathError(target); err != nil {
		return fmt.Errorf("unsafe VS Code workspace target %s: %w", target.Path, err)
	}
	if target.Status == StatusBackupOverwrite {
		if err := fsutil.BackupExisting(target.Path, target.BackupPath); err != nil {
			return err
		}
	}
	return fsutil.WriteFileAtomic(target.Path, []byte(fsutil.NormalizeText(target.RenderedText)))
}

func targetPathError(target Target) error {
	if target.BoundaryRoot == "" {
		return nil
	}
	return pathWithinBoundary(target.BoundaryRoot, target.Path)
}

func pathWithinBoundary(root string, path string) error {
	rootAbs, err := fsutil.Abs(root)
	if err != nil {
		return err
	}
	pathAbs, err := fsutil.Abs(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return err
	}
	if escapesBoundary(rel) {
		return fmt.Errorf("%s is outside %s", path, root)
	}
	if info, err := os.Lstat(pathAbs); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symlink", path)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	realRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return fmt.Errorf("boundary root %s is not accessible: %w", root, err)
	}
	ancestor, err := deepestExistingAncestor(pathAbs)
	if err != nil {
		return err
	}
	realAncestor, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return fmt.Errorf("%s is not accessible: %w", ancestor, err)
	}
	realRel, err := filepath.Rel(realRoot, realAncestor)
	if err != nil {
		return err
	}
	if escapesBoundary(realRel) {
		return fmt.Errorf("%s resolves outside %s", ancestor, root)
	}
	return nil
}

func deepestExistingAncestor(path string) (string, error) {
	current := path
	for {
		if _, err := os.Lstat(current); err == nil {
			return current, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("no existing ancestor for %s", path)
		}
		current = parent
	}
}

func escapesBoundary(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel)
}

func unifiedDiff(fromFile string, toFile string, current string, desired string) []string {
	if current == desired {
		return []string{"# unchanged\n"}
	}
	oldLines := splitLines(current)
	newLines := splitLines(desired)
	lines := []string{
		"--- " + fromFile + "\n",
		"+++ " + toFile + "\n",
		fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines)),
	}
	for _, line := range oldLines {
		lines = append(lines, "-"+line)
	}
	for _, line := range newLines {
		lines = append(lines, "+"+line)
	}
	return lines
}

func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	raw := strings.SplitAfter(text, "\n")
	if raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}
	return raw
}
