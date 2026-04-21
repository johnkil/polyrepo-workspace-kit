package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

type Report struct {
	Errors   []string
	Warnings []string
}

func (r Report) OK() bool {
	return len(r.Errors) == 0
}

var allowedRuleKinds = map[string]struct{}{
	"rollout-order": {},
}

var allowedCheckStatuses = map[string]struct{}{
	"planned": {},
	"passed":  {},
	"failed":  {},
	"skipped": {},
	"blocked": {},
}

var allowedRolloutOrders = map[string]struct{}{
	"provider-before-consumer": {},
	"consumer-after-provider":  {},
}

func Workspace(root string) Report {
	report := Report{}
	workspacePath := filepath.Join(root, workspace.WorkspaceFile)
	contextsPath := filepath.Join(root, workspace.ContextsFile)

	doc, err := loadWorkspace(workspacePath, &report)
	if err != nil {
		return report
	}
	contexts := loadContexts(contextsPath, &report)
	repoIDs := workspace.RepoIDs(doc)

	validateWorkspaceShape(doc, &report)
	validateRepoManifests(root, doc, &report)
	validateRelations(doc, repoIDs, &report)
	validateContexts(contexts, repoIDs, &report)
	validateRules(root, doc, repoIDs, contexts, &report)
	validateBindings(root, repoIDs, &report)
	validateChanges(root, repoIDs, contexts, &report)
	validateScenarios(root, repoIDs, contexts, &report)
	validateReports(root, &report)

	return report
}

func loadWorkspace(path string, report *Report) (model.WorkspaceDocument, error) {
	var doc model.WorkspaceDocument
	if err := manifest.LoadYAML(path, &doc); err != nil {
		if manifest.IsMissing(err) {
			report.Errors = append(report.Errors, "missing coordination/workspace.yaml")
		} else {
			report.Errors = append(report.Errors, err.Error())
		}
		return doc, err
	}
	return doc, nil
}

func loadContexts(path string, report *Report) model.ContextsDocument {
	var doc model.ContextsDocument
	if err := manifest.LoadYAML(path, &doc); err != nil {
		if manifest.IsMissing(err) {
			report.Errors = append(report.Errors, "missing coordination/contexts.yaml")
		} else {
			report.Errors = append(report.Errors, err.Error())
		}
	}
	if doc.Contexts == nil {
		doc.Contexts = map[string]model.Context{}
	}
	return doc
}

func validateWorkspaceShape(doc model.WorkspaceDocument, report *Report) {
	if doc.Workspace.ID == "" {
		report.Errors = append(report.Errors, "coordination/workspace.yaml: workspace.id is required")
	}
	if doc.Workspace.Model == "" {
		report.Errors = append(report.Errors, "coordination/workspace.yaml: workspace.model is required")
	}
	seen := map[string]struct{}{}
	for _, repoID := range doc.Repos {
		if err := workspace.ValidateID("repo id", repoID); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("coordination/workspace.yaml: invalid repo id %q: %v", repoID, err))
			continue
		}
		if _, ok := seen[repoID]; ok {
			report.Errors = append(report.Errors, fmt.Sprintf("coordination/workspace.yaml: duplicate repo %q", repoID))
		}
		seen[repoID] = struct{}{}
	}
}

