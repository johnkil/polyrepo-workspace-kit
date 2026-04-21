package handoff

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/orient"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

type Options struct {
	ScenarioID string
}

type selectedScenario struct {
	ID   string
	Lock model.ScenarioLockDocument
}

type latestReport struct {
	Path         string
	TextPath     string
	MarkdownPath string
	Report       model.ScenarioReportDocument
}

func Markdown(root string, changeID string, opts Options) (string, error) {
	changeDoc, err := workspace.LoadChange(root, changeID)
	if err != nil {
		return "", err
	}
	contexts, err := workspace.LoadContexts(root)
	if err != nil {
		return "", err
	}
	bindings, err := workspace.LoadBindings(root)
	if err != nil {
		return "", err
	}

	selected, err := selectScenario(root, changeID, opts.ScenarioID)
	if err != nil {
		return "", err
	}

	var status orient.ScenarioStatusReport
	hasStatus := false
	if selected != nil {
		status, err = orient.ScenarioStatus(root, selected.ID)
		if err != nil {
			return "", err
		}
		hasStatus = true
	}

	var report *latestReport
	if selected != nil {
		report, err = loadLatestReport(root, selected.ID)
		if err != nil {
			return "", err
		}
	}

	change := changeDoc.Change
	var b strings.Builder
	fmt.Fprintf(&b, "# Handoff: %s\n\n", markdownText(change.Title))
	fmt.Fprintf(&b, "- change: `%s`\n", change.ID)
	fmt.Fprintf(&b, "- kind: `%s`\n", valueOrDash(change.Kind))
	fmt.Fprintf(&b, "- context: `%s`\n", valueOrDash(change.Context))
	fmt.Fprintf(&b, "- repos: %d\n", len(change.Repos))
	if selected != nil {
		fmt.Fprintf(&b, "- scenario: `%s`\n", selected.ID)
	} else {
		fmt.Fprintf(&b, "- scenario: none found for this change\n")
	}
	if report != nil {
		fmt.Fprintf(&b, "- latest report: `%s`\n", rel(root, report.Path))
		if report.TextPath != "" {
			fmt.Fprintf(&b, "- latest text report: `%s`\n", rel(root, report.TextPath))
		}
		if report.MarkdownPath != "" {
			fmt.Fprintf(&b, "- latest markdown report: `%s`\n", rel(root, report.MarkdownPath))
		}
	}
	b.WriteString("\n")

	writeRepoSection(&b, change, contexts, bindings, status, hasStatus)
	writeScenarioSection(&b, root, selected, report)
	b.WriteString("Generated from local `wkit` state. This is a derived handoff artifact, not canonical workspace state.\n")
	return b.String(), nil
}

