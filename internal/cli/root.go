package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/install"
	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
	"github.com/johnkil/polyrepo-workspace-kit/internal/validate"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"

	"github.com/spf13/cobra"
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
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Err != nil {
				_, _ = fmt.Fprintln(root.ErrOrStderr(), exitErr.Err)
			}
			if exitErr.Code == 0 {
				return 1
			}
			return exitErr.Code
		}
		_, _ = fmt.Fprintln(root.ErrOrStderr(), err)
		return 1
	}
	return 0
}

func newRootCommand() *cobra.Command {
	var workspaceFlag string

	root := &cobra.Command{
		Use:           "wkit",
		Short:         "Coordinate local polyrepo workspaces",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&workspaceFlag, "workspace", "", "workspace root")

	root.AddCommand(newInitCommand())
	root.AddCommand(newRepoCommand(&workspaceFlag))
	root.AddCommand(newBindCommand(&workspaceFlag))
	root.AddCommand(newChangeCommand(&workspaceFlag))
	root.AddCommand(newScenarioCommand(&workspaceFlag))
	root.AddCommand(newInstallCommand(&workspaceFlag))
	root.AddCommand(newValidateCommand(&workspaceFlag))

	return root
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

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init <path>",
		Short: "Initialize a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := workspace.Init(args[0]); err != nil {
				return err
			}
			root, err := workspace.FindRoot(args[0])
			if err != nil {
				return err
			}
			return writef(cmd, "initialized workspace at %s\n", root)
		},
	}
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

func newScenarioCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scenario",
		Short: "Manage reviewable validation snapshots",
	}
	cmd.AddCommand(newScenarioPinCommand(workspaceFlag))
	cmd.AddCommand(newScenarioShowCommand(workspaceFlag))
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
			return scenarioExitError(result)
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

func scenarioExitError(result scenario.RunResult) error {
	if result.Failed {
		return &ExitError{Code: 5}
	}
	if result.Drift || result.Blocked {
		return &ExitError{Code: 4}
	}
	return nil
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
