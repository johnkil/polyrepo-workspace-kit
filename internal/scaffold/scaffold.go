package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/fsutil"
	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

const defaultContextID = "default"

type Options struct {
	Root        string
	Repos       []RepoSpec
	Relations   []RelationSpec
	ContextID   string
	ChangeTitle string
	ChangeKind  string
	Now         time.Time
}

type RepoSpec struct {
	ID   string
	Path string
	Kind string
}

type RelationSpec struct {
	From string
	To   string
	Kind string
}

type Result struct {
	Root       string
	Repos      []RepoResult
	Relations  []RelationSpec
	ContextID  string
	ChangeID   string
	ChangePath string
}

type RepoResult struct {
	ID           string
	ManifestPath string
	BindingPath  string
}

func ParseRepoSpec(value string) (RepoSpec, error) {
	id, path, err := splitPair(value, "=")
	if err != nil {
		return RepoSpec{}, fmt.Errorf("repo must use id=path: %w", err)
	}
	if path == "" {
		return RepoSpec{}, fmt.Errorf("repo %q has empty path", id)
	}
	return RepoSpec{ID: id, Path: path, Kind: "app"}, nil
}

func ParseRepoKindSpec(value string) (string, string, error) {
	id, kind, err := splitPair(value, "=")
	if err != nil {
		return "", "", fmt.Errorf("repo kind must use id=kind: %w", err)
	}
	if kind == "" {
		return "", "", fmt.Errorf("repo %q has empty kind", id)
	}
	return id, kind, nil
}

func ParseRelationSpec(value string) (RelationSpec, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return RelationSpec{}, fmt.Errorf("relation must use from:to:kind")
	}
	relation := RelationSpec{
		From: strings.TrimSpace(parts[0]),
		To:   strings.TrimSpace(parts[1]),
		Kind: strings.TrimSpace(parts[2]),
	}
	if relation.Kind == "" {
		relation.Kind = "contract"
	}
	if !model.IsRelationKind(relation.Kind) {
		return RelationSpec{}, fmt.Errorf("relation %s -> %s uses unsupported kind %q", relation.From, relation.To, relation.Kind)
	}
	return relation, nil
}

