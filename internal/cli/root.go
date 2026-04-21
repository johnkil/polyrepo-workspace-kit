package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/buildinfo"
	"github.com/johnkil/polyrepo-workspace-kit/internal/demo"
	"github.com/johnkil/polyrepo-workspace-kit/internal/handoff"
	"github.com/johnkil/polyrepo-workspace-kit/internal/install"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/orient"
	"github.com/johnkil/polyrepo-workspace-kit/internal/relations"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scaffold"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
	"github.com/johnkil/polyrepo-workspace-kit/internal/telemetry"
	"github.com/johnkil/polyrepo-workspace-kit/internal/validate"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func Execute() int {
	record := &cliRunRecord{}
	root := newRootCommandWithRecorder(record)
	return executeRoot(root, record, time.Now())
}

type cliRunRecord struct {
	Command       string
	Args          []string
	WorkspaceFlag string
}

func executeRoot(root *cobra.Command, record *cliRunRecord, started time.Time) int {
	executed, err := root.ExecuteC()
	code := exitCode(err)
	completeRunRecord(record, executed)
	recordTelemetryEvent(record, code, time.Since(started), time.Now())
	if err == nil {
		return 0
	}
	if printErr := printableError(err); printErr != nil {
		_, _ = fmt.Fprintln(root.ErrOrStderr(), printErr)
	}
	return code
}

func completeRunRecord(record *cliRunRecord, cmd *cobra.Command) {
	if record == nil || record.Command != "" || cmd == nil {
		return
	}
	record.Command = cmd.CommandPath()
	record.Args = captureArgs(cmd, cmd.Flags().Args())
	record.WorkspaceFlag = captureWorkspaceFlag(cmd)
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		if exitErr.Code == 0 {
			return 1
		}
		return exitErr.Code
	}
	return 1
}

func printableError(err error) error {
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Err
	}
	return err
}

func newRootCommand() *cobra.Command {
	return newRootCommandWithRecorder(nil)
}

func newRootCommandWithRecorder(record *cliRunRecord) *cobra.Command {
	var workspaceFlag string

	root := &cobra.Command{
		Use:           "wkit",
		Short:         "Coordinate local polyrepo workspaces",
		Version:       buildinfo.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("wkit {{ .Version }}\n")
	root.PersistentFlags().StringVar(&workspaceFlag, "workspace", "", "workspace root")
	if record != nil {
		root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			record.Command = cmd.CommandPath()
			record.Args = captureArgs(cmd, args)
			record.WorkspaceFlag = workspaceFlag
		}
	}

	root.AddCommand(newInitCommand())
	root.AddCommand(newDemoCommand())
	root.AddCommand(newRepoCommand(&workspaceFlag))
	root.AddCommand(newBindCommand(&workspaceFlag))
	root.AddCommand(newContextCommand(&workspaceFlag))
	root.AddCommand(newRelationsCommand(&workspaceFlag))
	root.AddCommand(newChangeCommand(&workspaceFlag))
	root.AddCommand(newScenarioCommand(&workspaceFlag))
	root.AddCommand(newInfoCommand(&workspaceFlag))
	root.AddCommand(newStatusCommand(&workspaceFlag))
	root.AddCommand(newDoctorCommand(&workspaceFlag))
	root.AddCommand(newHandoffCommand(&workspaceFlag))
	root.AddCommand(newInstallCommand(&workspaceFlag))
	root.AddCommand(newTelemetryCommand(&workspaceFlag))
	root.AddCommand(newVSCodeCommand(&workspaceFlag))
	root.AddCommand(newValidateCommand(&workspaceFlag))
	root.AddCommand(newVersionCommand())

	return root
}

func captureArgs(cmd *cobra.Command, args []string) []string {
	out := append([]string(nil), args...)
	visit := func(flag *pflag.Flag) {
		out = append(out, "--"+flag.Name, flag.Value.String())
	}
	cmd.Flags().Visit(visit)
	cmd.InheritedFlags().Visit(visit)
	return out
}

func captureWorkspaceFlag(cmd *cobra.Command) string {
	flag := cmd.Flag("workspace")
	if flag == nil {
		return ""
	}
	return flag.Value.String()
}

func recordTelemetryEvent(record *cliRunRecord, code int, duration time.Duration, now time.Time) {
	if record == nil || record.Command == "" {
		return
	}
	root, err := telemetryRoot(record.WorkspaceFlag)
	if err != nil {
		return
	}
	_ = telemetry.RecordIfEnabled(root, telemetry.Event{
		Timestamp:  now.UTC().Format(time.RFC3339),
		Workspace:  root,
		Command:    record.Command,
		Args:       record.Args,
		ExitCode:   code,
		DurationMS: duration.Milliseconds(),
	})
}

