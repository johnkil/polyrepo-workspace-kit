package orient

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/johnkil/polyrepo-workspace-kit/internal/fsutil"
	"github.com/johnkil/polyrepo-workspace-kit/internal/gitstate"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
	"github.com/johnkil/polyrepo-workspace-kit/internal/validate"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

type Count struct {
	Name  string
	Count int
}

type ContextSummary struct {
	ID        string
	RepoCount int
}

type Info struct {
	WorkspaceID    string
	Root           string
	RepoCount      int
	RepoKinds      []Count
	RelationKinds  []Count
	Contexts       []ContextSummary
	ChangeCount    int
	LatestChange   string
	ScenarioCount  int
	LatestScenario string
	BoundRepos     int
	TotalRepos     int
	GuidanceRules  int
	GuidanceSkills int
}

type StatusOptions struct {
	ContextID string
}

type RepoStatus struct {
	RepoID         string
	BindingStatus  string
	BindingPath    string
	GitStatus      string
	Branch         string
	Detached       bool
	Commit         string
	FullCommit     string
	DirtyFiles     int
	UntrackedFiles int
	Upstream       string
	HasUpstream    bool
	HasDivergence  bool
	Ahead          int
	Behind         int
	Reason         string
	ScenarioStatus string
	ScenarioReason string
	PinnedCommit   string
	CurrentCommit  string
}

type StatusReport struct {
	Repos []RepoStatus
}

type ScenarioStatusReport struct {
	ScenarioID string
	Repos      []RepoStatus
	Drift      bool
	Blocked    bool
	Missing    bool
}

type DoctorReport struct {
	Errors   []string
	Warnings []string
}

func ListContexts(root string) ([]ContextSummary, error) {
	contexts, err := workspace.LoadContexts(root)
	if err != nil {
		return nil, err
	}
	items := make([]ContextSummary, 0, len(contexts.Contexts))
	for id, context := range contexts.Contexts {
		items = append(items, ContextSummary{ID: id, RepoCount: len(context.Repos)})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func GetContext(root string, contextID string) (model.Context, error) {
	contexts, err := workspace.LoadContexts(root)
	if err != nil {
		return model.Context{}, err
	}
	context, ok := contexts.Contexts[contextID]
	if !ok {
		return model.Context{}, fmt.Errorf("unknown context %q", contextID)
	}
	return context, nil
}

func WorkspaceInfo(root string) (Info, error) {
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return Info{}, err
	}
	contexts, err := ListContexts(root)
	if err != nil {
		return Info{}, err
	}
	bindings, err := workspace.LoadBindings(root)
	if err != nil {
		return Info{}, err
	}

	repoKinds := map[string]int{}
	for _, repoID := range doc.Repos {
		repoDoc, err := workspace.LoadRepo(root, repoID)
		if err != nil {
			return Info{}, fmt.Errorf("load repo %q: %w", repoID, err)
		}
		repoKinds[repoDoc.Repo.Kind]++
	}
	relationKinds := map[string]int{}
	for _, relation := range doc.Relations {
		relationKinds[relation.Kind]++
	}
	changes := changeIDs(root)
	scenarios := scenarioIDs(root)

	return Info{
		WorkspaceID:    doc.Workspace.ID,
		Root:           root,
		RepoCount:      len(doc.Repos),
		RepoKinds:      counts(repoKinds),
		RelationKinds:  counts(relationKinds),
		Contexts:       contexts,
		ChangeCount:    len(changes),
		LatestChange:   last(changes),
		ScenarioCount:  len(scenarios),
		LatestScenario: last(scenarios),
		BoundRepos:     boundRepoCount(doc.Repos, bindings),
		TotalRepos:     len(doc.Repos),
		GuidanceRules:  fileCount(filepath.Join(root, "guidance", "rules"), "*.md"),
		GuidanceSkills: skillCount(filepath.Join(root, "guidance", "skills")),
	}, nil
}

func WorkspaceStatus(root string, opts StatusOptions) (StatusReport, error) {
	repoIDs, err := statusRepoIDs(root, opts.ContextID)
	if err != nil {
		return StatusReport{}, err
	}
	bindings, err := workspace.LoadBindings(root)
	if err != nil {
		return StatusReport{}, err
	}
	report := StatusReport{Repos: make([]RepoStatus, 0, len(repoIDs))}
	for _, repoID := range repoIDs {
		report.Repos = append(report.Repos, inspectRepo(bindings, repoID))
	}
	return report, nil
}

func ScenarioStatus(root string, scenarioID string) (ScenarioStatusReport, error) {
	lock, err := scenario.Load(root, scenarioID)
	if err != nil {
		return ScenarioStatusReport{}, err
	}
	bindings, err := workspace.LoadBindings(root)
	if err != nil {
		return ScenarioStatusReport{}, err
	}
	report := ScenarioStatusReport{ScenarioID: scenarioID, Repos: make([]RepoStatus, 0, len(lock.Repos))}
	for _, pinned := range lock.Repos {
		status := inspectRepo(bindings, pinned.Repo)
		status.PinnedCommit = short(pinned.Revision.Commit)
		status.CurrentCommit = status.Commit
		switch {
		case pinned.Revision.Commit == "":
			status.ScenarioStatus = "missing"
			status.ScenarioReason = "scenario lock has no pinned commit"
			report.Missing = true
		case status.GitStatus != "ok":
			status.ScenarioStatus = "blocked"
			status.ScenarioReason = status.Reason
			report.Blocked = true
		case status.FullCommit != pinned.Revision.Commit:
			status.ScenarioStatus = "drift"
			status.ScenarioReason = fmt.Sprintf("current HEAD %s does not match scenario lock %s", status.CurrentCommit, short(pinned.Revision.Commit))
			report.Drift = true
		default:
			status.ScenarioStatus = "ok"
		}
		report.Repos = append(report.Repos, status)
	}
	return report, nil
}

func Doctor(root string) DoctorReport {
	base := validate.Workspace(root)
	report := DoctorReport{
		Errors:   append([]string{}, base.Errors...),
		Warnings: append([]string{}, base.Warnings...),
	}
	local := localDiagnostics(root)
	report.Errors = appendUnique(report.Errors, local.Errors...)
	report.Warnings = appendUnique(report.Warnings, local.Warnings...)
	sort.Strings(report.Errors)
	sort.Strings(report.Warnings)
	return report
}

func statusRepoIDs(root string, contextID string) ([]string, error) {
	if contextID != "" {
		context, err := GetContext(root, contextID)
		if err != nil {
			return nil, err
		}
		doc, err := workspace.LoadWorkspace(root)
		if err != nil {
			return nil, err
		}
		repoIDs := workspace.RepoIDs(doc)
		for _, repoID := range context.Repos {
			if _, ok := repoIDs[repoID]; !ok {
				return nil, fmt.Errorf("context %q references unknown repo %q", contextID, repoID)
			}
		}
		return append([]string{}, context.Repos...), nil
	}
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return nil, err
	}
	return append([]string{}, doc.Repos...), nil
}

