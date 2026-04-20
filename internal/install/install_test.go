package install_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/install"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

func TestPortableRepoPlan(t *testing.T) {
	root, checkout := seedWorkspace(t)

	plan, err := install.BuildPlan(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Now:    time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Tool != "portable" || plan.Scope != install.ScopeRepo {
		t.Fatalf("unexpected plan metadata: %#v", plan)
	}
	assertTarget(t, plan, filepath.Join(checkout, "AGENTS.md"), "instructions", install.StatusNew)
	assertTarget(t, plan, filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md"), "skill", install.StatusNew)
}

func TestPortableUserPlanIsSkillsOnly(t *testing.T) {
	root, _ := seedWorkspace(t)
	userRoot := t.TempDir()

	plan, err := install.BuildPlan(root, install.PlanOptions{
		Tool:     "portable",
		Scope:    install.ScopeUser,
		UserRoot: userRoot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Targets) != 1 {
		t.Fatalf("expected one skill target, got %#v", plan.Targets)
	}
	assertTarget(t, plan, filepath.Join(userRoot, ".agents", "skills", "release-note", "SKILL.md"), "skill", install.StatusNew)
}

func TestRepoScopeAdapterPlansMatchSpecTargets(t *testing.T) {
	root, checkout := seedWorkspace(t)

	codex, err := install.BuildPlan(root, install.PlanOptions{Tool: "codex", Scope: install.ScopeRepo, RepoID: "app-web"})
	if err != nil {
		t.Fatal(err)
	}
	assertToolTarget(t, codex, filepath.Join(checkout, "AGENTS.md"), "codex", "instructions", install.StatusNew)
	assertToolTarget(t, codex, filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md"), "codex", "skill", install.StatusNew)

	opencode, err := install.BuildPlan(root, install.PlanOptions{Tool: "opencode", Scope: install.ScopeRepo, RepoID: "app-web"})
	if err != nil {
		t.Fatal(err)
	}
	assertToolTarget(t, opencode, filepath.Join(checkout, "AGENTS.md"), "opencode", "instructions", install.StatusNew)
	assertToolTarget(t, opencode, filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md"), "opencode", "skill", install.StatusNew)

	copilot, err := install.BuildPlan(root, install.PlanOptions{Tool: "copilot", Scope: install.ScopeRepo, RepoID: "app-web"})
	if err != nil {
		t.Fatal(err)
	}
	if len(copilot.Targets) != 1 {
		t.Fatalf("copilot should plan exactly one repo-scope target, got %#v", copilot.Targets)
	}
	assertToolTarget(t, copilot, filepath.Join(checkout, ".github", "copilot-instructions.md"), "copilot", "instructions", install.StatusNew)

	claude, err := install.BuildPlan(root, install.PlanOptions{Tool: "claude", Scope: install.ScopeRepo, RepoID: "app-web"})
	if err != nil {
		t.Fatal(err)
	}
	if len(claude.Targets) != 2 {
		t.Fatalf("claude should plan CLAUDE.md plus one skill, got %#v", claude.Targets)
	}
	assertToolTarget(t, claude, filepath.Join(checkout, "CLAUDE.md"), "claude", "instructions", install.StatusNew)
	assertToolTarget(t, claude, filepath.Join(checkout, ".claude", "skills", "release-note", "SKILL.md"), "claude", "skill", install.StatusNew)
}

func TestToolSpecificUserScopeIsOutOfScope(t *testing.T) {
	root, _ := seedWorkspace(t)
	for _, tool := range []string{"codex", "opencode", "copilot", "claude"} {
		_, err := install.BuildPlan(root, install.PlanOptions{
			Tool:   tool,
			Scope:  install.ScopeUser,
			RepoID: "",
		})
		if err == nil {
			t.Fatalf("expected %s user scope to be rejected", tool)
		}
	}
}

func TestPortablePlanMarksBlockedAndBackupOverwrite(t *testing.T) {
	root, checkout := seedWorkspace(t)
	target := filepath.Join(checkout, "AGENTS.md")
	if err := os.WriteFile(target, []byte("custom\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	blocked, err := install.BuildPlan(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertTarget(t, blocked, target, "instructions", install.StatusBlocked)

	backedUp, err := install.BuildPlan(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Backup: true,
		Now:    time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	found := findTarget(t, backedUp, target)
	if found.Status != install.StatusBackupOverwrite {
		t.Fatalf("expected backup+overwrite, got %s", found.Status)
	}
	if found.BackupPath != target+".bak.20260419T120000Z" {
		t.Fatalf("unexpected backup path: %s", found.BackupPath)
	}
}

func TestPortableDiffShowsTextualChanges(t *testing.T) {
	root, checkout := seedWorkspace(t)
	target := filepath.Join(checkout, "AGENTS.md")
	if err := os.WriteFile(target, []byte("custom override\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	diff, err := install.BuildDiff(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Items) == 0 {
		t.Fatal("expected diff items")
	}
	joined := strings.Join(diff.Items[0].Lines, "")
	if !strings.Contains(joined, "-custom override") {
		t.Fatalf("expected current content in diff:\n%s", joined)
	}
	if !strings.Contains(joined, "+# AGENTS.md") {
		t.Fatalf("expected generated content in diff:\n%s", joined)
	}
}

func TestAdapterDiffsShowRenderedInstructionContent(t *testing.T) {
	root, _ := seedWorkspace(t)

	cases := []struct {
		tool string
		want string
	}{
		{tool: "codex", want: "+# AGENTS.md"},
		{tool: "opencode", want: "+# AGENTS.md"},
		{tool: "copilot", want: "+# Copilot instructions for app-web"},
		{tool: "claude", want: "+# Claude instructions for app-web"},
	}
	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			diff, err := install.BuildDiff(root, install.PlanOptions{
				Tool:   tc.tool,
				Scope:  install.ScopeRepo,
				RepoID: "app-web",
			})
			if err != nil {
				t.Fatal(err)
			}
			joined := diffText(diff)
			if !strings.Contains(joined, "+<!-- generated by wkit -->") {
				t.Fatalf("expected ownership marker in diff:\n%s", joined)
			}
			if !strings.Contains(joined, tc.want) {
				t.Fatalf("expected %q in diff:\n%s", tc.want, joined)
			}
		})
	}
}

func TestPortableApplyWritesTargets(t *testing.T) {
	root, checkout := seedWorkspace(t)

	result, err := install.Apply(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Written) != 2 {
		t.Fatalf("expected 2 written targets, got %#v", result.Written)
	}
	if _, err := os.Stat(filepath.Join(checkout, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
}

func TestAdapterApplyWritesOnlySpecTargets(t *testing.T) {
	t.Run("codex", func(t *testing.T) {
		root, checkout := seedWorkspace(t)
		applyTool(t, root, "codex")
		assertFileExists(t, filepath.Join(checkout, "AGENTS.md"))
		assertFileExists(t, filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md"))
	})

	t.Run("opencode", func(t *testing.T) {
		root, checkout := seedWorkspace(t)
		applyTool(t, root, "opencode")
		assertFileExists(t, filepath.Join(checkout, "AGENTS.md"))
		assertFileExists(t, filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md"))
	})

	t.Run("copilot", func(t *testing.T) {
		root, checkout := seedWorkspace(t)
		applyTool(t, root, "copilot")
		assertFileExists(t, filepath.Join(checkout, ".github", "copilot-instructions.md"))
		assertFileMissing(t, filepath.Join(checkout, "AGENTS.md"))
		assertFileMissing(t, filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md"))
	})

	t.Run("claude", func(t *testing.T) {
		root, checkout := seedWorkspace(t)
		applyTool(t, root, "claude")
		assertFileExists(t, filepath.Join(checkout, "CLAUDE.md"))
		assertFileExists(t, filepath.Join(checkout, ".claude", "skills", "release-note", "SKILL.md"))
		assertFileMissing(t, filepath.Join(checkout, "AGENTS.md"))
		assertFileMissing(t, filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md"))
	})
}

func TestRepoScopeAdaptersBlockSymlinkParentEscapes(t *testing.T) {
	cases := []struct {
		tool        string
		linkName    string
		targetRel   string
		outsideRel  string
		blockedKind string
	}{
		{
			tool:        "portable",
			linkName:    ".agents",
			targetRel:   filepath.Join(".agents", "skills", "release-note", "SKILL.md"),
			outsideRel:  filepath.Join("skills", "release-note", "SKILL.md"),
			blockedKind: "skill",
		},
		{
			tool:        "codex",
			linkName:    ".agents",
			targetRel:   filepath.Join(".agents", "skills", "release-note", "SKILL.md"),
			outsideRel:  filepath.Join("skills", "release-note", "SKILL.md"),
			blockedKind: "skill",
		},
		{
			tool:        "opencode",
			linkName:    ".agents",
			targetRel:   filepath.Join(".agents", "skills", "release-note", "SKILL.md"),
			outsideRel:  filepath.Join("skills", "release-note", "SKILL.md"),
			blockedKind: "skill",
		},
		{
			tool:        "copilot",
			linkName:    ".github",
			targetRel:   filepath.Join(".github", "copilot-instructions.md"),
			outsideRel:  "copilot-instructions.md",
			blockedKind: "instructions",
		},
		{
			tool:        "claude",
			linkName:    ".claude",
			targetRel:   filepath.Join(".claude", "skills", "release-note", "SKILL.md"),
			outsideRel:  filepath.Join("skills", "release-note", "SKILL.md"),
			blockedKind: "skill",
		},
	}

	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			root, checkout := seedWorkspace(t)
			outside := t.TempDir()
			if err := os.Symlink(outside, filepath.Join(checkout, tc.linkName)); err != nil {
				t.Fatal(err)
			}

			plan, err := install.BuildPlan(root, install.PlanOptions{
				Tool:   tc.tool,
				Scope:  install.ScopeRepo,
				RepoID: "app-web",
			})
			if err != nil {
				t.Fatal(err)
			}
			target := findTarget(t, plan, filepath.Join(checkout, tc.targetRel))
			if target.Status != install.StatusBlocked || target.Kind != tc.blockedKind {
				t.Fatalf("expected blocked %s target, got %#v", tc.blockedKind, target)
			}
			assertTargetNoteContains(t, target, "unsafe target path")

			_, err = install.Apply(root, install.PlanOptions{
				Tool:   tc.tool,
				Scope:  install.ScopeRepo,
				RepoID: "app-web",
			})
			if err == nil {
				t.Fatal("expected apply to fail on unsafe symlink parent")
			}
			assertFileMissing(t, filepath.Join(outside, tc.outsideRel))
		})
	}
}

func TestInstallDiffDoesNotReadSymlinkedExistingTarget(t *testing.T) {
	root, checkout := seedWorkspace(t)
	secretPath := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secretPath, []byte("do-not-leak\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	targetPath := filepath.Join(checkout, "AGENTS.md")
	if err := os.Symlink(secretPath, targetPath); err != nil {
		t.Fatal(err)
	}

	diff, err := install.BuildDiff(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Backup: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	target := findTarget(t, diff.Plan, targetPath)
	if target.Status != install.StatusBlocked {
		t.Fatalf("expected symlinked target to be blocked, got %#v", target)
	}
	assertTargetNoteContains(t, target, "symlink")
	if joined := diffText(diff); strings.Contains(joined, "do-not-leak") {
		t.Fatalf("diff leaked symlinked target content:\n%s", joined)
	}

	_, err = install.Apply(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Backup: true,
	})
	if err == nil {
		t.Fatal("expected apply to fail on symlinked target")
	}
	data, err := os.ReadFile(secretPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "do-not-leak\n" {
		t.Fatalf("secret target was changed: %q", string(data))
	}
}

func TestInstallDiffDoesNotReadSymlinkedSkillSource(t *testing.T) {
	root, checkout := seedWorkspace(t)
	secretPath := filepath.Join(t.TempDir(), "secret-skill.md")
	if err := os.WriteFile(secretPath, []byte("skill secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(root, "guidance", "skills", "release-note", "SKILL.md")
	if err := os.Remove(sourcePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secretPath, sourcePath); err != nil {
		t.Fatal(err)
	}

	diff, err := install.BuildDiff(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
	})
	if err != nil {
		t.Fatal(err)
	}
	target := findTarget(t, diff.Plan, filepath.Join(checkout, ".agents", "skills", "release-note", "SKILL.md"))
	if target.Status != install.StatusBlocked {
		t.Fatalf("expected symlinked source to be blocked, got %#v", target)
	}
	assertTargetNoteContains(t, target, "unsafe source path")
	if joined := diffText(diff); strings.Contains(joined, "skill secret") {
		t.Fatalf("diff leaked symlinked source content:\n%s", joined)
	}
}

func TestRenderedRulesRejectSymlinkedRuleSource(t *testing.T) {
	root, _ := seedWorkspace(t)
	secretPath := filepath.Join(t.TempDir(), "secret-rule.md")
	if err := os.WriteFile(secretPath, []byte("rule secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rulePath := filepath.Join(root, "guidance", "rules", "always-on.md")
	if err := os.Remove(rulePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secretPath, rulePath); err != nil {
		t.Fatal(err)
	}

	_, err := install.BuildPlan(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
	})
	if err == nil {
		t.Fatal("expected symlinked guidance rule to be rejected")
	}
	if strings.Contains(err.Error(), "rule secret") {
		t.Fatalf("error leaked rule content: %v", err)
	}
}

func TestInstallRejectsSymlinkedGuidanceSourceRoots(t *testing.T) {
	t.Run("rules", func(t *testing.T) {
		root, _ := seedWorkspace(t)
		outside := t.TempDir()
		if err := os.WriteFile(filepath.Join(outside, "always-on.md"), []byte("root rule secret\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		rulesRoot := filepath.Join(root, "guidance", "rules")
		if err := os.RemoveAll(rulesRoot); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, rulesRoot); err != nil {
			t.Fatal(err)
		}

		_, err := install.BuildPlan(root, install.PlanOptions{
			Tool:   "portable",
			Scope:  install.ScopeRepo,
			RepoID: "app-web",
		})
		if err == nil {
			t.Fatal("expected symlinked guidance rules root to be rejected")
		}
		if strings.Contains(err.Error(), "root rule secret") {
			t.Fatalf("error leaked rule root content: %v", err)
		}
	})

	t.Run("skills", func(t *testing.T) {
		root, _ := seedWorkspace(t)
		outside := filepath.Join(t.TempDir(), "skills")
		outsideSkill := filepath.Join(outside, "release-note")
		if err := os.MkdirAll(outsideSkill, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(outsideSkill, "SKILL.md"), []byte("root skill secret\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		skillsRoot := filepath.Join(root, "guidance", "skills")
		if err := os.RemoveAll(skillsRoot); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, skillsRoot); err != nil {
			t.Fatal(err)
		}

		_, err := install.BuildPlan(root, install.PlanOptions{
			Tool:   "portable",
			Scope:  install.ScopeRepo,
			RepoID: "app-web",
		})
		if err == nil {
			t.Fatal("expected symlinked guidance skills root to be rejected")
		}
		if strings.Contains(err.Error(), "root skill secret") {
			t.Fatalf("error leaked skill root content: %v", err)
		}
	})
}

func TestPortableApplyBlocksWithoutForceOrBackup(t *testing.T) {
	root, checkout := seedWorkspace(t)
	target := filepath.Join(checkout, "AGENTS.md")
	if err := os.WriteFile(target, []byte("custom override\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := install.Apply(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
	})
	if err == nil {
		t.Fatal("expected blocked apply error")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "custom override\n" {
		t.Fatalf("target was overwritten: %q", string(data))
	}
}

func TestPortableApplyBackupOverwrite(t *testing.T) {
	root, checkout := seedWorkspace(t)
	target := filepath.Join(checkout, "AGENTS.md")
	if err := os.WriteFile(target, []byte("custom override\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := install.Apply(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Backup: true,
		Now:    time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Written) == 0 {
		t.Fatal("expected writes")
	}
	backupPath := target + ".bak.20260419T120000Z"
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != "custom override\n" {
		t.Fatalf("unexpected backup content: %q", string(backup))
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "generated by wkit") {
		t.Fatalf("expected generated content, got %q", string(data))
	}
}

func TestPortableApplyBackupOverwriteAvoidsExistingBackupPath(t *testing.T) {
	root, checkout := seedWorkspace(t)
	target := filepath.Join(checkout, "AGENTS.md")
	if err := os.WriteFile(target, []byte("custom override\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	existingBackup := target + ".bak.20260419T120000Z"
	if err := os.WriteFile(existingBackup, []byte("older backup\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := install.Apply(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Backup: true,
		Now:    time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Written) == 0 {
		t.Fatal("expected writes")
	}
	backupPath := target + ".bak.20260419T120000Z.001"
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != "custom override\n" {
		t.Fatalf("unexpected backup content: %q", string(backup))
	}
	existing, err := os.ReadFile(existingBackup)
	if err != nil {
		t.Fatal(err)
	}
	if string(existing) != "older backup\n" {
		t.Fatalf("existing backup was overwritten: %q", string(existing))
	}
}

func TestPortableApplyForceOverwrite(t *testing.T) {
	root, checkout := seedWorkspace(t)
	target := filepath.Join(checkout, "AGENTS.md")
	if err := os.WriteFile(target, []byte("custom override\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := install.Apply(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Force:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Written) == 0 {
		t.Fatal("expected writes")
	}
	if _, err := os.Stat(target + ".bak.20260419T120000Z"); err == nil {
		t.Fatal("force overwrite should not create a backup")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "generated by wkit") {
		t.Fatalf("expected generated content, got %q", string(data))
	}
}

func TestPortableApplyPreservesUnplannedSkillFiles(t *testing.T) {
	root, checkout := seedWorkspace(t)
	skillDir := filepath.Join(checkout, ".agents", "skills", "release-note")
	target := filepath.Join(skillDir, "SKILL.md")
	extra := filepath.Join(skillDir, "notes.md")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("local skill override\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(extra, []byte("keep this local note\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := install.BuildPlan(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Force:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	found := findTarget(t, plan, target)
	if found.Status != install.StatusOverwrite {
		t.Fatalf("expected skill target overwrite, got %s", found.Status)
	}

	if _, err := install.Apply(root, install.PlanOptions{
		Tool:   "portable",
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
		Force:  true,
	}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Draft release notes.") {
		t.Fatalf("expected canonical skill content, got %q", string(data))
	}
	extraData, err := os.ReadFile(extra)
	if err != nil {
		t.Fatal(err)
	}
	if string(extraData) != "keep this local note\n" {
		t.Fatalf("unplanned skill file changed: %q", string(extraData))
	}
}

func seedWorkspace(t *testing.T) (string, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace")
	checkout := filepath.Join(t.TempDir(), "app-web")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RegisterRepo(root, "app-web", "app"); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.SetBinding(root, "app-web", checkout); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(root, "guidance", "skills", "release-note")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: release-note\n---\nDraft release notes.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, checkout
}

func assertTarget(t *testing.T, plan install.Plan, path string, kind string, status string) {
	t.Helper()
	target := findTarget(t, plan, path)
	if target.Kind != kind {
		t.Fatalf("target %s kind: expected %s, got %s", path, kind, target.Kind)
	}
	if target.Status != status {
		t.Fatalf("target %s status: expected %s, got %s", path, status, target.Status)
	}
}

func assertToolTarget(t *testing.T, plan install.Plan, path string, tool string, kind string, status string) {
	t.Helper()
	target := findTarget(t, plan, path)
	if target.Tool != tool {
		t.Fatalf("target %s tool: expected %s, got %s", path, tool, target.Tool)
	}
	if target.Kind != kind {
		t.Fatalf("target %s kind: expected %s, got %s", path, kind, target.Kind)
	}
	if target.Status != status {
		t.Fatalf("target %s status: expected %s, got %s", path, status, target.Status)
	}
}

func assertTargetNoteContains(t *testing.T, target install.Target, want string) {
	t.Helper()
	for _, note := range target.Notes {
		if strings.Contains(note, want) {
			return
		}
	}
	t.Fatalf("expected target note containing %q, got %#v", want, target.Notes)
}

func findTarget(t *testing.T, plan install.Plan, path string) install.Target {
	t.Helper()
	for _, target := range plan.Targets {
		if target.Path == path {
			return target
		}
	}
	t.Fatalf("target not found: %s in %#v", path, plan.Targets)
	return install.Target{}
}

func diffText(diff install.DiffPlan) string {
	var lines []string
	for _, item := range diff.Items {
		lines = append(lines, item.Lines...)
	}
	return strings.Join(lines, "")
}

func applyTool(t *testing.T, root string, tool string) {
	t.Helper()
	result, err := install.Apply(root, install.PlanOptions{
		Tool:   tool,
		Scope:  install.ScopeRepo,
		RepoID: "app-web",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Written) == 0 {
		t.Fatalf("expected %s to write targets", tool)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s: %v", path, err)
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected file to be missing: %s", path)
	}
}