func telemetryRoot(workspaceFlag string) (string, error) {
	if workspaceFlag != "" {
		return workspace.FindRoot(workspaceFlag)
	}
	return workspace.FindRoot("")
}

func resolveWorkspaceRoot(flag *string) (string, error) {
	if flag != nil && *flag != "" {
		return workspace.FindRoot(*flag)
	}
	return workspace.FindRoot("")
}

func writef(cmd *cobra.Command, format string, args ...any) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), format, args...)
	return err
}

func writeln(cmd *cobra.Command, args ...any) error {
	_, err := fmt.Fprintln(cmd.OutOrStdout(), args...)
	return err
}

func write(cmd *cobra.Command, value string) error {
	_, err := fmt.Fprint(cmd.OutOrStdout(), value)
	return err
}

func newDemoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "demo [minimal|failure]",
		Short: "Run a self-contained first-run demo",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind := demo.KindMinimal
			if len(args) == 1 {
				kind = args[0]
			}
			result, err := demo.Run(kind, time.Now())
			if err != nil {
				return err
			}
			if err := printDemoResult(cmd, result); err != nil {
				return err
			}
			return nil
		},
	}
}

func printDemoResult(cmd *cobra.Command, result demo.Result) error {
	if err := writef(cmd, "demo: %s\n", result.Kind); err != nil {
		return err
	}
	if err := writef(cmd, "workspace: %s\n", result.WorkspaceRoot); err != nil {
		return err
	}
	if err := writef(cmd, "repos: %s\n", result.ReposRoot); err != nil {
		return err
	}
	if err := writef(cmd, "change: %s\n", result.ChangeID); err != nil {
		return err
	}
	if err := writef(cmd, "scenario: %s\n", result.ScenarioID); err != nil {
		return err
	}
	if err := writef(cmd, "report: %s\n", result.ReportPath); err != nil {
		return err
	}
	if err := writef(cmd, "text-report: %s\n", result.TextReportPath); err != nil {
		return err
	}
	if err := writef(cmd, "markdown-report: %s\n", result.MarkdownReportPath); err != nil {
		return err
	}
	if err := writef(cmd, "handoff-command: wkit --workspace %s handoff %s\n", result.WorkspaceRoot, result.ChangeID); err != nil {
		return err
	}
	if result.Kind == demo.KindFailure {
		if err := writef(cmd, "expected: drift=%t blocked=%t failed=%t\n", result.Drift, result.Blocked, result.Failed); err != nil {
			return err
		}
	}
	if err := writeln(cmd, "\n--- markdown report ---"); err != nil {
		return err
	}
	return write(cmd, result.MarkdownReport)
}

func newInitCommand() *cobra.Command {
	var repoValues []string
	var repoKindValues []string
	var relationValues []string
	var contextID string
	var changeTitle string
	var changeKind string

	cmd := &cobra.Command{
		Use:   "init <path>",
		Short: "Initialize a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := initScaffoldOptions(args[0], repoValues, repoKindValues, relationValues, contextID, changeTitle, changeKind)
			if err != nil {
				return err
			}
			result, err := scaffold.Apply(opts)
			if err != nil {
				return err
			}
			return printInitResult(cmd, result)
		},
	}
	cmd.Flags().StringArrayVar(&repoValues, "repo", nil, "register and bind a repo as id=path")
	cmd.Flags().StringArrayVar(&repoKindValues, "repo-kind", nil, "set kind for a scaffold repo as id=kind")
	cmd.Flags().StringArrayVar(&relationValues, "relation", nil, "add a relation as from:to:kind")
	cmd.Flags().StringVar(&contextID, "context", "", "create a context for scaffold repos")
	cmd.Flags().StringVar(&changeTitle, "change-title", "", "create an initial change from the scaffold context")
	cmd.Flags().StringVar(&changeKind, "change-kind", "contract", "kind for --change-title")
	return cmd
}