func Apply(opts Options) (Result, error) {
	if opts.Root == "" {
		return Result{}, fmt.Errorf("workspace path is required")
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	plan, err := planScaffold(opts)
	if err != nil {
		return Result{}, err
	}
	if err := workspace.Init(plan.Root); err != nil {
		return Result{}, err
	}
	root, err := workspace.FindRoot(plan.Root)
	if err != nil {
		return Result{}, err
	}
	result := Result{Root: root}

	repoIDs := make([]string, 0, len(plan.Repos))
	for _, repo := range plan.Repos {
		manifestPath, err := workspace.RegisterRepo(root, repo.ID, repo.Kind)
		if err != nil {
			return Result{}, err
		}
		bindingPath, err := workspace.SetBinding(root, repo.ID, repo.Path)
		if err != nil {
			return Result{}, err
		}
		repoIDs = append(repoIDs, repo.ID)
		result.Repos = append(result.Repos, RepoResult{
			ID:           repo.ID,
			ManifestPath: manifestPath,
			BindingPath:  bindingPath,
		})
	}

	if len(plan.Relations) > 0 {
		relations, err := addRelations(root, plan.Relations)
		if err != nil {
			return Result{}, err
		}
		result.Relations = relations
	}

	if plan.ContextID != "" {
		if err := writeContext(root, plan.ContextID, repoIDs); err != nil {
			return Result{}, err
		}
		result.ContextID = plan.ContextID
	}

	if strings.TrimSpace(plan.ChangeTitle) != "" {
		changeID, err := workspace.CreateChange(root, result.ContextID, plan.ChangeTitle, plan.ChangeKind, opts.Now)
		if err != nil {
			return Result{}, err
		}
		changePath, err := workspace.ChangePath(root, changeID)
		if err != nil {
			return Result{}, err
		}
		result.ChangeID = changeID
		result.ChangePath = changePath
	}

	return result, nil
}

type scaffoldPlan struct {
	Root        string
	Repos       []RepoSpec
	Relations   []RelationSpec
	ContextID   string
	ChangeTitle string
	ChangeKind  string
}

func planScaffold(opts Options) (scaffoldPlan, error) {
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return scaffoldPlan{}, err
	}
	plan := scaffoldPlan{
		Root:        root,
		Repos:       make([]RepoSpec, 0, len(opts.Repos)),
		Relations:   make([]RelationSpec, 0, len(opts.Relations)),
		ContextID:   strings.TrimSpace(opts.ContextID),
		ChangeTitle: opts.ChangeTitle,
		ChangeKind:  opts.ChangeKind,
	}

	existingRepoIDs, err := existingRepoIDs(root)
	if err != nil {
		return scaffoldPlan{}, err
	}
	knownRepoIDs := map[string]struct{}{}
	for repoID := range existingRepoIDs {
		knownRepoIDs[repoID] = struct{}{}
	}
	seenRepos := map[string]struct{}{}
	for _, repo := range opts.Repos {
		repo.ID = strings.TrimSpace(repo.ID)
		repo.Kind = strings.TrimSpace(repo.Kind)
		repo.Path = strings.TrimSpace(repo.Path)
		if repo.Kind == "" {
			repo.Kind = "app"
		}
		if _, ok := seenRepos[repo.ID]; ok {
			return scaffoldPlan{}, fmt.Errorf("duplicate repo %q", repo.ID)
		}
		if err := workspace.ValidateID("repo id", repo.ID); err != nil {
			return scaffoldPlan{}, err
		}
		if err := validateBindingPath(repo.ID, repo.Path); err != nil {
			return scaffoldPlan{}, err
		}
		seenRepos[repo.ID] = struct{}{}
		knownRepoIDs[repo.ID] = struct{}{}
		plan.Repos = append(plan.Repos, repo)
	}

	for _, relation := range opts.Relations {
		relation.From = strings.TrimSpace(relation.From)
		relation.To = strings.TrimSpace(relation.To)
		relation.Kind = strings.TrimSpace(relation.Kind)
		if relation.Kind == "" {
			relation.Kind = "contract"
		}
		if !model.IsRelationKind(relation.Kind) {
			return scaffoldPlan{}, fmt.Errorf("relation %s -> %s uses unsupported kind %q", relation.From, relation.To, relation.Kind)
		}
		if err := workspace.ValidateID("relation from repo id", relation.From); err != nil {
			return scaffoldPlan{}, err
		}
		if err := workspace.ValidateID("relation to repo id", relation.To); err != nil {
			return scaffoldPlan{}, err
		}
		if _, ok := knownRepoIDs[relation.From]; !ok {
			return scaffoldPlan{}, fmt.Errorf("unknown relation repo %q; register it first", relation.From)
		}
		if _, ok := knownRepoIDs[relation.To]; !ok {
			return scaffoldPlan{}, fmt.Errorf("unknown relation repo %q; register it first", relation.To)
		}
		plan.Relations = append(plan.Relations, relation)
	}

	if len(plan.Repos) > 0 || plan.ContextID != "" || strings.TrimSpace(plan.ChangeTitle) != "" {
		if plan.ContextID == "" {
			plan.ContextID = defaultContextID
		}
		if len(plan.Repos) == 0 {
			return scaffoldPlan{}, fmt.Errorf("context scaffold requires at least one --repo")
		}
		if err := workspace.ValidateID("context id", plan.ContextID); err != nil {
			return scaffoldPlan{}, err
		}
		if err := validateContextPreflight(root, plan.ContextID, repoIDs(plan.Repos)); err != nil {
			return scaffoldPlan{}, err
		}
	}

	return plan, nil
}

