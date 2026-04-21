package scaffold

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time"

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
	if err := workspace.Init(opts.Root); err != nil {
		return Result{}, err
	}
	root, err := workspace.FindRoot(opts.Root)
	if err != nil {
		return Result{}, err
	}
	result := Result{Root: root}

	seenRepos := map[string]struct{}{}
	repoIDs := make([]string, 0, len(opts.Repos))
	for _, repo := range opts.Repos {
		repo.ID = strings.TrimSpace(repo.ID)
		repo.Kind = strings.TrimSpace(repo.Kind)
		repo.Path = strings.TrimSpace(repo.Path)
		if repo.Kind == "" {
			repo.Kind = "app"
		}
		if _, ok := seenRepos[repo.ID]; ok {
			return Result{}, fmt.Errorf("duplicate repo %q", repo.ID)
		}
		seenRepos[repo.ID] = struct{}{}
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

	if len(opts.Relations) > 0 {
		relations, err := addRelations(root, opts.Relations)
		if err != nil {
			return Result{}, err
		}
		result.Relations = relations
	}

	if len(repoIDs) > 0 || opts.ContextID != "" || opts.ChangeTitle != "" {
		contextID := strings.TrimSpace(opts.ContextID)
		if contextID == "" {
			contextID = defaultContextID
		}
		if len(repoIDs) == 0 {
			return Result{}, fmt.Errorf("context scaffold requires at least one --repo")
		}
		if err := writeContext(root, contextID, repoIDs); err != nil {
			return Result{}, err
		}
		result.ContextID = contextID
	}

	if strings.TrimSpace(opts.ChangeTitle) != "" {
		changeID, err := workspace.CreateChange(root, result.ContextID, opts.ChangeTitle, opts.ChangeKind, opts.Now)
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
