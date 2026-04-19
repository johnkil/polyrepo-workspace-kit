package install

import (
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

type Scope string

const (
	ScopeRepo Scope = "repo"
	ScopeUser Scope = "user"
)

const (
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
	Tool     string
	Scope    Scope
	RepoID   string
	UserRoot string
	Force    bool
	Backup   bool
	Now      time.Time
}

type Target struct {
	Tool         string
	Scope        Scope
	Path         string
	Kind         string
	Source       string
	Status       string
	Ownership    string
	BackupPath   string
	Notes        []string
	RenderedText string
	SourcePath   string
}

type Plan struct {
	Tool    string
	Scope   Scope
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

func BuildPlan(root string, opts PlanOptions) (Plan, error) {
	if opts.Tool == "" {
		opts.Tool = "portable"
	}
	if opts.Scope == "" {
		opts.Scope = ScopeRepo
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	targets, notes, err := adapterTargets(root, opts)
	if err != nil {
		return Plan{}, err
	}
	for index := range targets {
		targets[index] = evaluateTarget(targets[index], opts)
	}
	summary := map[string]int{}
	for _, target := range targets {
		summary[target.Status]++
	}
	return Plan{
		Tool:    opts.Tool,
		Scope:   opts.Scope,
		Targets: targets,
		Notes:   notes,
		Summary: summary,
	}, nil
}

func adapterTargets(root string, opts PlanOptions) ([]Target, []string, error) {
	switch opts.Tool {
	case "portable":
		return portableTargets(root, opts)
	case "codex":
		return portableBaselineAdapterTargets(root, opts, "codex", "Codex repo scope reuses the portable baseline directly.")
	case "opencode":
		return portableBaselineAdapterTargets(root, opts, "opencode", "OpenCode repo scope reuses the portable baseline directly.")
	case "copilot":
		return copilotTargets(root, opts)
	case "claude":
		return claudeTargets(root, opts)
	default:
		return nil, nil, fmt.Errorf("unsupported install tool %q", opts.Tool)
	}
}

func portableTargets(root string, opts PlanOptions) ([]Target, []string, error) {
	switch opts.Scope {
	case ScopeRepo:
		targets, err := portableRepoTargets(root, opts, "portable")
		if err != nil {
			return nil, nil, err
		}
		return targets, nil, nil
	case ScopeUser:
		if opts.RepoID != "" {
			return nil, nil, fmt.Errorf("repo id must be omitted for user-scope installs")
		}
		userRoot := opts.UserRoot
		if userRoot == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, nil, err
			}
			userRoot = home
		}
		absUserRoot, err := fsutil.Abs(userRoot)
		if err != nil {
			return nil, nil, err
		}
		targets := plannedSkillTargets(root, filepath.Join(absUserRoot, ".agents", "skills"), ScopeUser, "portable", "portable")
		notes := []string{"Portable user scope is skills-only; it does not invent a universal global instructions path."}
		return targets, notes, nil
	default:
		return nil, nil, fmt.Errorf("unsupported install scope %q", opts.Scope)
	}
}

func portableBaselineAdapterTargets(root string, opts PlanOptions, tool string, note string) ([]Target, []string, error) {
	if opts.Scope != ScopeRepo {
		return nil, nil, fmt.Errorf("%s user-scope installs are not part of the v0.x target surface", tool)
	}
	targets, err := portableRepoTargets(root, opts, tool)
	if err != nil {
		return nil, nil, err
	}
	return targets, []string{note}, nil
}

func portableRepoTargets(root string, opts PlanOptions, tool string) ([]Target, error) {
	if opts.RepoID == "" {
		return nil, fmt.Errorf("repo id is required for repo-scope installs")
	}
	checkout, err := workspace.ResolveRepoCheckout(root, opts.RepoID)
	if err != nil {
		return nil, err
	}
	text, err := RenderAgentsMD(root, opts.RepoID)
	if err != nil {
		return nil, err
	}
	targets := []Target{{
		Tool:         tool,
		Scope:        ScopeRepo,
		Path:         filepath.Join(checkout, "AGENTS.md"),
		Kind:         "instructions",
		Source:       "portable",
		RenderedText: text,
	}}
	targets = append(targets, plannedSkillTargets(root, filepath.Join(checkout, ".agents", "skills"), ScopeRepo, tool, "portable")...)
	return targets, nil
}

func copilotTargets(root string, opts PlanOptions) ([]Target, []string, error) {
	if opts.Scope != ScopeRepo {
		return nil, nil, fmt.Errorf("copilot user-scope installs are not part of the v0.x target surface")
	}
	if opts.RepoID == "" {
		return nil, nil, fmt.Errorf("repo id is required for repo-scope installs")
	}
	checkout, err := workspace.ResolveRepoCheckout(root, opts.RepoID)
	if err != nil {
		return nil, nil, err
	}
	text, err := RenderCopilotInstructions(root, opts.RepoID)
	if err != nil {
		return nil, nil, err
	}
	return []Target{{
		Tool:         "copilot",
		Scope:        ScopeRepo,
		Path:         filepath.Join(checkout, ".github", "copilot-instructions.md"),
		Kind:         "instructions",
		Source:       "copilot",
		RenderedText: text,
	}}, nil, nil
}

func claudeTargets(root string, opts PlanOptions) ([]Target, []string, error) {
	if opts.Scope != ScopeRepo {
		return nil, nil, fmt.Errorf("claude user-scope installs are not part of the v0.x target surface")
	}
	if opts.RepoID == "" {
		return nil, nil, fmt.Errorf("repo id is required for repo-scope installs")
	}
	checkout, err := workspace.ResolveRepoCheckout(root, opts.RepoID)
	if err != nil {
		return nil, nil, err
	}
	text, err := RenderClaudeMD(root, opts.RepoID)
	if err != nil {
		return nil, nil, err
	}
	targets := []Target{{
		Tool:         "claude",
		Scope:        ScopeRepo,
		Path:         filepath.Join(checkout, "CLAUDE.md"),
		Kind:         "instructions",
		Source:       "claude",
		RenderedText: text,
	}}
	targets = append(targets, plannedSkillTargets(root, filepath.Join(checkout, ".claude", "skills"), ScopeRepo, "claude", "claude")...)
	return targets, nil, nil
}

func RenderAgentsMD(root string, repoID string) (string, error) {
	workspaceDoc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return "", err
	}
	repoDoc, err := workspace.LoadRepo(root, repoID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`<!-- generated by wkit -->
<!-- canonical-source: coordination/workspace.yaml, repos/%s/repo.yaml, guidance/rules/* -->
<!-- derived-artifact: do not edit by hand -->

# AGENTS.md

Generated by Polyrepo Workspace Kit for %s.

## Core operating assumptions

- Start repo-local first.
- Treat repo-local executable truth as authoritative.
- Expand cross-repo only by declared relation, active change, or explicit task signal.
- Workspace model: %s.

## Read first

%s

## Stable entrypoints

%s

## Shared guidance

%s
`, repoID, repoID, workspaceDoc.Workspace.Model, renderReadFirst(repoDoc.ReadFirst), renderEntrypoints(repoDoc.Entrypoints), renderedRules(root)), nil
}