func initScaffoldOptions(root string, repoValues []string, repoKindValues []string, relationValues []string, contextID string, changeTitle string, changeKind string) (scaffold.Options, error) {
	repoKinds := map[string]string{}
	for _, value := range repoKindValues {
		id, kind, err := scaffold.ParseRepoKindSpec(value)
		if err != nil {
			return scaffold.Options{}, err
		}
		if _, exists := repoKinds[id]; exists {
			return scaffold.Options{}, fmt.Errorf("duplicate repo kind for %q", id)
		}
		repoKinds[id] = kind
	}

	repos := make([]scaffold.RepoSpec, 0, len(repoValues))
	seenRepos := map[string]struct{}{}
	for _, value := range repoValues {
		repo, err := scaffold.ParseRepoSpec(value)
		if err != nil {
			return scaffold.Options{}, err
		}
		if _, exists := seenRepos[repo.ID]; exists {
			return scaffold.Options{}, fmt.Errorf("duplicate repo %q", repo.ID)
		}
		seenRepos[repo.ID] = struct{}{}
		if kind, ok := repoKinds[repo.ID]; ok {
			repo.Kind = kind
			delete(repoKinds, repo.ID)
		}
		repos = append(repos, repo)
	}
	for repoID := range repoKinds {
		return scaffold.Options{}, fmt.Errorf("--repo-kind references unknown --repo %q", repoID)
	}

	relations := make([]scaffold.RelationSpec, 0, len(relationValues))
	for _, value := range relationValues {
		relation, err := scaffold.ParseRelationSpec(value)
		if err != nil {
			return scaffold.Options{}, err
		}
		relations = append(relations, relation)
	}

	return scaffold.Options{
		Root:        root,
		Repos:       repos,
		Relations:   relations,
		ContextID:   contextID,
		ChangeTitle: changeTitle,
		ChangeKind:  changeKind,
		Now:         time.Now(),
	}, nil
}

func printInitResult(cmd *cobra.Command, result scaffold.Result) error {
	if err := writef(cmd, "initialized workspace at %s\n", result.Root); err != nil {
		return err
	}
	for _, repo := range result.Repos {
		if err := writef(cmd, "registered %s at %s\n", repo.ID, repo.ManifestPath); err != nil {
			return err
		}
		if err := writef(cmd, "bound %s to %s\n", repo.ID, repo.BindingPath); err != nil {
			return err
		}
	}
	for _, relation := range result.Relations {
		if err := writef(cmd, "relation: %s -> %s kind=%s\n", relation.From, relation.To, relation.Kind); err != nil {
			return err
		}
	}
	if result.ContextID != "" {
		if err := writef(cmd, "context: %s\n", result.ContextID); err != nil {
			return err
		}
	}
	if result.ChangeID != "" {
		if err := writef(cmd, "change: %s at %s\n", result.ChangeID, result.ChangePath); err != nil {
			return err
		}
	}
	return nil
}

func newRepoCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage workspace repo manifests",
	}
	cmd.AddCommand(newRepoRegisterCommand(workspaceFlag))
	return cmd
}

func newRepoRegisterCommand(workspaceFlag *string) *cobra.Command {
	var kind string
	cmd := &cobra.Command{
		Use:   "register <repo-id>",
		Short: "Register a repo in the workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			path, err := workspace.RegisterRepo(root, args[0], kind)
			if err != nil {
				return err
			}
			return writef(cmd, "registered %s at %s\n", args[0], path)
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "repo kind")
	_ = cmd.MarkFlagRequired("kind")
	return cmd
}

func newBindCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Manage local repo bindings",
	}
	cmd.AddCommand(newBindSetCommand(workspaceFlag))
	return cmd
}

func newBindSetCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set <repo-id> <path>",
		Short: "Bind a repo id to a local checkout path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			path, err := workspace.SetBinding(root, args[0], args[1])
			if err != nil {
				return err
			}
			return writef(cmd, "bound %s to %s\n", args[0], path)
		},
	}
}

func newContextCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Inspect named workspace contexts",
	}
	cmd.AddCommand(newContextListCommand(workspaceFlag))
	cmd.AddCommand(newContextShowCommand(workspaceFlag))
	return cmd
}

func newContextListCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List named contexts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			contexts, err := orient.ListContexts(root)
			if err != nil {
				return err
			}
			if err := writeln(cmd, "contexts:"); err != nil {
				return err
			}
			for _, context := range contexts {
				if err := writef(cmd, "- %s repos=%d\n", context.ID, context.RepoCount); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newContextShowCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show <context-id>",
		Short: "Show a named context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			context, err := orient.GetContext(root, args[0])
			if err != nil {
				return err
			}
			if err := writef(cmd, "context: %s\n", args[0]); err != nil {
				return err
			}
			if err := writeln(cmd, "repos:"); err != nil {
				return err
			}
			for _, repoID := range context.Repos {
				if err := writef(cmd, "- %s\n", repoID); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newRelationsCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relations",
		Short: "Inspect workspace relation candidates",
	}
	cmd.AddCommand(newRelationsSuggestCommand(workspaceFlag))
	return cmd
}