func validateRepoManifests(root string, doc model.WorkspaceDocument, report *Report) {
	for _, repoID := range doc.Repos {
		path, err := workspace.RepoManifestPath(root, repoID)
		if err != nil {
			continue
		}
		var repo model.RepoDocument
		if err := manifest.LoadYAML(path, &repo); err != nil {
			if manifest.IsMissing(err) {
				report.Errors = append(report.Errors, fmt.Sprintf("missing repo manifest: repos/%s/repo.yaml", repoID))
			} else {
				report.Errors = append(report.Errors, err.Error())
			}
			continue
		}
		if repo.Repo.ID != repoID {
			report.Errors = append(report.Errors, fmt.Sprintf("repos/%s/repo.yaml: repo.id is %q", repoID, repo.Repo.ID))
		}
		if repo.Repo.Kind == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("repos/%s/repo.yaml: repo.kind is required", repoID))
		}
		if len(repo.Entrypoints) == 0 {
			report.Warnings = append(report.Warnings, fmt.Sprintf("repos/%s/repo.yaml: no entrypoints declared", repoID))
		} else if _, ok := repo.Entrypoints["test"]; !ok {
			report.Warnings = append(report.Warnings, fmt.Sprintf("repos/%s/repo.yaml: no test entrypoint declared", repoID))
		}
		for name, entrypoint := range repo.Entrypoints {
			if entrypoint.Run == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("repos/%s/repo.yaml: entrypoints.%s.run is required", repoID, name))
			} else if err := validateCommandRun(entrypoint.Run); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("repos/%s/repo.yaml: entrypoints.%s.run is unsupported: %v", repoID, name, err))
			}
		}
	}
}

func validateRelations(doc model.WorkspaceDocument, repoIDs map[string]struct{}, report *Report) {
	for _, relation := range doc.Relations {
		if _, ok := repoIDs[relation.From]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("relation source %q is not declared in workspace.repos", relation.From))
		}
		if _, ok := repoIDs[relation.To]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("relation target %q is not declared in workspace.repos", relation.To))
		}
		if !model.IsRelationKind(relation.Kind) {
			report.Errors = append(report.Errors, fmt.Sprintf("relation %s -> %s uses unsupported kind %q", relation.From, relation.To, relation.Kind))
		}
	}
}

func validateContexts(contexts model.ContextsDocument, repoIDs map[string]struct{}, report *Report) {
	for name, context := range contexts.Contexts {
		if len(context.Repos) == 0 {
			report.Errors = append(report.Errors, fmt.Sprintf("context %q must include at least one repo", name))
		}
		for _, repoID := range context.Repos {
			if _, ok := repoIDs[repoID]; !ok {
				report.Errors = append(report.Errors, fmt.Sprintf("context %q references unknown repo %q", name, repoID))
			}
		}
	}
}

func validateRules(root string, doc model.WorkspaceDocument, repoIDs map[string]struct{}, contexts model.ContextsDocument, report *Report) {
	referenced := map[string]struct{}{}
	for _, ruleID := range doc.Rules {
		if err := workspace.ValidateID("rule id", ruleID); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("coordination/workspace.yaml: invalid rule id %q: %v", ruleID, err))
			continue
		}
		referenced[ruleID] = struct{}{}
		path := filepath.Join(root, "coordination", "rules", ruleID+".yaml")
		var ruleDoc model.RuleDocument
		if err := manifest.LoadYAML(path, &ruleDoc); err != nil {
			if manifest.IsMissing(err) {
				report.Errors = append(report.Errors, fmt.Sprintf("workspace rule %q has no coordination/rules/%s.yaml file", ruleID, ruleID))
			} else {
				report.Errors = append(report.Errors, err.Error())
			}
			continue
		}
		validateRule(ruleID, ruleDoc.Rule, repoIDs, contexts, report)
	}
	paths, _ := filepath.Glob(filepath.Join(root, "coordination", "rules", "*.yaml"))
	for _, path := range paths {
		fileID := strings.TrimSuffix(filepath.Base(path), ".yaml")
		if _, ok := referenced[fileID]; ok {
			continue
		}
		rel, _ := filepath.Rel(root, path)
		if err := workspace.ValidateID("rule id", fileID); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s has invalid rule file name %q: %v", rel, fileID, err))
			continue
		}
		var ruleDoc model.RuleDocument
		if err := manifest.LoadYAML(path, &ruleDoc); err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}
		report.Warnings = append(report.Warnings, fmt.Sprintf("%s is not referenced by coordination/workspace.yaml rules", rel))
		validateRule(fileID, ruleDoc.Rule, repoIDs, contexts, report)
	}
}