func RenderCopilotInstructions(root string, repoID string) (string, error) {
	repoDoc, err := workspace.LoadRepo(root, repoID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`<!-- generated by wkit -->
<!-- canonical-source: repos/%s/repo.yaml, guidance/rules/* -->
<!-- derived-artifact: do not edit by hand -->

# Copilot instructions for %s

Use repo-local commands and documentation as the primary source of truth.

- Start in this repository first.
- Prefer stable entrypoints from repos/%s/repo.yaml.
- Expand cross-repo only when a relation, live change, or explicit task signal requires it.

## Stable entrypoints

%s

## Shared guidance

%s
`, repoID, repoID, repoID, renderEntrypoints(repoDoc.Entrypoints), renderedRules(root)), nil
}

func renderReadFirst(items []string) string {
	if len(items) == 0 {
		return "- None declared"
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, "- `"+item+"`")
	}
	return strings.Join(lines, "\n")
}

func renderEntrypoints(entrypoints map[string]model.Entrypoint) string {
	if len(entrypoints) == 0 {
		return "- None declared"
	}
	names := make([]string, 0, len(entrypoints))
	for name := range entrypoints {
		names = append(names, name)
	}
	sort.Strings(names)
	lines := make([]string, 0, len(names))
	for _, name := range names {
		entrypoint := entrypoints[name]
		lines = append(lines, fmt.Sprintf("- `%s`: `%s`", name, entrypoint.Run))
	}
	return strings.Join(lines, "\n")
}

func RenderClaudeMD(root string, repoID string) (string, error) {
	if _, err := workspace.LoadRepo(root, repoID); err != nil {
		return "", err
	}
	return fmt.Sprintf(`<!-- generated by wkit -->
<!-- canonical-source: repos/%s/repo.yaml, guidance/rules/*, guidance/skills/* -->
<!-- derived-artifact: do not edit by hand -->

# Claude instructions for %s

- Use repo-local commands and documentation as the primary source of truth.
- Keep Claude-specific guidance small.
- Use .claude/skills/* for repeatable Claude-native workflows.
- Expand cross-repo only when a declared relation, live change, or explicit task signal requires it.

## Shared guidance

%s
`, repoID, repoID, renderedRules(root)), nil
}

