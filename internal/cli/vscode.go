package cli

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	vscodeworkspace "github.com/johnkil/polyrepo-workspace-kit/internal/vscode"

	"github.com/spf13/cobra"
)

type vscodeFlags struct {
	force  bool
	backup bool
	dryRun bool
	yes    bool
}

func newVSCodeCommand(workspaceFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vscode",
		Short: "Generate a local VS Code multi-root workspace",
	}
	cmd.AddCommand(newVSCodePlanCommand(workspaceFlag))
	cmd.AddCommand(newVSCodeDiffCommand(workspaceFlag))
	cmd.AddCommand(newVSCodeApplyCommand(workspaceFlag))
	cmd.AddCommand(newVSCodeOpenCommand(workspaceFlag))
	return cmd
}

func addVSCodeWriteFlags(cmd *cobra.Command, flags *vscodeFlags) {
	cmd.Flags().BoolVar(&flags.force, "force", false, "allow overwriting a changed workspace file")
	cmd.Flags().BoolVar(&flags.backup, "backup", false, "backup a changed workspace file before overwriting")
}

func newVSCodePlanCommand(workspaceFlag *string) *cobra.Command {
	var flags vscodeFlags
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan the local VS Code workspace file without writing",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			plan, err := vscodeworkspace.BuildPlan(root, vscodeworkspace.PlanOptions{
				Force:  flags.force,
				Backup: flags.backup,
				Now:    time.Now(),
			})
			if err != nil {
				return err
			}
			return printVSCodePlan(cmd, plan)
		},
	}
	addVSCodeWriteFlags(cmd, &flags)
	return cmd
}

func newVSCodeDiffCommand(workspaceFlag *string) *cobra.Command {
	var flags vscodeFlags
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show the textual diff for the local VS Code workspace file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			diff, err := vscodeworkspace.BuildDiff(root, vscodeworkspace.PlanOptions{
				Force:  flags.force,
				Backup: flags.backup,
				Now:    time.Now(),
			})
			if err != nil {
				return err
			}
			if err := printVSCodePlan(cmd, diff.Plan); err != nil {
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
	addVSCodeWriteFlags(cmd, &flags)
	return cmd
}

func newVSCodeApplyCommand(workspaceFlag *string) *cobra.Command {
	var flags vscodeFlags
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Write the local VS Code workspace file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			now := time.Now()
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			plan, err := vscodeworkspace.BuildPlan(root, vscodeworkspace.PlanOptions{
				Force:  flags.force,
				Backup: flags.backup,
				Now:    now,
			})
			if err != nil {
				return err
			}
			if err := printVSCodePlan(cmd, plan); err != nil {
				return err
			}
			if len(vscodeworkspace.BlockedTargets(plan)) > 0 {
				return &ExitError{Code: 3}
			}
			if flags.dryRun {
				return writeln(cmd, "Dry run: no files were written.")
			}
			if !flags.yes {
				return fmt.Errorf("refusing to write without --yes; use --dry-run to preview")
			}
			result, err := vscodeworkspace.Apply(root, vscodeworkspace.PlanOptions{
				Force:  flags.force,
				Backup: flags.backup,
				Now:    now,
			})
			if err != nil {
				if len(vscodeworkspace.BlockedTargets(result.Plan)) > 0 {
					return &ExitError{Code: 3, Err: err}
				}
				return err
			}
			return printVSCodeApplyResult(cmd, result)
		},
	}
	addVSCodeWriteFlags(cmd, &flags)
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "preview apply without writing")
	cmd.Flags().BoolVar(&flags.yes, "yes", false, "confirm writes")
	return cmd
}

func newVSCodeOpenCommand(workspaceFlag *string) *cobra.Command {
	var flags vscodeFlags
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open the generated VS Code workspace file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			now := time.Now()
			root, err := resolveWorkspaceRoot(workspaceFlag)
			if err != nil {
				return err
			}
			plan, err := vscodeworkspace.BuildPlan(root, vscodeworkspace.PlanOptions{
				Force:  flags.force,
				Backup: flags.backup,
				Now:    now,
			})
			if err != nil {
				return err
			}
			if len(vscodeworkspace.BlockedTargets(plan)) > 0 {
				if err := printVSCodePlan(cmd, plan); err != nil {
					return err
				}
				return &ExitError{Code: 3}
			}
			if len(plan.Targets) != 1 {
				return fmt.Errorf("unexpected VS Code plan target count: %d", len(plan.Targets))
			}
			target := plan.Targets[0]
			if target.Status != vscodeworkspace.StatusUnchanged {
				if !flags.yes {
					return fmt.Errorf("VS Code workspace file is not current; run `wkit vscode apply --yes` first or pass `wkit vscode open --yes`")
				}
				result, err := vscodeworkspace.Apply(root, vscodeworkspace.PlanOptions{
					Force:  flags.force,
					Backup: flags.backup,
					Now:    now,
				})
				if err != nil {
					return err
				}
				if err := printVSCodeApplyResult(cmd, result); err != nil {
					return err
				}
				target = result.Plan.Targets[0]
			}
			open := exec.Command("code", target.Path)
			open.Stdout = cmd.OutOrStdout()
			open.Stderr = cmd.ErrOrStderr()
			if err := open.Run(); err != nil {
				return fmt.Errorf("failed to run `code %s`: %w", target.Path, err)
			}
			return writef(cmd, "opened: %s\n", target.Path)
		},
	}
	addVSCodeWriteFlags(cmd, &flags)
	cmd.Flags().BoolVar(&flags.yes, "yes", false, "write or update the workspace file before opening when needed")
	return cmd
}

func printVSCodePlan(cmd *cobra.Command, plan vscodeworkspace.Plan) error {
	if err := writeln(cmd, "target: vscode"); err != nil {
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
	for _, note := range plan.Notes {
		if err := writef(cmd, "note: %s\n", note); err != nil {
			return err
		}
	}
	return printVSCodeSummary(cmd, plan)
}

func printVSCodeSummary(cmd *cobra.Command, plan vscodeworkspace.Plan) error {
	keys := []string{
		vscodeworkspace.StatusNew,
		vscodeworkspace.StatusUnchanged,
		vscodeworkspace.StatusBlocked,
		vscodeworkspace.StatusOverwrite,
		vscodeworkspace.StatusBackupOverwrite,
	}
	var parts []string
	for _, key := range keys {
		if count := plan.Summary[key]; count > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", key, count))
		}
	}
	if len(parts) == 0 {
		parts = append(parts, "none=0")
	}
	return writef(cmd, "summary: %s\n", strings.Join(parts, " "))
}

func printVSCodeApplyResult(cmd *cobra.Command, result vscodeworkspace.ApplyResult) error {
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
}