func newRelationsSuggestCommand(workspaceFlag *string) *cobra.Command {
	var contextID string
	cmd := &cobra.Command{
		Use:   "suggest",
		Short: "Suggest missing relations from local dependency manifests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			report, err := relations.Suggest(root, relations.Options{ContextID: contextID})
			if err != nil {
				return err
			}
			return printRelationSuggestions(cmd, report)
		},
	}
	cmd.Flags().StringVar(&contextID, "context", "", "limit suggestions to repos in a named context")
	return cmd
}

func printRelationSuggestions(cmd *cobra.Command, report relations.Report) error {
	if err := writeln(cmd, "suggestions:"); err != nil {
		return err
	}
	if len(report.Suggestions) == 0 {
		if err := writeln(cmd, "- none"); err != nil {
			return err
		}
	} else {
		for _, suggestion := range report.Suggestions {
			if err := writef(
				cmd,
				"- %s -> %s kind=%s source=%q evidence=%q matched=%q\n",
				suggestion.From,
				suggestion.To,
				suggestion.Kind,
				suggestion.Source,
				suggestion.Evidence,
				suggestion.Matched,
			); err != nil {
				return err
			}
		}
		if err := writeln(cmd, "candidate-flags:"); err != nil {
			return err
		}
		for _, suggestion := range report.Suggestions {
			if err := writef(cmd, "- --relation %s:%s:%s\n", suggestion.From, suggestion.To, suggestion.Kind); err != nil {
				return err
			}
		}
	}
	if len(report.Skipped) > 0 {
		if err := writeln(cmd, "skipped:"); err != nil {
			return err
		}
		for _, skipped := range report.Skipped {
			if err := writef(cmd, "- %s: %s\n", skipped.Repo, skipped.Reason); err != nil {
				return err
			}
		}
	}
	return writeln(cmd, "note: suggestions are read-only; accept candidates by editing coordination/workspace.yaml or using explicit init --relation flags.")
}

func newChangeCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change",
		Short: "Manage cross-repo changes",
	}
	cmd.AddCommand(newChangeNewCommand(workspaceFlag))
	cmd.AddCommand(newChangeShowCommand(workspaceFlag))
	return cmd
}

func newChangeNewCommand(workspaceFlag *string) *cobra.Command {
	var title string
	var kind string
	cmd := &cobra.Command{
		Use:   "new <context>",
		Short: "Create a change from a named context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			changeID, err := workspace.CreateChange(root, args[0], title, kind, time.Now())
			if err != nil {
				return err
			}
			path, err := workspace.ChangePath(root, changeID)
			if err != nil {
				return err
			}
			return writef(cmd, "created %s at %s\n", changeID, path)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "change title")
	cmd.Flags().StringVar(&kind, "kind", "contract", "change kind")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func newChangeShowCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show <change-id>",
		Short: "Show a change manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			path, err := workspace.ChangePath(root, args[0])
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}

func newTelemetryCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Manage local opt-in pilot telemetry",
	}
	cmd.AddCommand(newTelemetryEnableCommand(workspaceFlag))
	cmd.AddCommand(newTelemetryDisableCommand(workspaceFlag))
	cmd.AddCommand(newTelemetryStatusCommand(workspaceFlag))
	cmd.AddCommand(newTelemetryExportCommand(workspaceFlag))
	return cmd
}

func newTelemetryEnableCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable local command event logging for this workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			status, err := telemetry.Enable(root, time.Now())
			if err != nil {
				return err
			}
			if err := writeln(cmd, "telemetry: enabled"); err != nil {
				return err
			}
			return printTelemetryStatus(cmd, status)
		},
	}
}

func newTelemetryDisableCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable local command event logging for this workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			status, err := telemetry.Disable(root, time.Now())
			if err != nil {
				return err
			}
			if err := writeln(cmd, "telemetry: disabled"); err != nil {
				return err
			}
			return printTelemetryStatus(cmd, status)
		},
	}
}

func newTelemetryStatusCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show local telemetry status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			status, err := telemetry.ReadStatus(root)
			if err != nil {
				return err
			}
			return printTelemetryStatus(cmd, status)
		},
	}
}

func newTelemetryExportCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Print local telemetry events as JSONL",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			data, err := telemetry.Export(root)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				return nil
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}

func printTelemetryStatus(cmd *cobra.Command, status telemetry.Status) error {
	if err := writef(cmd, "enabled: %t\n", status.Enabled); err != nil {
		return err
	}
	if err := writef(cmd, "config: %s\n", status.ConfigPath); err != nil {
		return err
	}
	if err := writef(cmd, "events: %s\n", status.EventsPath); err != nil {
		return err
	}
	return writef(cmd, "event_count: %d\n", status.EventCount)
}

func newValidateCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate workspace manifests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			report := validate.Workspace(root)
			for _, item := range report.Errors {
				if err := writef(cmd, "error: %s\n", item); err != nil {
					return err
				}
			}
			for _, warning := range report.Warnings {
				if err := writef(cmd, "warning: %s\n", warning); err != nil {
					return err
				}
			}
			if !report.OK() {
				return &ExitError{Code: 2}
			}
			if len(report.Warnings) == 0 {
				return writeln(cmd, "ok: workspace is valid")
			}
			return writef(cmd, "ok: workspace is valid with %d warning(s)\n", len(report.Warnings))
		},
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show wkit build version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := buildinfo.Current()
			if err := writef(cmd, "wkit %s\n", info.Version); err != nil {
				return err
			}
			if err := writef(cmd, "commit: %s\n", info.Commit); err != nil {
				return err
			}
			if err := writef(cmd, "date: %s\n", info.Date); err != nil {
				return err
			}
			if err := writef(cmd, "dirty: %s\n", info.Dirty); err != nil {
				return err
			}
			return writef(cmd, "builtBy: %s\n", info.BuiltBy)
		},
	}
}

func newScenarioCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scenario",
		Short: "Manage reviewable validation snapshots",
	}
	cmd.AddCommand(newScenarioPinCommand(workspaceFlag))
	cmd.AddCommand(newScenarioShowCommand(workspaceFlag))
	cmd.AddCommand(newScenarioStatusCommand(workspaceFlag))
	cmd.AddCommand(newScenarioRunCommand(workspaceFlag))
	return cmd
}

func newScenarioPinCommand(workspaceFlag *string) *cobra.Command {
	var changeID string
	cmd := &cobra.Command{
		Use:   "pin <scenario-id>",
		Short: "Pin a scenario lock from a change",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			path, err := scenario.Pin(root, args[0], changeID, time.Now())
			if err != nil {
				return err
			}
			return writef(cmd, "pinned %s at %s\n", args[0], path)
		},
	}
	cmd.Flags().StringVar(&changeID, "change", "", "change id")
	_ = cmd.MarkFlagRequired("change")
	return cmd
}

func newScenarioShowCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "show <scenario-id>",
		Short: "Show a scenario lock manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			path, err := workspace.ScenarioPath(root, args[0])
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}

func newScenarioStatusCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status <scenario-id>",
		Short: "Compare current checkouts with a scenario lock",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			report, err := orient.ScenarioStatus(root, args[0])
			if err != nil {
				return err
			}
			if err := printScenarioStatus(cmd, report); err != nil {
				return err
			}
			if report.Drift || report.Blocked || report.Missing {
				return &ExitError{Code: 4}
			}
			return nil
		},
	}
}

func newScenarioRunCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "run <scenario-id>",
		Short: "Run a pinned scenario and write a local report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			result, err := scenario.Run(root, args[0], time.Now())
			if err != nil {
				return err
			}
			for _, outcome := range result.Outcomes {
				if err := printScenarioOutcome(cmd, outcome); err != nil {
					return err
				}
			}
			if err := writef(cmd, "report: %s\n", result.ReportPath); err != nil {
				return err
			}
			if err := writef(cmd, "text-report: %s\n", result.TextReportPath); err != nil {
				return err
			}
			if err := writef(cmd, "markdown-report: %s\n", result.MarkdownReportPath); err != nil {
				return err
			}
			return scenarioExitError(result)
		},
	}
}

func newInfoCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:     "info",
		Aliases: []string{"overview"},
		Short:   "Show a workspace overview",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			info, err := orient.WorkspaceInfo(root)
			if err != nil {
				return err
			}
			return printInfo(cmd, info)
		},
	}
}

func newStatusCommand(workspaceFlag *string) *cobra.Command {
	var contextID string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show local checkout status without fetching remotes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			report, err := orient.WorkspaceStatus(root, orient.StatusOptions{ContextID: contextID})
			if err != nil {
				return err
			}
			return printWorkspaceStatus(cmd, report)
		},
	}
	cmd.Flags().StringVar(&contextID, "context", "", "limit status to a named context")
	return cmd
}

func newDoctorCommand(workspaceFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose workspace and local checkout readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			report := orient.Doctor(root)
			for _, warning := range report.Warnings {
				if err := writef(cmd, "warning: %s\n", warning); err != nil {
					return err
				}
			}
			for _, item := range report.Errors {
				if err := writef(cmd, "error: %s\n", item); err != nil {
					return err
				}
			}
			if err := writef(cmd, "summary: errors=%d warnings=%d\n", len(report.Errors), len(report.Warnings)); err != nil {
				return err
			}
			if len(report.Errors) > 0 {
				return &ExitError{Code: 2}
			}
			return nil
		},
	}
}

