package scenario

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/buildinfo"
	"github.com/johnkil/polyrepo-workspace-kit/internal/fsutil"
	"github.com/johnkil/polyrepo-workspace-kit/internal/gitstate"
	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

type RunResult struct {
	ReportPath         string
	TextReportPath     string
	MarkdownReportPath string
	Outcomes           []model.ScenarioRunOutcome
	Drift              bool
	Failed             bool
	Blocked            bool
}

func Pin(root string, scenarioID string, changeID string, now time.Time) (string, error) {
	changeDoc, err := workspace.LoadChange(root, changeID)
	if err != nil {
		return "", err
	}
	var repos []model.ScenarioRepo
	var checks []model.ScenarioCheck
	for _, repoID := range changeDoc.Change.Repos {
		checkout, err := workspace.ResolveRepoCheckout(root, repoID)
		if err != nil {
			return "", err
		}
		state, err := gitstate.Capture(checkout)
		if err != nil {
			return "", fmt.Errorf("%s: %w", repoID, err)
		}
		repos = append(repos, model.ScenarioRepo{
			Repo: repoID,
			Revision: model.Revision{
				Commit: state.Commit,
				Short:  state.Short,
				Branch: state.Branch,
			},
			Worktree: model.Worktree{
				Clean:          state.Clean,
				DirtyFiles:     len(state.DirtyPaths),
				UntrackedFiles: len(state.UntrackedPaths),
				DirtyPaths:     state.DirtyPaths,
				UntrackedPaths: state.UntrackedPaths,
			},
			DependencyHints: model.DependencyHints{Lockfiles: state.Lockfiles},
		})
		repoDoc, err := workspace.LoadRepo(root, repoID)
		if err != nil {
			return "", err
		}
		check, ok := derivedCheck(repoID, repoDoc)
		if ok {
			if _, err := parseCommand(check.Run); err != nil {
				return "", fmt.Errorf("%s entrypoint %q: %w", repoID, check.ID, err)
			}
			checks = append(checks, check)
		}
	}
	if len(checks) == 0 {
		return "", fmt.Errorf("could not derive any scenario checks from repo entrypoints")
	}
	wkitVersion := buildinfo.Current().Version
	doc := model.ScenarioLockDocument{
		Version: 1,
		Scenario: model.ScenarioMeta{
			ID:          scenarioID,
			Change:      changeDoc.Change.ID,
			Context:     changeDoc.Change.Context,
			GeneratedAt: now.UTC().Format(time.RFC3339),
			GeneratedBy: model.GeneratedBy{Tool: "wkit", Version: wkitVersion},
			Semantics:   "reviewable-local-validation-snapshot",
			Notes: []string{
				"v0.x scenarios pin revisions and local checks but do not guarantee full environment replay.",
			},
		},
		ToolVersions: model.ToolVersions{
			WKit:  wkitVersion,
			Git:   gitstate.Version(),
			Extra: map[string]string{},
		},
		Repos:  repos,
		Checks: checks,
	}
	path, err := workspace.ScenarioPath(root, scenarioID)
	if err != nil {
		return "", err
	}
	if err := manifest.WriteYAML(path, doc); err != nil {
		return "", err
	}
	return path, nil
}

func Load(root string, scenarioID string) (model.ScenarioLockDocument, error) {
	var doc model.ScenarioLockDocument
	path, err := workspace.ScenarioPath(root, scenarioID)
	if err != nil {
		return doc, err
	}
	err = manifest.LoadYAML(path, &doc)
	return doc, err
}

