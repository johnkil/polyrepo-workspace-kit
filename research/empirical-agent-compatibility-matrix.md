# Empirical Agent Compatibility Matrix

Date: 2026-04-19
Scope: observed local behavior plus docs-backed gaps for Codex, Claude Code, OpenCode, and GitHub Copilot.

## Purpose

This file exists because documentation is not enough for adapter design. Agent discovery behavior is often more subtle than the docs imply:

- nested instruction precedence;
- user-scope versus repo-scope loading;
- skill discovery paths;
- cloud versus CLI versus IDE behavior;
- generated-file conflicts;
- settings and feature flags.

This is a living matrix. Every empirical claim should include date, tool version, probe method, and confidence.

## Local Environment

Observed local tools:

| Tool | Local availability | Observed version | Notes |
| --- | --- | --- | --- |
| Codex CLI | Installed | `codex-cli 0.117.0` | Non-interactive `codex exec` available |
| Claude Code | Installed | `2.1.90 (Claude Code)` | Non-interactive `claude -p` available; default model attempted 1M context and required `--model sonnet` for this pass |
| GitHub CLI | Installed | `gh 2.86.0` | `gh-copilot` extension not installed |
| GitHub Copilot CLI | Not found as standalone command | Not tested | Could be installed through GitHub/Copilot surfaces later |
| OpenCode | Not found | Not tested | Requires separate install before empirical test |

## Probe Method

A clean temporary Git repository fixture was created and then removed. It contained:

- root `AGENTS.md`;
- nested `AGENTS.md`;
- root `CLAUDE.md`;
- nested `CLAUDE.md`;
- `.codex/skills/wkit-probe/SKILL.md`;
- `.claude/skills/wkit-probe/SKILL.md`;
- `.agents/skills/wkit-probe/SKILL.md`.

The probes asked each CLI to answer using sentinel tokens from automatically loaded instructions, without intentionally inspecting files or running commands. Codex was run with `--sandbox read-only --ephemeral`; Claude was run with `--permission-mode dontAsk --tools "" --no-session-persistence --model sonnet`.

This is a limited behavioral test, not a full certification.

Limitations:

- user-level instruction directories were not remapped to a temporary home;
- Copilot and OpenCode were not run locally;
- Codex emitted unrelated user-config/state warnings during the run, including an invalid user skill outside the fixture;
- project skill behavior may depend on project root, Git root, config, trigger phrasing, slash invocation, or version;
- results can change with tool updates.

## Observed Results

| Tool | Probe | Result | Interpretation | Confidence |
| --- | --- | --- | --- | --- |
| Codex CLI 0.117.0 | Clean Git repo; run from nested folder with root and nested `AGENTS.md` | Returned `WKIT_ROOT_AGENTS_LOADED` and `WKIT_NESTED_AGENTS_LOADED` | Codex loaded both ancestor and nested `AGENTS.md` files in this clean-repo probe. | Medium-high |
| Codex CLI 0.117.0 | Clean Git repo; prompted to use local `.codex/skills/wkit-probe/SKILL.md` | Returned `NONE` | This probe did not demonstrate project skill discovery for `.codex/skills` in the fixture. | Low-medium |
| Claude Code 2.1.90 | Clean Git repo; run from nested folder with root and nested `CLAUDE.md` | Returned `WKIT_ROOT_CLAUDE_LOADED` and `WKIT_NESTED_CLAUDE_LOADED` | Claude loaded both ancestor and nested `CLAUDE.md` files in this probe. | Medium-high |
| Claude Code 2.1.90 | Clean Git repo; natural-language prompt to use `.claude/skills/wkit-probe/SKILL.md` | Returned `WKIT_ROOT_CLAUDE_LOADED` and `NONE` | Natural-language trigger did not demonstrate skill activation in this constrained probe. | Medium |
| Claude Code 2.1.90 | Clean Git repo; explicit `/wkit-probe` invocation for `.claude/skills/wkit-probe/SKILL.md` | Returned `WKIT_ROOT_CLAUDE_LOADED` and `WKIT_CLAUDE_SKILL_LOADED` | Claude discovered and loaded the project skill when invoked explicitly by slash-style skill name. | Medium-high |
| GitHub Copilot | Not run | Not empirical | Local `gh-copilot` extension and standalone `copilot` command were not installed. | None |
| OpenCode | Not run | Not empirical | `opencode` command was not installed. | None |