func inspectRepo(bindings model.BindingsDocument, repoID string) RepoStatus {
	status := RepoStatus{
		RepoID:        repoID,
		BindingStatus: "missing",
		GitStatus:     "unknown",
		Upstream:      "none",
	}
	binding, ok := bindings.Bindings[repoID]
	if !ok || strings.TrimSpace(binding.Path) == "" {
		status.Reason = fmt.Sprintf("missing local binding for repo %q", repoID)
		return status
	}
	absPath, err := fsutil.Abs(binding.Path)
	if err != nil {
		status.BindingStatus = "invalid"
		status.Reason = err.Error()
		return status
	}
	status.BindingPath = absPath
	info, err := os.Stat(absPath)
	if err != nil {
		status.BindingStatus = "inaccessible"
		status.Reason = fmt.Sprintf("bound path is not accessible: %v", err)
		return status
	}
	if !info.IsDir() {
		status.BindingStatus = "not-directory"
		status.Reason = "bound path is not a directory"
		return status
	}
	status.BindingStatus = "ok"

	git, err := gitstate.Inspect(absPath)
	if err != nil {
		status.GitStatus = "error"
		status.Reason = err.Error()
		return status
	}
	if !git.Git {
		status.GitStatus = "not-git"
		status.Reason = "bound path is not a git worktree"
		return status
	}
	status.GitStatus = "ok"
	status.Branch = git.Branch
	status.Detached = git.Detached
	status.Commit = git.Short
	status.FullCommit = git.Commit
	status.DirtyFiles = len(git.DirtyPaths)
	status.UntrackedFiles = len(git.UntrackedPaths)
	status.HasUpstream = git.HasUpstream
	if git.HasUpstream {
		status.Upstream = git.Upstream
		status.HasDivergence = git.HasDivergence
		status.Ahead = git.Ahead
		status.Behind = git.Behind
	}
	return status
}

