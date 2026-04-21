package model

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

type WorkspaceDocument struct {
	Version   int           `yaml:"version,omitempty"`
	Workspace WorkspaceMeta `yaml:"workspace"`
	Repos     []string      `yaml:"repos"`
	Relations []Relation    `yaml:"relations"`
	Rules     []string      `yaml:"rules"`
}

type WorkspaceMeta struct {
	ID    string `yaml:"id"`
	Model string `yaml:"model"`
}

type Relation struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
	Kind string `yaml:"kind"`
}

const (
	RelationKindRuntime  = "runtime"
	RelationKindBuild    = "build"
	RelationKindContract = "contract"
	RelationKindRelease  = "release"
	RelationKindDocs     = "docs"
)

var relationKinds = []string{
	RelationKindRuntime,
	RelationKindBuild,
	RelationKindContract,
	RelationKindRelease,
	RelationKindDocs,
}

var relationKindSet = map[string]struct{}{
	RelationKindRuntime:  {},
	RelationKindBuild:    {},
	RelationKindContract: {},
	RelationKindRelease:  {},
	RelationKindDocs:     {},
}

func IsRelationKind(value string) bool {
	_, ok := relationKindSet[value]
	return ok
}

func RelationKinds() []string {
	return append([]string(nil), relationKinds...)
}

type RepoDocument struct {
	Version     int                   `yaml:"version,omitempty"`
	Repo        RepoMeta              `yaml:"repo"`
	ReadFirst   []string              `yaml:"read_first"`
	Entrypoints map[string]Entrypoint `yaml:"entrypoints"`
}

type RepoMeta struct {
	ID   string `yaml:"id"`
	Kind string `yaml:"kind"`
}

type Entrypoint struct {
	Run               string   `yaml:"run"`
	CWD               string   `yaml:"cwd,omitempty"`
	TimeoutSeconds    int      `yaml:"timeout_seconds,omitempty"`
	EnvProfile        string   `yaml:"env_profile,omitempty"`
	EnvRequirements   []string `yaml:"env_requirements,omitempty"`
	ExpectedArtifacts []string `yaml:"expected_artifacts,omitempty"`
	RequiresCleanTree bool     `yaml:"requires_clean_worktree,omitempty"`
}

var entrypointFields = map[string]struct{}{
	"run":                     {},
	"cwd":                     {},
	"timeout_seconds":         {},
	"env_profile":             {},
	"env_requirements":        {},
	"expected_artifacts":      {},
	"requires_clean_worktree": {},
}

func (e *Entrypoint) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var run string
		if err := value.Decode(&run); err != nil {
			return err
		}
		*e = Entrypoint{Run: run, CWD: "."}
		return nil
	case yaml.MappingNode:
		if err := rejectUnknownFields(value, "entrypoint", entrypointFields); err != nil {
			return err
		}
		type raw Entrypoint
		var decoded raw
		if err := value.Decode(&decoded); err != nil {
			return err
		}
		*e = Entrypoint(decoded)
		if e.CWD == "" {
			e.CWD = "."
		}
		return nil
	default:
		return fmt.Errorf("entrypoint must be a command string or object")
	}
}

type ContextsDocument struct {
	Version  int                `yaml:"version,omitempty"`
	Contexts map[string]Context `yaml:"contexts"`
}

type Context struct {
	Repos []string `yaml:"repos"`
}

type ChangeDocument struct {
	Version int    `yaml:"version,omitempty"`
	Change  Change `yaml:"change"`
}

type Change struct {
	ID      string   `yaml:"id"`
	Title   string   `yaml:"title"`
	Kind    string   `yaml:"kind"`
	Context string   `yaml:"context"`
	Repos   []string `yaml:"repos"`
}

type BindingsDocument struct {
	Version  int                `yaml:"version,omitempty"`
	Bindings map[string]Binding `yaml:"bindings"`
}

type Binding struct {
	Path string `yaml:"path"`
}

var bindingFields = map[string]struct{}{
	"path": {},
}

func (b *Binding) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var path string
		if err := value.Decode(&path); err != nil {
			return err
		}
		*b = Binding{Path: path}
		return nil
	case yaml.MappingNode:
		if err := rejectUnknownFields(value, "binding", bindingFields); err != nil {
			return err
		}
		type raw Binding
		var decoded raw
		if err := value.Decode(&decoded); err != nil {
			return err
		}
		*b = Binding(decoded)
		return nil
	default:
		return fmt.Errorf("binding must be a path string or object")
	}
}