func plannedSkillTargets(root string, dstRoot string, scope Scope, tool string, source string) []Target {
	sourceRoot := filepath.Join(root, "guidance", "skills")
	entries, err := os.ReadDir(sourceRoot)
	if err != nil {
		return nil
	}
	var targets []Target
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sourceDir := filepath.Join(sourceRoot, entry.Name())
		sourcePath := filepath.Join(sourceDir, "SKILL.md")
		if !fsutil.Exists(sourcePath) {
			continue
		}
		targetDir := filepath.Join(dstRoot, entry.Name())
		targets = append(targets, Target{
			Tool:       tool,
			Scope:      scope,
			Path:       filepath.Join(targetDir, "SKILL.md"),
			Kind:       "skill",
			Source:     source,
			SourcePath: sourcePath,
		})
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Path < targets[j].Path
	})
	return targets
}

func renderedRules(root string) string {
	rulesRoot := filepath.Join(root, "guidance", "rules")
	entries, err := os.ReadDir(rulesRoot)
	if err != nil {
		return "_No shared guidance rules defined._"
	}
	var blocks []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(rulesRoot, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		body := strings.TrimSpace(string(data))
		if body == "" {
			continue
		}
		blocks = append(blocks, fmt.Sprintf("### %s\n\n%s", entry.Name(), body))
	}
	if len(blocks) == 0 {
		return "_No shared guidance rules defined._"
	}
	return strings.Join(blocks, "\n\n")
}

func evaluateTarget(target Target, opts PlanOptions) Target {
	target.Ownership = ownership(target)
	if target.SourcePath != "" {
		target.Status = evaluateFileStatus(target.SourcePath, target.Path, opts)
	} else {
		target.Status = evaluateTextStatus(target.Path, target.RenderedText, opts)
	}
	if target.Status == StatusBackupOverwrite {
		target.BackupPath = fsutil.BackupPath(backupSubject(target), opts.Now)
	}
	return target
}

func evaluateTextStatus(path string, text string, opts PlanOptions) string {
	exists := fsutil.Exists(path)
	return evaluateStatus(exists, exists && fsutil.SameText(path, text), opts)
}

func evaluateFileStatus(source string, target string, opts PlanOptions) string {
	exists := fsutil.Exists(target)
	return evaluateStatus(exists, exists && fsutil.SameFile(source, target), opts)
}

func evaluateStatus(exists bool, unchanged bool, opts PlanOptions) string {
	if !exists {
		return StatusNew
	}
	if unchanged {
		return StatusUnchanged
	}
	if opts.Backup {
		return StatusBackupOverwrite
	}
	if opts.Force {
		return StatusOverwrite
	}
	return StatusBlocked
}

func ownership(target Target) string {
	if !fsutil.Exists(target.Path) {
		return OwnershipUnknown
	}
	if target.Kind == "instructions" {
		data, err := os.ReadFile(target.Path)
		if err == nil && strings.Contains(string(data), "generated by wkit") {
			return OwnershipWKitOwned
		}
		return OwnershipUnmarked
	}
	if target.Kind == "skill" {
		return OwnershipUnknown
	}
	return OwnershipUnknown
}

func backupSubject(target Target) string {
	return target.Path
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
		return ApplyResult{Plan: plan}, fmt.Errorf("%d blocked install target(s)", len(blocked))
	}
	var result ApplyResult
	result.Plan = plan
	for _, target := range plan.Targets {
		switch target.Status {
		case StatusUnchanged:
			result.Skipped = append(result.Skipped, target)
			continue
		case StatusNew, StatusOverwrite, StatusBackupOverwrite:
			if err := applyTarget(target); err != nil {
				return result, err
			}
			result.Written = append(result.Written, target)
		}
	}
	return result, nil
}

func applyTarget(target Target) error {
	if target.Status == StatusBackupOverwrite {
		if err := fsutil.BackupExisting(backupSubject(target), target.BackupPath); err != nil {
			return err
		}
	}
	if target.SourcePath != "" {
		return fsutil.CopyFile(target.SourcePath, target.Path)
	}
	return fsutil.WriteFileAtomic(target.Path, []byte(fsutil.NormalizeText(target.RenderedText)))
}

func DiffTarget(target Target) []string {
	desired, ok := desiredText(target)
	if !ok {
		return []string{fmt.Sprintf("# no textual diff available for %s\n", target.Path)}
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
	return unifiedDiff(fromFile, target.Path, current, desired)
}

func desiredText(target Target) (string, bool) {
	if target.RenderedText != "" {
		return fsutil.NormalizeText(target.RenderedText), true
	}
	if target.SourcePath != "" {
		data, err := os.ReadFile(target.SourcePath)
		if err != nil {
			return "", false
		}
		return fsutil.NormalizeText(string(data)), true
	}
	return "", false
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
	parts := strings.SplitAfter(text, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func SummaryKeys() []string {
	return []string{StatusNew, StatusUnchanged, StatusBlocked, StatusOverwrite, StatusBackupOverwrite}
}
