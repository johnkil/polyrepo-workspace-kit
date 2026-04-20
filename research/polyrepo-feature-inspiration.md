# Polyrepo Feature Inspiration

Date: 2026-04-20
Scope: feature-level research after reviewing `polyrepopro/polyrepo` as an inspiration source.

## Question

Which ideas from bulk polyrepo Git managers are worth adapting into `wkit`, and where do stronger adjacent tools suggest a better shape?

## Short Answer

The most useful ideas are not bulk mutation commands. They are **read-only orientation and safety diagnostics** around a local polyrepo workspace:

1. `wkit status` for repo checkout state before pinning or running scenarios.
2. `wkit doctor` for actionable workspace/binding/entrypoint diagnostics.
3. `wkit info` or `wkit overview` for cold-start understanding.

The risky ideas are mass `commit`, `push`, `pull`, `switch`, and a central arbitrary command runner. Those belong to bulk Git-operation tools, not to `wkit`'s core coordination model.

## 1. Workspace Status

### Inspiration

`polyrepopro/polyrepo` includes a multi-repo `status` command with clean/dirty and remote sync indicators.

### Better references

- Git's stable machine-readable status output includes branch headers such as `branch.head`, `branch.upstream`, and `branch.ab +<ahead> -<behind>` when `--branch` is used. It also documents that ahead/behind can be shown relative to the upstream branch.
- `gita` focuses heavily on side-by-side repo status, including local/remote branch relationship states and symbols for staged, unstaged, untracked, and stashed changes.
- Android `repo status` presents status grouped by project and uses file-level status codes.
- Zephyr `west status` is explicitly a workspace-level wrapper around `git status` in local project repositories.

### Recommendation for `wkit`

Add a read-only `wkit status` command before any bulk mutation command.

Suggested output:

- repo id;
- binding state: missing, inaccessible, non-git, ok;
- branch or detached HEAD;
- commit short hash;
- dirty/untracked counts;
- upstream if configured;
- ahead/behind counts from current local remote-tracking refs;
- scenario drift marker when a scenario is provided.

Important safety rule:

- Default status should not contact remotes.
- If freshness matters, require an explicit `--fetch` or `--refresh-remotes` flag and describe it as network access, not as sync.

Why this is better for `wkit`:

- It improves scenario readiness without becoming a Git operations manager.
- It can reuse the existing `internal/gitstate` direction and extend it carefully.

## 2. Scope Filters: Tags, Groups, And Contexts

### Inspiration

`polyrepopro/polyrepo` supports repository tags for filtering operations.

### Better references

- Zephyr west has project groups and group filters. Projects can become active or inactive based on group filtering, and commands such as `west update` and `west list` generally respect active project sets.
- `gita` supports repository groups for status and delegated commands.
- Android `repo` commands accept project lists by name or path, keeping scope explicit at command time.

### Recommendation for `wkit`

Do not add canonical `tags` yet.

Use the existing `context` entity as the main scoping primitive:

- `wkit context list`
- `wkit context show <context>`
- `wkit status --context <context>`
- `wkit scenario pin <scenario-id> --change <change-id>` continues to derive scope from the change context.

If pilots reveal repeated orthogonal grouping pain, add a carefully scoped metadata field later, but keep it non-execution-centric.

Why this is better for `wkit`:

- `context` already carries meaning for a change or validation boundary.
- Generic tags can become an attractive nuisance that turns the model into a command-dispatch system.

## 3. Missing Checkout Sync

### Inspiration

`polyrepopro/polyrepo sync` can ensure repositories exist locally and update them.

### Better references

- Android `repo sync` clones missing projects and updates existing ones, with explicit project lists and parallelism flags.
- Zephyr `west update` initializes missing local Git repositories and checks out manifest revisions, but it is intentionally a manifest/update tool.
- `vcstool import` clones repositories from a YAML file or URL, while `vcs validate` validates the repository file before use.

### Recommendation for `wkit`

Avoid a broad `sync` command in v0.x.

Better shape:

- `wkit doctor` reports missing bindings and missing checkout paths.
- `wkit bind set <repo-id> <path>` stays the explicit binding path.
- A future `wkit bind plan` or `wkit checkout plan` can show clone suggestions without mutating.
- If clone support is added later, make it opt-in and plan/apply based, with no pull/rebase behavior.

Why this is better for `wkit`:

- Binding paths are machine-local truth.
- Scenario evidence should not depend on hidden checkout mutation.

## 4. Workspace Info / Overview

### Inspiration

`polyrepopro/polyrepo info` prints workspace and repository information.

### Better references

- `west list` prints project information from the manifest and supports formatting.
- Homebrew has `brew info` for concise object-level information and `brew config` for debugging context.
- `gita ll` is useful because it compresses many repository states into a scannable table.

### Recommendation for `wkit`

Add `wkit info` or `wkit overview` early.

Suggested sections:

- workspace id and root;
- repos count by kind;
- relations count by kind;
- contexts and included repos;
- changes and latest scenario locks;
- binding coverage;
- guidance outputs available;
- next likely commands.

Why this is better for `wkit`:

- It supports cold-start onboarding without adding new canonical entities.
- It teaches users where truth lives: manifests, bindings, scenarios, and derived outputs.

## 5. Runner / Watch UX

### Inspiration

`polyrepopro/polyrepo` has runner/watch configuration and can run commands across repositories.

### Better references

- `just` is intentionally a project-specific command runner, not a build system.
- Taskfile supports watch mode, but requires `sources` so it knows what to watch and documents caveats around long-running server processes.
- `watchexec` is a dedicated file-watching command runner with ignore handling, event coalescing, and process-group behavior.
- Android `repo forall` and Zephyr `west forall` both support arbitrary commands across projects, but this is a broad power-user surface.