func validateRule(expectedID string, rule model.Rule, repoIDs map[string]struct{}, contexts model.ContextsDocument, report *Report) {
	if rule.ID != expectedID {
		report.Errors = append(report.Errors, fmt.Sprintf("coordination/rules/%s.yaml: rule.id is %q", expectedID, rule.ID))
	}
	if _, ok := allowedRuleKinds[rule.Kind]; !ok {
		report.Errors = append(report.Errors, fmt.Sprintf("coordination/rules/%s.yaml: unsupported rule kind %q", expectedID, rule.Kind))
	}
	if rule.AppliesTo.RelationKind != "" {
		if !model.IsRelationKind(rule.AppliesTo.RelationKind) {
			report.Errors = append(report.Errors, fmt.Sprintf("coordination/rules/%s.yaml: unsupported applies_to.relation_kind %q", expectedID, rule.AppliesTo.RelationKind))
		}
	}
	if rule.AppliesTo.FromRepo != "" {
		if _, ok := repoIDs[rule.AppliesTo.FromRepo]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("coordination/rules/%s.yaml: unknown applies_to.from_repo %q", expectedID, rule.AppliesTo.FromRepo))
		}
	}
	if rule.AppliesTo.ToRepo != "" {
		if _, ok := repoIDs[rule.AppliesTo.ToRepo]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("coordination/rules/%s.yaml: unknown applies_to.to_repo %q", expectedID, rule.AppliesTo.ToRepo))
		}
	}
	if rule.AppliesTo.Context != "" {
		if _, ok := contexts.Contexts[rule.AppliesTo.Context]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("coordination/rules/%s.yaml: unknown applies_to.context %q", expectedID, rule.AppliesTo.Context))
		}
	}
	if rule.Kind == "rollout-order" {
		if _, ok := allowedRolloutOrders[rule.Policy.Order]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("coordination/rules/%s.yaml: unsupported policy.order %q", expectedID, rule.Policy.Order))
		}
	}
}

func validateBindings(root string, repoIDs map[string]struct{}, report *Report) {
	var bindings model.BindingsDocument
	path := filepath.Join(root, workspace.BindingsFile)
	if err := manifest.LoadYAML(path, &bindings); err != nil {
		if manifest.IsMissing(err) {
			report.Warnings = append(report.Warnings, "missing local/bindings.yaml")
		} else {
			report.Errors = append(report.Errors, err.Error())
		}
		return
	}
	for repoID, binding := range bindings.Bindings {
		if _, ok := repoIDs[repoID]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("local/bindings.yaml: binding references unknown repo %q", repoID))
		}
		if binding.Path == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("local/bindings.yaml: binding for %q has empty path", repoID))
			continue
		}
		if _, err := os.Stat(binding.Path); err != nil {
			report.Warnings = append(report.Warnings, fmt.Sprintf("binding path for %q does not exist: %s", repoID, binding.Path))
		}
	}
	for repoID := range repoIDs {
		if _, ok := bindings.Bindings[repoID]; !ok {
			report.Warnings = append(report.Warnings, fmt.Sprintf("local/bindings.yaml: no local binding for repo %q", repoID))
		}
	}
}

func validateChanges(root string, repoIDs map[string]struct{}, contexts model.ContextsDocument, report *Report) {
	paths, _ := filepath.Glob(filepath.Join(root, "coordination", "changes", "*.yaml"))
	for _, path := range paths {
		var changeDoc model.ChangeDocument
		if err := manifest.LoadYAML(path, &changeDoc); err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}
		rel, _ := filepath.Rel(root, path)
		if changeDoc.Change.ID == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: change.id is required", rel))
		} else {
			if err := workspace.ValidateID("change id", changeDoc.Change.ID); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s has invalid change.id %q: %v", rel, changeDoc.Change.ID, err))
			}
			if fileID := strings.TrimSuffix(filepath.Base(path), ".yaml"); changeDoc.Change.ID != fileID {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: change.id %q does not match file name %q", rel, changeDoc.Change.ID, fileID))
			}
		}
		if _, ok := contexts.Contexts[changeDoc.Change.Context]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("%s references unknown context %q", rel, changeDoc.Change.Context))
		}
		for _, repoID := range changeDoc.Change.Repos {
			if _, ok := repoIDs[repoID]; !ok {
				report.Errors = append(report.Errors, fmt.Sprintf("%s references unknown repo %q", rel, repoID))
			}
		}
	}
}