func printScenarioOutcome(cmd *cobra.Command, outcome model.ScenarioRunOutcome) error {
	reason := ""
	if outcome.Reason != "" {
		reason = fmt.Sprintf(" (%s)", outcome.Reason)
	}
	if err := writef(cmd, "%s: %s%s\n", outcome.Status, outcome.Check, reason); err != nil {
		return err
	}
	if outcome.StdoutPath != nil {
		if err := writef(cmd, "  stdout: %s\n", *outcome.StdoutPath); err != nil {
			return err
		}
	}
	if outcome.StderrPath != nil {
		if err := writef(cmd, "  stderr: %s\n", *outcome.StderrPath); err != nil {
			return err
		}
	}
	return nil
}

func printInfo(cmd *cobra.Command, info orient.Info) error {
	if err := writef(cmd, "workspace: %s\n", info.WorkspaceID); err != nil {
		return err
	}
	if err := writef(cmd, "root: %s\n", info.Root); err != nil {
		return err
	}
	if err := writef(cmd, "repos: %d\n", info.RepoCount); err != nil {
		return err
	}
	if err := printCounts(cmd, "repo_kinds", info.RepoKinds); err != nil {
		return err
	}
	if err := printCounts(cmd, "relation_kinds", info.RelationKinds); err != nil {
		return err
	}
	if err := writeln(cmd, "contexts:"); err != nil {
		return err
	}
	for _, context := range info.Contexts {
		if err := writef(cmd, "- %s repos=%d\n", context.ID, context.RepoCount); err != nil {
			return err
		}
	}
	if err := writef(cmd, "changes: %d latest=%s\n", info.ChangeCount, info.LatestChange); err != nil {
		return err
	}
	if err := writef(cmd, "scenarios: %d latest=%s\n", info.ScenarioCount, info.LatestScenario); err != nil {
		return err
	}
	if err := writef(cmd, "bindings: %d/%d\n", info.BoundRepos, info.TotalRepos); err != nil {
		return err
	}
	if err := writef(cmd, "guidance: rules=%d skills=%d\n", info.GuidanceRules, info.GuidanceSkills); err != nil {
		return err
	}
	if err := writeln(cmd, "next:"); err != nil {
		return err
	}
	for _, next := range []string{"wkit validate", "wkit status", "wkit scenario pin <scenario-id> --change <change-id>", "wkit scenario run <scenario-id>"} {
		if err := writef(cmd, "- %s\n", next); err != nil {
			return err
		}
	}
	return nil
}

func printCounts(cmd *cobra.Command, label string, counts []orient.Count) error {
	if err := writef(cmd, "%s:\n", label); err != nil {
		return err
	}
	if len(counts) == 0 {
		return writeln(cmd, "- none: 0")
	}
	for _, count := range counts {
		if err := writef(cmd, "- %s: %d\n", count.Name, count.Count); err != nil {
			return err
		}
	}
	return nil
}

func printWorkspaceStatus(cmd *cobra.Command, report orient.StatusReport) error {
	if err := writeln(cmd, "repos:"); err != nil {
		return err
	}
	for _, repo := range report.Repos {
		if err := printRepoStatus(cmd, repo); err != nil {
			return err
		}
	}
	return nil
}

func printScenarioStatus(cmd *cobra.Command, report orient.ScenarioStatusReport) error {
	if err := writef(cmd, "scenario: %s\n", report.ScenarioID); err != nil {
		return err
	}
	if err := writeln(cmd, "repos:"); err != nil {
		return err
	}
	for _, repo := range report.Repos {
		line := fmt.Sprintf("- [%s] %s pinned=%s current=%s branch=%s", repo.ScenarioStatus, repo.RepoID, valueOrDash(repo.PinnedCommit), valueOrDash(repo.CurrentCommit), branchLabel(repo))
		if repo.ScenarioReason != "" {
			line += " reason=" + repo.ScenarioReason
		}
		if err := writeln(cmd, line); err != nil {
			return err
		}
	}
	return nil
}