func Run(root string, scenarioID string, now time.Time) (RunResult, error) {
	lock, err := Load(root, scenarioID)
	if err != nil {
		return RunResult{}, err
	}
	reportDir := filepath.Join(root, "local", "reports", scenarioID)
	stamp := now.UTC().Format("20060102T150405Z")
	runID, logRelDir, err := reserveRunPaths(reportDir, stamp)
	if err != nil {
		return RunResult{}, err
	}

	pinned := map[string]model.ScenarioRepo{}
	for _, repo := range lock.Repos {
		pinned[repo.Repo] = repo
	}

	var result RunResult
	for _, check := range lock.Checks {
		outcome := runCheck(root, pinned, check, reportDir, logRelDir)
		result.Outcomes = append(result.Outcomes, outcome)
		switch outcome.Status {
		case "failed":
			result.Failed = true
		case "blocked":
			result.Blocked = true
			if strings.Contains(outcome.Reason, "pinned ref drift") {
				result.Drift = true
			}
		}
	}

	report := model.ScenarioReportDocument{
		Version: 1,
		Report: model.ScenarioReportMeta{
			Scenario:    scenarioID,
			GeneratedAt: now.UTC().Format(time.RFC3339),
			ReportKind:  "local-validation-run",
		},
		Results: result.Outcomes,
	}
	result.ReportPath = filepath.Join(reportDir, runID+".yaml")
	if err := manifest.WriteYAMLExclusive(result.ReportPath, report); err != nil {
		return RunResult{}, err
	}
	result.TextReportPath = filepath.Join(reportDir, runID+".txt")
	if err := fsutil.WriteFileExclusive(result.TextReportPath, []byte(TextReport(scenarioID, result.Outcomes))); err != nil {
		return RunResult{}, err
	}
	result.MarkdownReportPath = filepath.Join(reportDir, runID+".md")
	if err := fsutil.WriteFileExclusive(result.MarkdownReportPath, []byte(MarkdownReport(scenarioID, result.Outcomes, reportDir))); err != nil {
		return RunResult{}, err
	}
	return result, nil
}

func reserveRunPaths(reportDir string, stamp string) (string, string, error) {
	logsRoot := filepath.Join(reportDir, "logs")
	if err := fsutil.EnsureDir(logsRoot); err != nil {
		return "", "", err
	}
	for index := 0; ; index++ {
		runID := stamp
		if index > 0 {
			runID = fmt.Sprintf("%s.%03d", stamp, index)
		}
		if fsutil.Exists(filepath.Join(reportDir, runID+".yaml")) || fsutil.Exists(filepath.Join(reportDir, runID+".txt")) || fsutil.Exists(filepath.Join(reportDir, runID+".md")) {
			continue
		}
		logDir := filepath.Join(logsRoot, runID)
		if err := os.Mkdir(logDir, 0o755); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", "", err
		}
		return runID, filepath.Join("logs", runID), nil
	}
}