func localDiagnostics(root string) DoctorReport {
	report := DoctorReport{}
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return report
	}
	bindings, err := workspace.LoadBindings(root)
	if err != nil {
		return report
	}
	for _, repoID := range doc.Repos {
		status := inspectRepo(bindings, repoID)
		switch status.BindingStatus {
		case "missing":
			report.Warnings = append(report.Warnings, fmt.Sprintf("repo %s: missing local binding (run: wkit bind set %s <path>)", repoID, repoID))
		case "inaccessible", "invalid", "not-directory":
			report.Warnings = append(report.Warnings, fmt.Sprintf("repo %s: unusable local binding %q (%s)", repoID, status.BindingPath, status.Reason))
		}
		if status.BindingStatus == "ok" && status.GitStatus == "not-git" {
			report.Warnings = append(report.Warnings, fmt.Sprintf("repo %s: bound path is not a git worktree: %s", repoID, status.BindingPath))
		}
		if status.BindingStatus == "ok" && status.GitStatus == "error" {
			report.Warnings = append(report.Warnings, fmt.Sprintf("repo %s: git inspection failed for %s (%s)", repoID, status.BindingPath, status.Reason))
		}
		repoDoc, err := workspace.LoadRepo(root, repoID)
		if err == nil && status.BindingStatus == "ok" {
			report.Errors = append(report.Errors, entrypointCWDErrors(repoID, status.BindingPath, repoDoc)...)
		}
	}
	scenarios := scenarioIDs(root)
	for _, scenarioID := range scenarios {
		status, err := ScenarioStatus(root, scenarioID)
		if err != nil {
			continue
		}
		for _, repo := range status.Repos {
			switch repo.ScenarioStatus {
			case "drift":
				report.Warnings = append(report.Warnings, fmt.Sprintf("scenario %s repo %s: %s", scenarioID, repo.RepoID, repo.ScenarioReason))
			case "missing":
				report.Warnings = append(report.Warnings, fmt.Sprintf("scenario %s repo %s: %s", scenarioID, repo.RepoID, repo.ScenarioReason))
			case "blocked":
				report.Warnings = append(report.Warnings, fmt.Sprintf("scenario %s repo %s: blocked (%s)", scenarioID, repo.RepoID, repo.ScenarioReason))
			}
		}
	}
	return report
}

func entrypointCWDErrors(repoID string, checkout string, repoDoc model.RepoDocument) []string {
	var errors []string
	for name, entrypoint := range repoDoc.Entrypoints {
		cwd := entrypoint.CWD
		if strings.TrimSpace(cwd) == "" {
			cwd = "."
		}
		if err := validateEntrypointCWD(checkout, cwd); err != nil {
			errors = append(errors, fmt.Sprintf("repo %s entrypoint %s cwd %q is invalid: %v", repoID, name, cwd, err))
		}
	}
	sort.Strings(errors)
	return errors
}

func validateEntrypointCWD(checkout string, cwd string) error {
	if filepath.IsAbs(cwd) {
		return fmt.Errorf("absolute paths are not allowed")
	}
	clean := filepath.Clean(cwd)
	if escapesBase(clean) {
		return fmt.Errorf("path escapes repo checkout")
	}
	dir := filepath.Join(checkout, clean)
	rel, err := filepath.Rel(checkout, dir)
	if err != nil {
		return err
	}
	if escapesBase(rel) {
		return fmt.Errorf("path escapes repo checkout")
	}
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}
	realCheckout, err := filepath.EvalSymlinks(checkout)
	if err != nil {
		return err
	}
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	realRel, err := filepath.Rel(realCheckout, realDir)
	if err != nil {
		return err
	}
	if escapesBase(realRel) {
		return fmt.Errorf("path escapes repo checkout")
	}
	return nil
}

func changeIDs(root string) []string {
	paths, _ := filepath.Glob(filepath.Join(root, "coordination", "changes", "*.yaml"))
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		out = append(out, strings.TrimSuffix(filepath.Base(path), ".yaml"))
	}
	sort.Strings(out)
	return out
}

func scenarioIDs(root string) []string {
	paths, _ := filepath.Glob(filepath.Join(root, "coordination", "scenarios", "*", "manifest.lock.yaml"))
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		out = append(out, filepath.Base(filepath.Dir(path)))
	}
	sort.Strings(out)
	return out
}

func counts(values map[string]int) []Count {
	out := make([]Count, 0, len(values))
	for name, count := range values {
		out = append(out, Count{Name: name, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func boundRepoCount(repoIDs []string, bindings model.BindingsDocument) int {
	count := 0
	for _, repoID := range repoIDs {
		if binding, ok := bindings.Bindings[repoID]; ok && strings.TrimSpace(binding.Path) != "" {
			count++
		}
	}
	return count
}

func fileCount(root string, pattern string) int {
	paths, _ := filepath.Glob(filepath.Join(root, pattern))
	return len(paths)
}

func skillCount(root string) int {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, entry.Name(), "SKILL.md")); err == nil {
			count++
		}
	}
	return count
}

func last(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return items[len(items)-1]
}

func appendUnique(base []string, additions ...string) []string {
	seen := map[string]struct{}{}
	for _, item := range base {
		seen[item] = struct{}{}
	}
	for _, item := range additions {
		if _, ok := seen[item]; ok {
			continue
		}
		base = append(base, item)
		seen[item] = struct{}{}
	}
	return base
}

func escapesBase(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel)
}

func short(ref string) string {
	if len(ref) <= 12 {
		return ref
	}
	return ref[:12]
}