func existingRepoIDs(root string) (map[string]struct{}, error) {
	path := filepath.Join(root, workspace.WorkspaceFile)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return nil, err
	}
	return workspace.RepoIDs(doc), nil
}

func validateBindingPath(repoID string, path string) error {
	if path == "" {
		return fmt.Errorf("repo %q has empty path", repoID)
	}
	absPath, err := fsutil.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("binding path for repo %q is not accessible: %w", repoID, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("binding path for repo %q is not a directory: %s", repoID, absPath)
	}
	return nil
}

func validateContextPreflight(root string, contextID string, ids []string) error {
	path := filepath.Join(root, workspace.ContextsFile)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	contexts, err := workspace.LoadContexts(root)
	if err != nil {
		return err
	}
	next := model.Context{Repos: append([]string(nil), ids...)}
	if existing, ok := contexts.Contexts[contextID]; ok && !reflect.DeepEqual(existing.Repos, next.Repos) {
		return fmt.Errorf("context %q already exists with different repos", contextID)
	}
	return nil
}

func repoIDs(repos []RepoSpec) []string {
	out := make([]string, 0, len(repos))
	for _, repo := range repos {
		out = append(out, repo.ID)
	}
	return out
}

func addRelations(root string, specs []RelationSpec) ([]RelationSpec, error) {
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return nil, err
	}
	repoIDs := workspace.RepoIDs(doc)
	added := make([]RelationSpec, 0, len(specs))
	for _, spec := range specs {
		spec.From = strings.TrimSpace(spec.From)
		spec.To = strings.TrimSpace(spec.To)
		spec.Kind = strings.TrimSpace(spec.Kind)
		if spec.Kind == "" {
			spec.Kind = "contract"
		}
		if !model.IsRelationKind(spec.Kind) {
			return nil, fmt.Errorf("relation %s -> %s uses unsupported kind %q", spec.From, spec.To, spec.Kind)
		}
		if err := workspace.ValidateID("relation from repo id", spec.From); err != nil {
			return nil, err
		}
		if err := workspace.ValidateID("relation to repo id", spec.To); err != nil {
			return nil, err
		}
		if _, ok := repoIDs[spec.From]; !ok {
			return nil, fmt.Errorf("unknown relation repo %q; register it first", spec.From)
		}
		if _, ok := repoIDs[spec.To]; !ok {
			return nil, fmt.Errorf("unknown relation repo %q; register it first", spec.To)
		}
		relation := model.Relation{From: spec.From, To: spec.To, Kind: spec.Kind}
		if containsRelation(doc.Relations, relation) {
			continue
		}
		doc.Relations = append(doc.Relations, relation)
		added = append(added, spec)
	}
	if err := workspace.SaveWorkspace(root, doc); err != nil {
		return nil, err
	}
	return added, nil
}

func writeContext(root string, contextID string, repoIDs []string) error {
	if err := workspace.ValidateID("context id", contextID); err != nil {
		return err
	}
	contexts, err := workspace.LoadContexts(root)
	if err != nil {
		return err
	}
	next := model.Context{Repos: append([]string(nil), repoIDs...)}
	if existing, ok := contexts.Contexts[contextID]; ok && !reflect.DeepEqual(existing.Repos, next.Repos) {
		return fmt.Errorf("context %q already exists with different repos", contextID)
	}
	contexts.Contexts[contextID] = next
	return manifest.WriteYAML(filepath.Join(root, workspace.ContextsFile), contexts)
}

func containsRelation(relations []model.Relation, relation model.Relation) bool {
	for _, existing := range relations {
		if existing == relation {
			return true
		}
	}
	return false
}

func splitPair(value string, separator string) (string, string, error) {
	left, right, ok := strings.Cut(value, separator)
	if !ok {
		return "", "", fmt.Errorf("%q is missing %q", value, separator)
	}
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" {
		return "", "", fmt.Errorf("%q has empty id", value)
	}
	return left, right, nil
}