## Docs-Backed Compatibility Snapshot

This section is not empirical; it is included to define the next test targets.

| Tool | Repo instructions | Nested/path instructions | User instructions | Skills | Custom agents/subagents |
| --- | --- | --- | --- | --- | --- |
| Codex | `AGENTS.md` | Directory-scoped `AGENTS.md` per docs | User skills under Codex config paths per docs | `SKILL.md` skills | Subagents not a stable portable target for `wkit` v0 |
| Claude Code | `CLAUDE.md`, optional import from `AGENTS.md` | Ancestor and nested `CLAUDE.md` behavior; project rules also exist | `~/.claude/CLAUDE.md` | `.claude/skills/<name>/SKILL.md` | Subagents supported but tool-specific |
| GitHub Copilot | `.github/copilot-instructions.md`, `AGENTS.md` | `.github/instructions/*.instructions.md` with `applyTo`; nearest `AGENTS.md` behavior documented | Personal/org instructions depending on surface | Agent skills in newer Copilot surfaces | `.github/agents/*.md` and user-level agents in Copilot CLI/cloud |
| OpenCode | Agent config and markdown guidance | Project discovery depends on current worktree | Global config locations | `.opencode/skills`, `.claude/skills`, `.agents/skills` | Primary agents and subagents in OpenCode config |

Sources:

- [OpenAI Codex AGENTS.md](https://developers.openai.com/codex/guides/agents-md)
- [OpenAI Codex Skills](https://developers.openai.com/codex/skills)
- [Claude Code Memory](https://code.claude.com/docs/en/memory)
- [Claude Code Skills](https://code.claude.com/docs/en/skills)
- [GitHub Copilot repository instructions](https://docs.github.com/copilot/how-tos/configure-custom-instructions/add-repository-instructions)
- [GitHub Copilot custom agents](https://docs.github.com/en/copilot/concepts/agents/coding-agent/about-custom-agents)
- [OpenCode Skills](https://opencode.ai/docs/skills)
- [OpenCode Agents](https://opencode.ai/docs/agents)

## Adapter Implications For `wkit`

1. Treat discovery behavior as versioned compatibility, not timeless truth.

`wkit` should record adapter assumptions in tests and docs. A generated file target can be correct today and subtly wrong after a tool update.

2. Do not assume all tools merge ancestor instructions the same way.

The clean local probe observed both Codex and Claude loading ancestor plus nested instruction files, but earlier non-clean or version-sensitive probes differed. Treat this behavior as empirical for the exact version and fixture, not permanent truth.

3. Avoid conflicting generated instructions.

If a tool concatenates instructions, conflicts become dangerous. If a tool prioritizes nearest files, root guidance may be ignored. Generated output should be short and scoped.

4. Treat skills as tool-specific until empirically confirmed.

`SKILL.md` is converging, but discovery paths and activation behavior still differ. The clean local probe confirmed Claude skill loading through explicit `/wkit-probe`; it did not confirm Codex project skill discovery.

5. Build adapter tests into `wkit`.

The adapter layer should eventually have fixtures that can be run against installed CLIs:

```text
wkit compat probe codex
wkit compat probe claude
wkit compat probe opencode
wkit compat probe copilot
```

## Next Empirical Pass

The next pass should test:

- user-scope instructions using a temporary home directory;
- skill discovery from every supported path;
- generated-file overwrite behavior;
- exact target paths for repo and user installs;
- behavior differences between CLI, IDE, and cloud surfaces where applicable.

Required local setup:

- install OpenCode;
- install or enable Copilot CLI / `gh-copilot` / VS Code Copilot test surface;
- use temporary home/config directories to avoid polluting user-level agent configs;
- capture exact tool versions and command outputs.