func rejectUnknownFields(value *yaml.Node, kind string, allowed map[string]struct{}) error {
	for index := 0; index < len(value.Content); index += 2 {
		key := value.Content[index].Value
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("%s has unknown field %q", kind, key)
		}
	}
	return nil
}

type RuleDocument struct {
	Version int  `yaml:"version,omitempty"`
	Rule    Rule `yaml:"rule"`
}

type Rule struct {
	ID        string        `yaml:"id"`
	Kind      string        `yaml:"kind"`
	AppliesTo RuleAppliesTo `yaml:"applies_to,omitempty"`
	Policy    RulePolicy    `yaml:"policy"`
}

type RuleAppliesTo struct {
	RelationKind string `yaml:"relation_kind,omitempty"`
	FromRepo     string `yaml:"from_repo,omitempty"`
	ToRepo       string `yaml:"to_repo,omitempty"`
	Context      string `yaml:"context,omitempty"`
}

type RulePolicy struct {
	Order string `yaml:"order,omitempty"`
}

type ScenarioLockDocument struct {
	Version      int             `yaml:"version,omitempty"`
	Scenario     ScenarioMeta    `yaml:"scenario"`
	ToolVersions ToolVersions    `yaml:"tool_versions"`
	Repos        []ScenarioRepo  `yaml:"repos"`
	Checks       []ScenarioCheck `yaml:"checks"`
}

type ScenarioMeta struct {
	ID          string      `yaml:"id"`
	Change      string      `yaml:"change"`
	Context     string      `yaml:"context"`
	GeneratedAt string      `yaml:"generated_at"`
	GeneratedBy GeneratedBy `yaml:"generated_by"`
	Semantics   string      `yaml:"semantics"`
	Notes       []string    `yaml:"notes,omitempty"`
}

type GeneratedBy struct {
	Tool    string `yaml:"tool"`
	Version string `yaml:"version"`
}

type ToolVersions struct {
	WKit  string            `yaml:"wkit"`
	Git   string            `yaml:"git"`
	Extra map[string]string `yaml:"extra"`
}

type ScenarioRepo struct {
	Repo            string          `yaml:"repo"`
	Revision        Revision        `yaml:"revision"`
	Worktree        Worktree        `yaml:"worktree"`
	DependencyHints DependencyHints `yaml:"dependency_hints"`
}

type Revision struct {
	Commit string `yaml:"commit"`
	Short  string `yaml:"short"`
	Branch string `yaml:"branch"`
}

type Worktree struct {
	Clean          bool     `yaml:"clean"`
	DirtyFiles     int      `yaml:"dirty_files"`
	UntrackedFiles int      `yaml:"untracked_files"`
	DirtyPaths     []string `yaml:"dirty_paths"`
	UntrackedPaths []string `yaml:"untracked_paths"`
}

type DependencyHints struct {
	Lockfiles []LockfileHint `yaml:"lockfiles"`
}

type LockfileHint struct {
	Path   string  `yaml:"path"`
	SHA256 *string `yaml:"sha256"`
}

type ScenarioCheck struct {
	ID                    string   `yaml:"id"`
	Repo                  string   `yaml:"repo"`
	CWD                   string   `yaml:"cwd"`
	Run                   string   `yaml:"run"`
	TimeoutSeconds        int      `yaml:"timeout_seconds"`
	EnvProfile            string   `yaml:"env_profile"`
	EnvRequirements       []string `yaml:"env_requirements"`
	ExpectedArtifacts     []string `yaml:"expected_artifacts"`
	RequiresCleanWorktree bool     `yaml:"requires_clean_worktree"`
	Status                string   `yaml:"status"`
}

type ScenarioReportDocument struct {
	Version int                  `yaml:"version,omitempty"`
	Report  ScenarioReportMeta   `yaml:"report"`
	Results []ScenarioRunOutcome `yaml:"results"`
}

type ScenarioReportMeta struct {
	Scenario    string `yaml:"scenario"`
	GeneratedAt string `yaml:"generated_at"`
	ReportKind  string `yaml:"report_kind"`
}

type ScenarioRunOutcome struct {
	Check           string   `yaml:"check"`
	Status          string   `yaml:"status"`
	DurationSeconds float64  `yaml:"duration_seconds"`
	Reason          string   `yaml:"reason,omitempty"`
	EnvProfile      string   `yaml:"env_profile,omitempty"`
	StdoutPath      *string  `yaml:"stdout_path"`
	StderrPath      *string  `yaml:"stderr_path"`
	Artifacts       []string `yaml:"artifacts"`
}