func printRepoStatus(cmd *cobra.Command, repo orient.RepoStatus) error {
	line := fmt.Sprintf("- %s binding=%s git=%s branch=%s commit=%s dirty=%s untracked=%s upstream=%s ahead=%s behind=%s",
		repo.RepoID,
		repo.BindingStatus,
		repo.GitStatus,
		branchLabel(repo),
		valueOrDash(repo.Commit),
		intOrDash(repo.GitStatus == "ok", repo.DirtyFiles),
		intOrDash(repo.GitStatus == "ok", repo.UntrackedFiles),
		repo.Upstream,
		intOrDash(repo.HasDivergence, repo.Ahead),
		intOrDash(repo.HasDivergence, repo.Behind),
	)
	if repo.Reason != "" {
		line += " reason=" + repo.Reason
	}
	return writeln(cmd, line)
}

func branchLabel(repo orient.RepoStatus) string {
	if repo.Detached {
		return "detached"
	}
	return valueOrDash(repo.Branch)
}

func valueOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func intOrDash(ok bool, value int) string {
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%d", value)
}

func scenarioExitError(result scenario.RunResult) error {
	if result.Failed {
		return &ExitError{Code: 5}
	}
	if result.Drift || result.Blocked {
		return &ExitError{Code: 4}
	}
	return nil
}

func newHandoffCommand(workspaceFlag *string) *cobra.Command {
	var scenarioID string
	cmd := &cobra.Command{
		Use:   "handoff <change-id>",
		Short: "Render a markdown handoff summary for a change",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			out, err := handoff.Markdown(root, args[0], handoff.Options{ScenarioID: scenarioID})
			if err != nil {
				return err
			}
			return write(cmd, out)
		},
	}
	cmd.Flags().StringVar(&scenarioID, "scenario", "", "scenario id to include")
	return cmd
}

func newInstallCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Plan and apply derived guidance outputs",
	}
	cmd.AddCommand(newInstallShowTargetsCommand(workspaceFlag))
	cmd.AddCommand(newInstallPlanCommand(workspaceFlag))
	cmd.AddCommand(newInstallDiffCommand(workspaceFlag))
	cmd.AddCommand(newInstallApplyCommand(workspaceFlag))
	return cmd
}

type installFlags struct {
	scope    string
	userRoot string
	force    bool
	backup   bool
	dryRun   bool
	yes      bool
}

func addInstallPlanFlags(cmd *cobra.Command, flags *installFlags) {
	cmd.Flags().StringVar(&flags.scope, "scope", "repo", "install scope")
	cmd.Flags().StringVar(&flags.userRoot, "user-root", "", "user root for user-scope installs")
	cmd.Flags().BoolVar(&flags.force, "force", false, "allow overwriting changed targets")
	cmd.Flags().BoolVar(&flags.backup, "backup", false, "backup changed targets before overwriting")
}

func newInstallShowTargetsCommand(workspaceFlag *string) *cobra.Command {
	var flags installFlags
	cmd := &cobra.Command{
		Use:   "show-targets <tool> [repo-id]",
		Short: "Show install target paths",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			plan, err := buildInstallPlanFromArgs(workspaceFlag, args, flags)
			if err != nil {
				return err
			}
			if err := writef(cmd, "%s %s targets:\n", plan.Tool, plan.Scope); err != nil {
				return err
			}
			for _, target := range plan.Targets {
				if err := writef(cmd, "- %s %s\n", target.Kind, target.Path); err != nil {
					return err
				}
			}
			for _, note := range plan.Notes {
				if err := writef(cmd, "note: %s\n", note); err != nil {
					return err
				}
			}
			return nil
		},
	}
	addInstallPlanFlags(cmd, &flags)
	return cmd
}

func newInstallPlanCommand(workspaceFlag *string) *cobra.Command {
	var flags installFlags
	cmd := &cobra.Command{
		Use:   "plan <tool> [repo-id]",
		Short: "Plan derived guidance outputs without writing files",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			plan, err := buildInstallPlanFromArgs(workspaceFlag, args, flags)
			if err != nil {
				return err
			}
			return printInstallPlan(cmd, plan)
		},
	}
	addInstallPlanFlags(cmd, &flags)
	return cmd
}