func validateScenarios(root string, repoIDs map[string]struct{}, contexts model.ContextsDocument, report *Report) {
	paths, _ := filepath.Glob(filepath.Join(root, "coordination", "scenarios", "*", "manifest.lock.yaml"))
	for _, path := range paths {
		var scenario model.ScenarioLockDocument
		if err := manifest.LoadYAML(path, &scenario); err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}
		rel, _ := filepath.Rel(root, path)
		scenarioID := filepath.Base(filepath.Dir(path))
		if err := workspace.ValidateID("scenario id", scenarioID); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s has invalid scenario directory %q: %v", rel, scenarioID, err))
		}
		if scenario.Scenario.ID == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: scenario.id is required", rel))
		} else {
			if err := workspace.ValidateID("scenario id", scenario.Scenario.ID); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s has invalid scenario.id %q: %v", rel, scenario.Scenario.ID, err))
			}
			if scenario.Scenario.ID != scenarioID {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: scenario.id %q does not match scenario directory %q", rel, scenario.Scenario.ID, scenarioID))
			}
		}
		if scenario.Scenario.Change == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: scenario.change is required", rel))
		} else if err := workspace.ValidateID("change id", scenario.Scenario.Change); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s has invalid scenario.change %q: %v", rel, scenario.Scenario.Change, err))
		} else {
			changePath, err := workspace.ChangePath(root, scenario.Scenario.Change)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s has invalid scenario.change %q: %v", rel, scenario.Scenario.Change, err))
			} else if _, err := os.Stat(changePath); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s references missing change %q", rel, scenario.Scenario.Change))
			}
		}
		if _, ok := contexts.Contexts[scenario.Scenario.Context]; !ok {
			report.Errors = append(report.Errors, fmt.Sprintf("%s references unknown context %q", rel, scenario.Scenario.Context))
		}
		pinnedRepos := map[string]struct{}{}
		for _, repo := range scenario.Repos {
			if err := workspace.ValidateID("repo id", repo.Repo); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s has invalid pinned repo %q: %v", rel, repo.Repo, err))
				continue
			}
			if _, ok := repoIDs[repo.Repo]; !ok {
				report.Errors = append(report.Errors, fmt.Sprintf("%s references unknown repo in repos: %q", rel, repo.Repo))
			}
			if _, ok := pinnedRepos[repo.Repo]; ok {
				report.Errors = append(report.Errors, fmt.Sprintf("%s has duplicate pinned repo %q", rel, repo.Repo))
			}
			pinnedRepos[repo.Repo] = struct{}{}
		}
		checkIDs := map[string]struct{}{}
		for _, check := range scenario.Checks {
			if check.ID == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s has check with empty id", rel))
			} else if _, ok := checkIDs[check.ID]; ok {
				report.Errors = append(report.Errors, fmt.Sprintf("%s has duplicate check id %q", rel, check.ID))
			}
			checkIDs[check.ID] = struct{}{}
			if err := workspace.ValidateID("repo id", check.Repo); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s check %q has invalid repo %q: %v", rel, check.ID, check.Repo, err))
				continue
			}
			if _, ok := repoIDs[check.Repo]; !ok {
				report.Errors = append(report.Errors, fmt.Sprintf("%s references unknown repo in checks: %q", rel, check.Repo))
			}
			if _, ok := pinnedRepos[check.Repo]; !ok {
				report.Errors = append(report.Errors, fmt.Sprintf("%s check %q references repo %q not present in pinned repos", rel, check.ID, check.Repo))
			}
			if check.Run == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s check %q has empty run command", rel, check.ID))
			} else if err := validateCommandRun(check.Run); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s check %q has unsupported run command: %v", rel, check.ID, err))
			}
			if err := validateRelativeCWD(check.CWD); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s check %q has invalid cwd %q: %v", rel, check.ID, check.CWD, err))
			}
			if _, ok := allowedCheckStatuses[check.Status]; !ok {
				report.Errors = append(report.Errors, fmt.Sprintf("%s check %q has unsupported status %q", rel, check.ID, check.Status))
			}
		}
	}
}