func TextReport(scenarioID string, outcomes []model.ScenarioRunOutcome) string {
	counts := map[string]int{}
	for _, outcome := range outcomes {
		counts[outcome.Status]++
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Scenario: %s\n", scenarioID)
	fmt.Fprintf(&b, "Results: passed=%d failed=%d blocked=%d skipped=%d\n\n", counts["passed"], counts["failed"], counts["blocked"], counts["skipped"])
	for _, outcome := range outcomes {
		fmt.Fprintf(&b, "- %s: %s", outcome.Status, outcome.Check)
		if outcome.Reason != "" {
			fmt.Fprintf(&b, " (%s)", outcome.Reason)
		}
		b.WriteString("\n")
		if outcome.StdoutPath != nil {
			fmt.Fprintf(&b, "  stdout: %s\n", *outcome.StdoutPath)
		}
		if outcome.StderrPath != nil {
			fmt.Fprintf(&b, "  stderr: %s\n", *outcome.StderrPath)
		}
		if outcome.EnvProfile != "" {
			fmt.Fprintf(&b, "  env_profile: %s\n", outcome.EnvProfile)
		}
	}
	return fsutil.NormalizeText(b.String())
}

func MarkdownReport(scenarioID string, outcomes []model.ScenarioRunOutcome, reportDir string) string {
	counts := map[string]int{}
	for _, outcome := range outcomes {
		counts[outcome.Status]++
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Scenario Report: %s\n\n", markdownText(scenarioID))
	fmt.Fprintf(&b, "Results: passed=%d failed=%d blocked=%d skipped=%d\n\n", counts["passed"], counts["failed"], counts["blocked"], counts["skipped"])
	b.WriteString("| check | status | reason | logs |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, outcome := range outcomes {
		fmt.Fprintf(&b, "| `%s` | `%s` | %s | %s |\n", markdownCell(outcome.Check), markdownCell(outcome.Status), markdownCell(valueOrDash(outcome.Reason)), markdownLogs(outcome))
	}
	b.WriteString("\n")

	diagnostics := markdownDiagnostics(outcomes, reportDir)
	if diagnostics != "" {
		b.WriteString("## Diagnostics\n\n")
		b.WriteString(diagnostics)
	}
	b.WriteString("Generated by `wkit scenario run`. This is derived local evidence, not canonical workspace state.\n")
	return fsutil.NormalizeText(b.String())
}

func markdownDiagnostics(outcomes []model.ScenarioRunOutcome, reportDir string) string {
	var b strings.Builder
	for _, outcome := range outcomes {
		if outcome.Status != "failed" || outcome.StderrPath == nil || *outcome.StderrPath == "" {
			continue
		}
		excerpt := logExcerpt(filepath.Join(reportDir, *outcome.StderrPath), 3)
		if excerpt == "" {
			continue
		}
		fmt.Fprintf(&b, "### %s\n\n", markdownText(outcome.Check))
		b.WriteString("```text\n")
		b.WriteString(excerpt)
		if !strings.HasSuffix(excerpt, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("```\n\n")
	}
	return b.String()
}

func logExcerpt(path string, maxLines int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	out := make([]string, 0, maxLines)
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) > 200 {
			line = line[:200] + "..."
		}
		line = strings.ReplaceAll(line, "```", "'''")
		out = append(out, line)
		if len(out) == maxLines {
			break
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n")
}

func markdownLogs(outcome model.ScenarioRunOutcome) string {
	logs := []string{}
	if outcome.StdoutPath != nil && *outcome.StdoutPath != "" {
		logs = append(logs, "`"+markdownCell(*outcome.StdoutPath)+"`")
	}
	if outcome.StderrPath != nil && *outcome.StderrPath != "" {
		logs = append(logs, "`"+markdownCell(*outcome.StderrPath)+"`")
	}
	if len(logs) == 0 {
		return "-"
	}
	return strings.Join(logs, "<br>")
}

func markdownText(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	return strings.TrimSpace(value)
}

func markdownCell(value string) string {
	value = markdownText(value)
	value = strings.ReplaceAll(value, "|", `\|`)
	if value == "" {
		return "-"
	}
	return value
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func derivedCheck(repoID string, repoDoc model.RepoDocument) (model.ScenarioCheck, bool) {
	if len(repoDoc.Entrypoints) == 0 {
		return model.ScenarioCheck{}, false
	}
	name := "test"
	entrypoint, ok := repoDoc.Entrypoints[name]
	if !ok {
		names := make([]string, 0, len(repoDoc.Entrypoints))
		for candidate := range repoDoc.Entrypoints {
			names = append(names, candidate)
		}
		sort.Strings(names)
		name = names[0]
		entrypoint = repoDoc.Entrypoints[name]
	}
	if entrypoint.CWD == "" {
		entrypoint.CWD = "."
	}
	return model.ScenarioCheck{
		ID:                    repoID + ":" + name,
		Repo:                  repoID,
		CWD:                   entrypoint.CWD,
		Run:                   entrypoint.Run,
		TimeoutSeconds:        entrypoint.TimeoutSeconds,
		EnvProfile:            entrypoint.EnvProfile,
		EnvRequirements:       entrypoint.EnvRequirements,
		ExpectedArtifacts:     entrypoint.ExpectedArtifacts,
		RequiresCleanWorktree: entrypoint.RequiresCleanTree,
		Status:                "planned",
	}, true
}

func runCheck(root string, pinned map[string]model.ScenarioRepo, check model.ScenarioCheck, reportDir string, logRelDir string) model.ScenarioRunOutcome {
	start := time.Now()
	outcome := model.ScenarioRunOutcome{
		Check:      check.ID,
		Status:     "blocked",
		EnvProfile: check.EnvProfile,
		Artifacts:  []string{},
	}
	finish := func(status string, reason string) model.ScenarioRunOutcome {
		outcome.Status = status
		outcome.Reason = reason
		outcome.DurationSeconds = time.Since(start).Seconds()
		return outcome
	}
	checkout, err := workspace.ResolveRepoCheckout(root, check.Repo)
	if err != nil {
		return finish("blocked", err.Error())
	}
	current, err := gitstate.Head(checkout)
	if err != nil {
		return finish("blocked", err.Error())
	}
	pinnedRepo, ok := pinned[check.Repo]
	if !ok {
		return finish("blocked", fmt.Sprintf("scenario lock has no pinned repo entry for %s", check.Repo))
	}
	if pinnedRepo.Revision.Commit != "" && current != pinnedRepo.Revision.Commit {
		return finish("blocked", fmt.Sprintf("pinned ref drift: current HEAD %s does not match scenario lock %s", short(current), short(pinnedRepo.Revision.Commit)))
	}
	if check.RequiresCleanWorktree {
		state, err := gitstate.Capture(checkout)
		if err != nil {
			return finish("blocked", err.Error())
		}
		if !state.Clean {
			return finish("blocked", dirtyReason(state))
		}
	}
	if strings.TrimSpace(check.Run) == "" {
		return finish("blocked", "missing command: check.run is empty")
	}
	parts, err := parseCommand(check.Run)
	if err != nil {
		return finish("blocked", "unsupported command: "+err.Error())
	}

	cmdDir, err := resolveCheckCWD(checkout, check.CWD)
	if err != nil {
		return finish("blocked", err.Error())
	}

	stdout, stderr, runErr := execute(cmdDir, check, parts)
	stdoutRel, stderrRel := logPaths(logRelDir, check.ID)
	stdoutAbs := filepath.Join(reportDir, stdoutRel)
	stderrAbs := filepath.Join(reportDir, stderrRel)
	if err := fsutil.WriteFileAtomic(stdoutAbs, stdout); err != nil {
		return finish("failed", "failed to write stdout log: "+err.Error())
	}
	if err := fsutil.WriteFileAtomic(stderrAbs, stderr); err != nil {
		return finish("failed", "failed to write stderr log: "+err.Error())
	}
	outcome.StdoutPath = &stdoutRel
	outcome.StderrPath = &stderrRel
	if runErr != nil {
		return finish("failed", runErr.Error())
	}
	return finish("passed", "")
}

func dirtyReason(state gitstate.State) string {
	parts := []string{
		fmt.Sprintf("dirty files: %d", len(state.DirtyPaths)),
		fmt.Sprintf("untracked files: %d", len(state.UntrackedPaths)),
	}
	if len(state.DirtyPaths) > 0 {
		parts = append(parts, "dirty paths: "+strings.Join(state.DirtyPaths, ", "))
	}
	if len(state.UntrackedPaths) > 0 {
		parts = append(parts, "untracked paths: "+strings.Join(state.UntrackedPaths, ", "))
	}
	return "worktree is not clean (" + strings.Join(parts, "; ") + ")"
}

func resolveCheckCWD(checkout string, cwd string) (string, error) {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}
	if filepath.IsAbs(cwd) {
		return "", fmt.Errorf("unsafe cwd %q: absolute paths are not allowed", cwd)
	}
	clean := filepath.Clean(cwd)
	if escapesBase(clean) {
		return "", fmt.Errorf("unsafe cwd %q: escapes repo checkout", cwd)
	}
	dir := filepath.Join(checkout, clean)
	rel, err := filepath.Rel(checkout, dir)
	if err != nil {
		return "", err
	}
	if escapesBase(rel) {
		return "", fmt.Errorf("unsafe cwd %q: escapes repo checkout", cwd)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("invalid cwd %q: %w", cwd, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("invalid cwd %q: not a directory", cwd)
	}
	realCheckout, err := filepath.EvalSymlinks(checkout)
	if err != nil {
		return "", fmt.Errorf("invalid checkout %q: %w", checkout, err)
	}
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", fmt.Errorf("invalid cwd %q: %w", cwd, err)
	}
	realRel, err := filepath.Rel(realCheckout, realDir)
	if err != nil {
		return "", err
	}
	if escapesBase(realRel) {
		return "", fmt.Errorf("unsafe cwd %q: escapes repo checkout", cwd)
	}
	return dir, nil
}

func escapesBase(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel)
}

func execute(cmdDir string, check model.ScenarioCheck, parts []string) ([]byte, []byte, error) {
	ctx := context.Background()
	cancel := func() {}
	if check.TimeoutSeconds > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(check.TimeoutSeconds)*time.Second)
	}
	defer cancel()
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = cmdDir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("command timed out after %d seconds", check.TimeoutSeconds)
	}
	return stdout.Bytes(), stderr.Bytes(), err
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

func logPaths(logRelDir string, checkID string) (string, string) {
	safe := strings.NewReplacer("/", "-", "\\", "-", ":", "-", " ", "-").Replace(checkID)
	return filepath.Join(logRelDir, safe+".stdout.txt"), filepath.Join(logRelDir, safe+".stderr.txt")
}

func short(ref string) string {
	if len(ref) <= 12 {
		return ref
	}
	return ref[:12]
}