func newInstallDiffCommand(workspaceFlag *string) *cobra.Command {
	var flags installFlags
	cmd := &cobra.Command{
		Use:   "diff <tool> [repo-id]",
		Short: "Show textual diffs for derived guidance outputs",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			repoID := ""
			if len(args) > 1 {
				repoID = args[1]
			}
			diff, err := install.BuildDiff(root, install.PlanOptions{
				Tool:     args[0],
				Scope:    install.Scope(flags.scope),
				RepoID:   repoID,
				UserRoot: flags.userRoot,
				Force:    flags.force,
				Backup:   flags.backup,
				Now:      time.Now(),
			})
			if err != nil {
				return err
			}
			if err := printInstallPlan(cmd, diff.Plan); err != nil {
				return err
			}
			if len(diff.Items) == 0 {
				return writeln(cmd, "diffs: none")
			}
			if err := writeln(cmd, "diffs:"); err != nil {
				return err
			}
			for _, item := range diff.Items {
				if err := writef(cmd, "### %s [%s]\n", item.Target.Path, item.Target.Status); err != nil {
					return err
				}
				for _, line := range item.Lines {
					if err := write(cmd, line); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	addInstallPlanFlags(cmd, &flags)
	return cmd
}

func newInstallApplyCommand(workspaceFlag *string) *cobra.Command {
	var flags installFlags
	cmd := &cobra.Command{
		Use:   "apply <tool> [repo-id]",
		Short: "Apply derived guidance outputs",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			now := time.Now()
			plan, err := buildInstallPlanFromArgsAt(workspaceFlag, args, flags, now)
			if err != nil {
				return err
			}
			if err := printInstallPlan(cmd, plan); err != nil {
				return err
			}
			if len(install.BlockedTargets(plan)) > 0 {
				return &ExitError{Code: 3}
			}
			if flags.dryRun {
				return writeln(cmd, "Dry run: no files were written.")
			}
			if !flags.yes {
				return fmt.Errorf("refusing to write without --yes; use --dry-run to preview")
			}
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			result, err := install.Apply(root, install.PlanOptions{
				Tool:     args[0],
				Scope:    install.Scope(flags.scope),
				RepoID:   repoIDFromArgs(args),
				UserRoot: flags.userRoot,
				Force:    flags.force,
				Backup:   flags.backup,
				Now:      now,
			})
			if err != nil {
				if len(install.BlockedTargets(result.Plan)) > 0 {
					return &ExitError{Code: 3, Err: err}
				}
				return err
			}
			for _, target := range result.Written {
				if err := writef(cmd, "written: %s\n", target.Path); err != nil {
					return err
				}
				if target.BackupPath != "" {
					if err := writef(cmd, "backup: %s\n", target.BackupPath); err != nil {
						return err
					}
				}
			}
			for _, target := range result.Skipped {
				if err := writef(cmd, "unchanged: %s\n", target.Path); err != nil {
					return err
				}
			}
			return nil
		},
	}
	addInstallPlanFlags(cmd, &flags)
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "preview apply without writing")
	cmd.Flags().BoolVar(&flags.yes, "yes", false, "confirm writes")
	return cmd
}

func buildInstallPlanFromArgs(workspaceFlag *string, args []string, flags installFlags) (install.Plan, error) {
	return buildInstallPlanFromArgsAt(workspaceFlag, args, flags, time.Now())
}

func buildInstallPlanFromArgsAt(workspaceFlag *string, args []string, flags installFlags, now time.Time) (install.Plan, error) {
	root, err := resolveWorkspaceRoot(workspaceFlag)
	if err != nil {
		return install.Plan{}, err
	}
	return install.BuildPlan(root, install.PlanOptions{
		Tool:     args[0],
		Scope:    install.Scope(flags.scope),
		RepoID:   repoIDFromArgs(args),
		UserRoot: flags.userRoot,
		Force:    flags.force,
		Backup:   flags.backup,
		Now:      now,
	})
}

func repoIDFromArgs(args []string) string {
	if len(args) > 1 {
		return args[1]
	}
	return ""
}

func printInstallPlan(cmd *cobra.Command, plan install.Plan) error {
	if err := writef(cmd, "tool: %s\n", plan.Tool); err != nil {
		return err
	}
	if err := writef(cmd, "scope: %s\n", plan.Scope); err != nil {
		return err
	}
	if err := writeln(cmd, "targets:"); err != nil {
		return err
	}
	for _, target := range plan.Targets {
		line := fmt.Sprintf("- [%s] %s %s (source: %s, ownership: %s)", target.Status, target.Kind, target.Path, target.Source, target.Ownership)
		if target.BackupPath != "" {
			line += " backup: " + target.BackupPath
		}
		if len(target.Notes) > 0 {
			line += " note: " + strings.Join(target.Notes, "; ")
		}
		if err := writeln(cmd, line); err != nil {
			return err
		}
	}
	if len(plan.Summary) > 0 {
		if err := writeln(cmd, "summary:"); err != nil {
			return err
		}
		for _, key := range install.SummaryKeys() {
			if count := plan.Summary[key]; count > 0 {
				if err := writef(cmd, "- %s: %d\n", key, count); err != nil {
					return err
				}
			}
		}
	}
	for _, note := range plan.Notes {
		if err := writef(cmd, "note: %s\n", note); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	cobra.EnableCommandSorting = false
}