func validateRelativeCWD(cwd string) error {
	if strings.TrimSpace(cwd) == "" {
		return nil
	}
	if filepath.IsAbs(cwd) {
		return fmt.Errorf("absolute paths are not allowed")
	}
	clean := filepath.Clean(cwd)
	if escapesBase(clean) {
		return fmt.Errorf("path escapes repo checkout")
	}
	return nil
}

func validateCommandRun(run string) error {
	if strings.TrimSpace(run) == "" {
		return fmt.Errorf("empty command")
	}
	if strings.ContainsAny(run, `"'`) {
		return fmt.Errorf("quoted arguments are not supported; use a repo-local wrapper script")
	}
	return nil
}

func validateReports(root string, report *Report) {
	paths, _ := filepath.Glob(filepath.Join(root, "local", "reports", "*", "*.yaml"))
	for _, path := range paths {
		var doc model.ScenarioReportDocument
		if err := manifest.LoadYAML(path, &doc); err != nil {
			report.Errors = append(report.Errors, err.Error())
			continue
		}
		rel, _ := filepath.Rel(root, path)
		reportDir := filepath.Dir(path)
		scenarioID := filepath.Base(reportDir)
		if doc.Report.Scenario == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: report.scenario is required", rel))
		} else if doc.Report.Scenario != scenarioID {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: report.scenario %q does not match report directory %q", rel, doc.Report.Scenario, scenarioID))
		}
		if doc.Report.ReportKind == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: report.report_kind is required", rel))
		}
		for _, result := range doc.Results {
			if result.Check == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: result.check is required", rel))
			}
			if _, ok := allowedCheckStatuses[result.Status]; !ok {
				report.Errors = append(report.Errors, fmt.Sprintf("%s result %q has unsupported status %q", rel, result.Check, result.Status))
			}
			validateReportPath(root, reportDir, rel, result.Check, "stdout_path", result.StdoutPath, report)
			validateReportPath(root, reportDir, rel, result.Check, "stderr_path", result.StderrPath, report)
			for _, artifact := range result.Artifacts {
				artifactCopy := artifact
				validateReportPath(root, reportDir, rel, result.Check, "artifact", &artifactCopy, report)
			}
		}
	}
}

func validateReportPath(root string, reportDir string, reportRel string, checkID string, field string, value *string, report *Report) {
	if value == nil || *value == "" {
		return
	}
	if filepath.IsAbs(*value) {
		report.Errors = append(report.Errors, fmt.Sprintf("%s result %q has absolute %s %q", reportRel, checkID, field, *value))
		return
	}
	clean := filepath.Clean(*value)
	if unsafeReportPath(clean) {
		report.Errors = append(report.Errors, fmt.Sprintf("%s result %q has unsafe %s %q", reportRel, checkID, field, *value))
		return
	}
	target := filepath.Join(reportDir, clean)
	if relToRoot, err := filepath.Rel(root, target); err == nil && escapesBase(relToRoot) {
		report.Errors = append(report.Errors, fmt.Sprintf("%s result %q has %s outside workspace %q", reportRel, checkID, field, *value))
		return
	}
	if _, err := os.Stat(target); err != nil {
		report.Warnings = append(report.Warnings, fmt.Sprintf("%s result %q references missing %s %q", reportRel, checkID, field, *value))
	}
}

func unsafeReportPath(clean string) bool {
	return clean == "" || clean == "." || escapesBase(clean)
}

func escapesBase(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel)
}
