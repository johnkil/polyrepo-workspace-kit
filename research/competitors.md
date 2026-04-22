# Competitive Research

Date: 2026-04-19
Last updated: 2026-04-22
Scope: direct competitors, adjacent categories, and differentiation.

## Question

Who already solves parts of this problem, and where can `wkit` be meaningfully different?

## Short Answer

The competitive landscape is crowded if `wkit` is described as "manage many repos" or "generate agent files." It is less crowded if `wkit` is described as a local coordination layer for cross-repo changes and scenario validation.

The most important direct-adjacent competitor is [agents.ge](https://agents.ge/) on the agent-guidance side. The strongest adjacent competitors are Sourcegraph Batch Changes for cross-repo changes, Android `repo` / Zephyr `west` for multi-repo manifests, and Nx for AI-enhanced monorepo workflows.

## Competitive Map

| Category | Examples | What they solve | Risk to `wkit` | Gap for `wkit` |
| --- | --- | --- | --- | --- |
| Agent guidance sync | agents.ge | Durable `.agents` directory, rules, project memory, MCP config, sync to multiple agents | High | Mostly repo-agent readiness, not explicit polyrepo `change` / `scenario` validation |
| Agent instruction standards | AGENTS.md, Agent Skills, Claude Skills, Copilot instructions | Common files agents can read | High for adapters | Standards do not model workspace topology by themselves |
| Cross-repo change platforms | Sourcegraph Batch Changes | Large-scale code changes and PR tracking | Medium-high | Enterprise/platform oriented, not a small local workspace source of truth |
| Single-repo stacked PR workflows | GitHub Stacked PRs | Reviewable branch/PR chains inside one repository | Medium | Strong for PR layering, but not a polyrepo workspace, relation, or scenario evidence model |
| Multi-repo manifests | Android `repo`, Zephyr `west`, `vcstool` | Checkout/update many repos from manifest files | Medium | Usually checkout/version focused, not agent context or change narratives |
| Multi-repo command utilities | `gita`, `myrepos`-style tools | Run git/status commands across many repos | Medium | Useful utilities, but shallow semantics |
| Bulk polyrepo git managers | `polyrepopro/polyrepo` | Sync, pull, push, commit, status, switch, and run commands across configured repos | Medium | Operates repositories, but does not model cross-repo change intent, validation evidence, or agent guidance |
| Bulk PR tools | `multi-gitter`, `git-xargs` | Run commands across repos and open PRs | Medium | Great for mechanical changes, weaker at workspace meaning |
| Monorepo/build platforms | Nx, Turborepo, Bazel, Pants | Build/test/task graphs, caching, dependency execution | Medium | Strong if user can be monorepo-like; not the same as polyrepo coordination |
| Developer portals | Backstage | Ownership, service catalog, metadata, discovery | Medium | Heavy platform category, not local active-change workflow |

## Direct-Adjacent: agents.ge

[agents.ge](https://agents.ge/) is very close to the agent-guidance part of `wkit`.

It describes itself as an open-source CLI that creates a versioned `.agents` directory with rules, project knowledge, MCP config, and capture workflows across Claude Code, Cursor, Codex, GitHub Copilot, Gemini CLI, and other tools.

This validates the problem: developers want one durable source of agent context instead of hand-maintaining several tool-specific files.

Threat:

- If `wkit` focuses mostly on agent-ready files, agents.ge is an obvious alternative.
- agents.ge already uses language similar to "source of truth" and generated/synced agent formats.

Differentiation:

- `wkit` should focus on multiple repositories as a coordinated workspace, not just a repository-level `.agents` directory.
- `wkit` should make `change` and `scenario` first-class.
- `wkit` should describe repo relationships and validation evidence, not just rules and memory.

## Cross-Repo Change Platforms

### Sourcegraph Batch Changes

[Sourcegraph Batch Changes](https://sourcegraph.com/docs/batch-changes) helps automate and ship large-scale code changes across many repositories and code hosts. It can create pull requests, track progress, preview changes, and update them.

Threat:

- Sourcegraph is strong for enterprise-scale cross-repo changes.
- It already owns the phrase "large-scale code changes across many repositories."

Differentiation:

- `wkit` can be local, lightweight, and source-controlled.
- `wkit` can model why repos belong together before a batch change exists.
- `wkit` can generate agent guidance and scenario reports for a human/agent loop.
- `wkit` does not need to create PRs in v0.

### multi-gitter and git-xargs

[multi-gitter](https://github.com/lindell/multi-gitter) and [git-xargs](https://github.com/gruntwork-io/git-xargs) are useful for running commands or scripts across many repos and opening PRs.

Threat:

- They are simple and direct for mechanical migrations.

Differentiation:

- They do not aim to be a durable workspace model.
- They do not explain relationships, contexts, or scenario validation.
- They are execution tools more than coordination tools.

## Single-Repo Stacked PR Workflows

### GitHub Stacked PRs

[GitHub Stacked PRs](https://github.github.com/gh-stack/) is a native GitHub
workflow, currently in private preview, for arranging pull requests in an
ordered stack inside one repository. Each PR targets the branch below it, the UI
shows stack navigation, and the `gh stack` CLI helps create branches, push,
submit PRs, rebase, sync, and navigate the stack.

Threat:

- GitHub is validating the same review pain: large changes are easier to review
  when split into focused layers.
- The agent integration story is close to the skill/advice surface: GitHub
  documents a `github/gh-stack` skill for AI coding agents.
- If `wkit` drifts into branch chains, PR creation, rebasing, or merge
  orchestration, it will compete with a host-native workflow.

Differentiation:

- `gh-stack` is about vertical review layering inside one repository; `wkit` is
  about horizontal coordination across several local repositories.
- `gh-stack` owns branch bases, PR linking, stack rebases, and merge behavior;
  `wkit` should own repo relationships, contexts, cross-repo changes, scenario
  locks, and local validation evidence.
- A healthy combined workflow is: use stacked PRs inside a repository when a
  repo-local change needs review layers, and use `wkit` to declare which
  repositories belong to the coordinated change and what evidence proves the
  whole cross-repo change.

Boundary:

- Do not add `wkit` commands for stack creation, PR submission, cascading
  rebases, merge queues, or GitHub PR lifecycle tracking during v0.x.
- If GitHub state is useful later, keep it in derived handoff/report context,
  not canonical workspace state, and require pilot evidence before revisiting
  the boundary.

## Bulk Polyrepo Git Managers

### polyrepopro/polyrepo

[polyrepopro/polyrepo](https://github.com/polyrepopro/polyrepo) is a small Go CLI that uses a `.polyrepo.yaml` file to operate across a configured set of repositories. Its documented surface is centered on local and remote Git operations such as `sync`, `pull`, `push`, `commit`, `status`, `switch`, and `run`/watch.

Threat:

- It owns the direct `polyrepo` CLI name and the clean "manage polyrepo workspaces" category label.
- It validates that developers want a shared config for repeated multi-repo local operations.
- If `wkit` is described as "manage many repos", users may reasonably compare it to this category first.

Observed limitations for `wkit` positioning:

- The model is repository-operation centric: workspace, repository path, branch, tags, runners, and auth.
- It does not appear to model cross-repo relationships, change narratives, scenario locks, validation reports, or derived agent guidance.
- The CLI encourages broad mutation operations such as committing, pushing, pulling, switching branches, and syncing all configured repositories.
- The public CLI repository currently has a brittle local-path test, which is a useful reminder that multi-repo tools need disciplined local-environment boundaries.

Differentiation:

- `wkit` should not compete as a safer or broader `pull/push/status` wrapper.
- `wkit` should present Git status and checkout binding as supporting context for `change` and `scenario`, not as the product center.
- `wkit` should keep repository mutation narrow and explicit; reviewable scenario evidence is more important than bulk repository operation.
- Messaging should use phrases like "coordinate cross-repo changes" and "pin validation evidence" rather than "manage polyrepo workspaces" by itself.

## Multi-Repo Manifest Tools

### Android repo

Android's [`repo` manifest format](https://android.googlesource.com/tools/repo/%2B/HEAD/docs/manifest-format.md) models many Git projects using fields such as `name`, `path`, `remote`, `revision`, `groups`, and sync behavior.

What `wkit` should learn:

- explicit project identity matters;
- path and revision are separate concerns;
- groups are useful for scoping;
- pinning and update semantics matter.

What `wkit` should avoid in v0:

- becoming a checkout manager;
- overfitting to one ecosystem;
- owning all Git update behavior.

### Zephyr west

[Zephyr west manifests](https://docs.zephyrproject.org/latest/develop/west/manifest.html) define projects in a workspace and support update behavior, imports, project groups, and multiple repositories.

What `wkit` should learn:

- manifest imports are powerful but complicate mental models;
- project groups are useful for partial workspace activation;
- lock/update behavior is central to reproducibility.

What `wkit` should avoid in v0:

- becoming a domain-specific dependency manager;
- making repo checkout state the only product value.

### vcstool and gita

[vcstool](https://github.com/dirk-thomas/vcstool) and [gita](https://github.com/nosarthur/gita) demonstrate durable demand for operating across many Git repositories.

Threat:

- For simple "run git status everywhere" workflows, these are enough.

Differentiation:

- `wkit` should be more semantic: relations, contexts, changes, scenarios, and generated guidance.

## Monorepo And Build Platforms

### Nx

[Nx Enhance Your AI Coding Agent](https://nx.dev/docs/features/enhance-ai) is an important adjacent signal. Nx can configure AI agents, generate files such as `CLAUDE.md` and `AGENTS.md`, and connect local agents to Nx Cloud context via skills and MCP.

Threat:

- For Nx workspaces, this is a strong, integrated path.
- It makes "AI agent setup" a feature of the build/workspace platform.

Differentiation:

- `wkit` is for polyrepo teams that do not want or cannot adopt a monorepo workspace.
- `wkit` should remain build-system agnostic.
- `wkit` should use repo-local entrypoints rather than becoming the task graph.

### Turborepo, Bazel, Pants

These tools focus on task execution, dependency graphs, hermeticity, caching, and build/test orchestration.

Threat:

- Some teams will solve coordination by moving to a monorepo or monorepo-like build graph.

Differentiation:

- `wkit` should not compete on build execution.
- It should coordinate existing repos and call their local commands.

## Developer Portals

[Backstage Software Catalog](https://backstage.io/docs/features/software-catalog/) addresses ownership, metadata, discovery, and service cataloging.

Threat:

- Larger organizations may already have catalogs.

Differentiation:

- `wkit` can be local, file-based, and workflow-level.
- It should help with "this change across these repos today", not become a company-wide system of record.

## Competitive Strategy

Do not compete head-on with:

- Sourcegraph on enterprise PR orchestration;
- GitHub Stacked PRs on single-repo stacked PR orchestration;
- Nx/Bazel/Pants on build graph execution;
- Backstage on organizational catalogs;
- agents.ge on generic repository-agent readiness alone.

Compete where the overlap is weaker:

- local polyrepo workspace truth;
- cross-repo relationships;
- change narratives;
- pinned validation scenarios;
- agent-readable context derived from the same model.

## Differentiation Checklist

Every feature should answer at least one of these:

- Does it help describe relationships between repos?
- Does it help scope agent context safely?
- Does it help validate a cross-repo change?
- Does it produce evidence a reviewer can trust?
- Does it avoid replacing repo-local build/test truth?

If a feature only answers "it exports another agent file", it belongs in adapter support, not the core strategy.