func writeRepoSection(b *strings.Builder, change model.Change, contexts model.ContextsDocument, bindings model.BindingsDocument, status orient.ScenarioStatusReport, hasStatus bool) {
	b.WriteString("## Repositories\n\n")
	if context, ok := contexts.Contexts[change.Context]; ok {
		fmt.Fprintf(b, "Context `%s` contains: %s\n\n", change.Context, inlineCodeList(context.Repos))
	}
	b.WriteString("| repo | binding | scenario | pinned | current | branch |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	statusByRepo := map[string]orient.RepoStatus{}
	if hasStatus {
		for _, repo := range status.Repos {
			statusByRepo[repo.RepoID] = repo
		}
	}
	for _, repoID := range change.Repos {
		binding := "unbound"
		if item, ok := bindings.Bindings[repoID]; ok && strings.TrimSpace(item.Path) != "" {
			binding = item.Path
		}
		scenarioStatus := "-"
		pinned := "-"
		current := "-"
		branch := "-"
		if repo, ok := statusByRepo[repoID]; ok {
			scenarioStatus = valueOrDash(repo.ScenarioStatus)
			pinned = valueOrDash(repo.PinnedCommit)
			current = valueOrDash(repo.CurrentCommit)
			branch = branchLabel(repo)
		}
		fmt.Fprintf(b, "| `%s` | %s | %s | `%s` | `%s` | `%s` |\n", markdownCell(repoID), markdownCell(binding), markdownCell(scenarioStatus), markdownCell(pinned), markdownCell(current), markdownCell(branch))
	}
	b.WriteString("\n")
}

func writeScenarioSection(b *strings.Builder, root string, selected *selectedScenario, report *latestReport) {
	b.WriteString("## Scenario Evidence\n\n")
	if selected == nil {
		b.WriteString("No scenario lock was found for this change yet.\n\n")
		b.WriteString("Next step: `wkit scenario pin <scenario-id> --change <change-id>`.\n\n")
		return
	}
	lock := selected.Lock
	fmt.Fprintf(b, "- scenario: `%s`\n", lock.Scenario.ID)
	fmt.Fprintf(b, "- generated_at: `%s`\n", valueOrDash(lock.Scenario.GeneratedAt))
	fmt.Fprintf(b, "- semantics: `%s`\n", valueOrDash(lock.Scenario.Semantics))
	fmt.Fprintf(b, "- checks: %d\n", len(lock.Checks))
	if report == nil {
		b.WriteString("- latest report: none\n\n")
		b.WriteString("Next step: `wkit scenario run " + lock.Scenario.ID + "`.\n\n")
		return
	}
	fmt.Fprintf(b, "- latest report: `%s`\n\n", rel(root, report.Path))
	writeResultsSection(b, root, report)
}

func writeResultsSection(b *strings.Builder, root string, report *latestReport) {
	counts := map[string]int{}
	for _, result := range report.Report.Results {
		counts[result.Status]++
	}
	fmt.Fprintf(b, "Results: passed=%d failed=%d blocked=%d skipped=%d\n\n", counts["passed"], counts["failed"], counts["blocked"], counts["skipped"])
	b.WriteString("| check | status | reason | logs |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	reportDir := filepath.Dir(report.Path)
	for _, result := range report.Report.Results {
		logs := []string{}
		if result.StdoutPath != nil && *result.StdoutPath != "" {
			logs = append(logs, "`"+markdownCell(rel(root, filepath.Join(reportDir, *result.StdoutPath)))+"`")
		}
		if result.StderrPath != nil && *result.StderrPath != "" {
			logs = append(logs, "`"+markdownCell(rel(root, filepath.Join(reportDir, *result.StderrPath)))+"`")
		}
		if len(logs) == 0 {
			logs = append(logs, "-")
		}
		fmt.Fprintf(b, "| `%s` | `%s` | %s | %s |\n", markdownCell(result.Check), markdownCell(result.Status), markdownCell(valueOrDash(result.Reason)), strings.Join(logs, "<br>"))
	}
	b.WriteString("\n")
}

func selectScenario(root string, changeID string, requested string) (*selectedScenario, error) {
	if requested != "" {
		lock, err := scenario.Load(root, requested)
		if err != nil {
			return nil, err
		}
		if lock.Scenario.Change != changeID {
			return nil, fmt.Errorf("scenario %q references change %q, not %q", requested, lock.Scenario.Change, changeID)
		}
		return &selectedScenario{ID: requested, Lock: lock}, nil
	}
	paths, _ := filepath.Glob(filepath.Join(root, "coordination", "scenarios", "*", "manifest.lock.yaml"))
	matches := []selectedScenario{}
	for _, path := range paths {
		scenarioID := filepath.Base(filepath.Dir(path))
		lock, err := scenario.Load(root, scenarioID)
		if err != nil {
			return nil, err
		}
		if lock.Scenario.Change == changeID {
			matches = append(matches, selectedScenario{ID: scenarioID, Lock: lock})
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}
	sort.Slice(matches, func(i, j int) bool {
		left := matches[i].Lock.Scenario.GeneratedAt
		right := matches[j].Lock.Scenario.GeneratedAt
		if left == right {
			return matches[i].ID < matches[j].ID
		}
		return left < right
	})
	return &matches[len(matches)-1], nil
}

func loadLatestReport(root string, scenarioID string) (*latestReport, error) {
	paths, _ := filepath.Glob(filepath.Join(root, "local", "reports", scenarioID, "*.yaml"))
	if len(paths) == 0 {
		return nil, nil
	}
	sort.Strings(paths)
	path := paths[len(paths)-1]
	var report model.ScenarioReportDocument
	if err := manifest.LoadYAML(path, &report); err != nil {
		return nil, err
	}
	out := &latestReport{Path: path, Report: report}
	textPath := strings.TrimSuffix(path, ".yaml") + ".txt"
	if exists(textPath) {
		out.TextPath = textPath
	}
	markdownPath := strings.TrimSuffix(path, ".yaml") + ".md"
	if exists(markdownPath) {
		out.MarkdownPath = markdownPath
	}
	return out, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func inlineCodeList(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, "`"+markdownText(item)+"`")
	}
	return strings.Join(out, ", ")
}

func branchLabel(repo orient.RepoStatus) string {
	if repo.Detached {
		return "detached"
	}
	return valueOrDash(repo.Branch)
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func rel(root string, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
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