### Recommendation for `wkit`

Do not build a central runner/watch system.

Better shape:

- Keep repo-local entrypoints authoritative.
- Let `wkit scenario run` execute only declared entrypoints and write evidence.
- Generate guidance that points agents to repo-local `just`, `make`, `task`, package scripts, or scripts where they already exist.
- If watch mode is requested later, prefer documenting integration with `watchexec` or repo-local tooling rather than owning the watcher.

Why this is better for `wkit`:

- It keeps scenario execution honest.
- It avoids becoming a universal command abstraction layer.

## 6. Branch Consistency And Scenario Drift

### Inspiration

Bulk polyrepo tools often support branch switching or branch status across repositories.

### Better references

- Zephyr `west compare` compares workspace state against the manifest.
- Zephyr `west manifest --freeze` produces a manifest where every project revision is a SHA.
- `vcstool export --exact` captures exact repository revisions for reproducibility.
- Git's `rev-list --left-right --count` can compute counts on each side of a symmetric difference.

### Recommendation for `wkit`

Double down on read-only drift detection.

Better shape:

- `wkit scenario run` already blocks on pinned commit drift.
- Add `wkit scenario status <scenario-id>` or allow `wkit status --scenario <scenario-id>`.
- Show branch changes as context, but treat commit drift as the hard signal.
- Do not add `switch all repos` or branch mutation.

Why this is better for `wkit`:

- Scenario locks are already the differentiated feature.
- Branch names are helpful labels, but commits are the reviewable evidence.

## 7. Remote Config Bootstrap

### Inspiration

`polyrepopro/polyrepo init -u` can download a `.polyrepo.yaml` from a URL.

### Better references

- Android `repo init -u` initializes a client from a manifest repository URL, with manifest file and branch options.
- Zephyr `west init -m` clones a manifest repository, supports choosing a manifest revision and file, and records the manifest repository path.
- `vcstool import` supports importing repository YAML from a file or URL and has a separate `vcs validate` command.

### Recommendation for `wkit`

Do not add raw URL init as a simple shortcut.

If this becomes important, use a safer two-step flow:

- `wkit init-from <url> --plan`
- strict YAML decode and schema validation;
- display source URL, resolved revision/checksum when possible, target paths, and overwrite plan;
- `--apply` only after plan review;
- never overwrite an existing workspace silently.

Why this is better for `wkit`:

- Bootstrap is safety-sensitive.
- A remote config can define file paths and derived output targets, so provenance and preview matter.

## 8. Parallel Execution

### Inspiration

`polyrepopro/polyrepo` uses parallel goroutines for repo operations.

### Better references

- Android `repo sync -j` parallelizes sync across threads.
- `vcstool` parallelizes work across repositories by default based on CPU cores, but recommends `--workers 1` when commands need separate stdin, such as interactive credentials.
- `myrepos` supports `mr -j5 update` for concurrent jobs.

### Recommendation for `wkit`

Keep v0.x scenario runs deterministic and simple.

Later, add parallelism only if reports stay clear:

- explicit `--jobs N`;
- deterministic report ordering by repo id/check id;
- per-check stdout/stderr capture;
- clear timeout and cancellation semantics;
- default sequential mode until pilots show runtime pain.

Why this is better for `wkit`:

- Scenario reports are review evidence.
- A slightly slower report that is easy to read is better than a fast report with interleaved ambiguity.

## Negative Controls: What Not To Borrow

Avoid these as `wkit` core features:

- mass `commit`;
- mass `push`;
- mass `pull`;
- mass branch `switch`;
- arbitrary central `run`;
- global watch orchestration;
- implicit remote config import;
- silent checkout mutation.

These are valid features for a different product category. For `wkit`, they blur the product into a polyrepo Git operator instead of a coordination and validation layer.

## Prioritized Candidate Work

1. `wkit info` / `wkit overview`
   - Highest onboarding value.
   - Low safety risk.

2. `wkit status`
   - High scenario-readiness value.
   - Should be read-only by default.

3. `wkit doctor`
   - High support value.
   - Should explain how to fix missing bindings, broken paths, stale scenario locks, and invalid entrypoints.

4. `wkit context list/show`
   - Makes existing scoping visible.
   - Avoids adding generic tags too early.

5. `wkit scenario status <scenario-id>`
   - Strengthens the existing scenario wedge.
   - Keeps branch/drift work centered on evidence.

## Sources

- [`polyrepopro/polyrepo` README](https://github.com/polyrepopro/polyrepo/blob/main/docs/readme.md)
- [Git `status` documentation](https://git-scm.com/docs/git-status)
- [Git `fetch` documentation](https://git-scm.com/docs/git-fetch)
- [Git `rev-list` documentation](https://git-scm.com/docs/git-rev-list)
- [Android Repo command reference](https://source.android.com/docs/setup/reference/repo)
- [Zephyr west built-in commands](https://docs.zephyrproject.org/latest/develop/west/built-in.html)
- [Zephyr west manifests](https://docs.zephyrproject.org/latest/develop/west/manifest.html)
- [`vcstool` README](https://github.com/dirk-thomas/vcstool)
- [`gita` README](https://github.com/nosarthur/gita)
- [myrepos](https://myrepos.branchable.com/)
- [`just` README](https://github.com/casey/just)
- [Taskfile usage guide](https://taskfile.dev/usage/)
- [`watchexec` README](https://github.com/watchexec/watchexec)
- [Homebrew manpage: `doctor`](https://docs.brew.sh/Manpage#doctor-options)
- [mise `doctor` reference](https://mise.jdx.dev/cli/doctor.html)
