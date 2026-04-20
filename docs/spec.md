# Technical Specification
## Polyrepo Workspace Kit

Version: 1.7
Status: Public draft with VS Code workspace pilot

## 1. Repository structure

```text
workspace/
  coordination/
    workspace.yaml
    contexts.yaml
    changes/
    scenarios/
      <scenario-id>/
        manifest.lock.yaml
    rules/
  guidance/
    rules/
      always-on.md
    skills/
      <skill-name>/
        SKILL.md
  repos/
    <repo-id>/
      repo.yaml
  local/
    bindings.yaml
    reports/
      <scenario-id>/
        <run-id>.yaml
    vscode/
      workspace.code-workspace
  runtime/
  config/
  bin/
```

## 2. State classes

### 2.1 Canonical shared state

- `coordination/workspace.yaml`
- `coordination/contexts.yaml`
- `coordination/changes/*`
- `coordination/scenarios/*/manifest.lock.yaml`
- `coordination/rules/*`
- `repos/*/repo.yaml`
- `guidance/rules/*`
- `guidance/skills/*`

### 2.2 Canonical machine-local state

- `local/bindings.yaml`

### 2.3 Derived state

- scenario execution reports under `local/reports/*`
- VS Code multi-root workspace exports under `local/vscode/*`
- portable outputs such as `AGENTS.md` and `.agents/skills/*`
- tool-specific adapter outputs such as `CLAUDE.md`, `.claude/*`, `.github/*`, `.opencode/*`, or other adapter targets in the initial v0.x target surface

Derived state is disposable and must never become the source of truth.

## 3. Canonical entities

The following entities are canonical in v0.x:

- `workspace`
- `repo`
- `relation`
- `rule`
- `context`
- `change`
- `scenario`
- `binding`
- `entrypoint`

No additional canonical entity should be introduced without passing the YAGNI gate.

### 3.1 File-backed id safety

Any id used to construct a workspace-relative file path must be path-safe.

This applies at least to:

- repo ids;
- change ids;
- scenario ids;
- coordination rule ids.

Path-safe ids must:

- match `[A-Za-z0-9][A-Za-z0-9_-]*(\.[A-Za-z0-9][A-Za-z0-9_-]*)*`;
- avoid empty dot segments such as `..`;
- avoid absolute paths;
- avoid `/` and `\` path separators;
- avoid leading or trailing whitespace.

## 4. Canonical files

### 4.1 `coordination/workspace.yaml`

```yaml
version: 1

workspace:
  id: demo-workspace
  model: thin-coordination-layer

repos:
  - app-web
  - shared-schema

relations:
  - from: app-web
    to: shared-schema
    kind: contract

rules:
  - contract-rollout-order
```

`workspace.yaml rules:` references coordination rule ids. It does not embed full rule bodies and does not point to guidance rules.

### 4.1a `coordination/rules/<rule-id>.yaml`

```yaml
version: 1

rule:
  id: contract-rollout-order
  kind: rollout-order
  applies_to:
    relation_kind: contract
  policy:
    order: provider-before-consumer
```

### 4.1b Coordination rules vs guidance rules

- `coordination/rules/*` are canonical coordination constraints, such as rollout order or validation policy.
- `guidance/rules/*` are canonical portable agent guidance, such as short always-on instructions.
- `workspace.yaml rules:` references coordination rule ids, not guidance rule files.

Allowed coordination rule kinds in v0.x:

- `rollout-order`

For `rollout-order` rules in v0.x:

- `applies_to` may reference one or more of: `relation_kind`, `from_repo`, `to_repo`, or `context`;
- `policy.order` must be one of: `provider-before-consumer` or `consumer-after-provider`.

In v0.x, coordination rules are advisory validation constraints. They may produce warnings or blocked scenario checks, but they do not mutate repository state or execute commands by themselves.

### 4.2 `repos/<repo-id>/repo.yaml`

```yaml
version: 1

repo:
  id: app-web
  kind: app

read_first:
  - docs/architecture.md

entrypoints:
  test:
    run: bin/test
    cwd: .
    timeout_seconds: 600
    env_profile: default
```

Shorthand string form is allowed:

```yaml
entrypoints:
  test: bin/test
```

In v0.x, `env_profile` is an opaque descriptive label. It does not imply automatic environment loading, secret loading, shell activation, or toolchain management. Scenario runners may display it in reports, but repo-local commands remain responsible for their own environment setup.

### 4.3 `coordination/contexts.yaml`

```yaml
version: 1

contexts:
  schema-rollout:
    repos:
      - shared-schema
      - app-web
```

### 4.4 `coordination/changes/CHG-YYYY-MM-DD-NNN.yaml`

```yaml
version: 1

change:
  id: CHG-2026-04-19-001
  title: payload v3 rollout
  kind: contract
  context: schema-rollout
  repos:
    - shared-schema
    - app-web
```

### 4.5 `local/bindings.yaml`

```yaml
version: 1

bindings:
  app-web:
    path: /absolute/path/to/app-web
  shared-schema:
    path: /absolute/path/to/shared-schema
```

Bindings are machine-local canonical state. They map logical repo ids to local checkouts.

### 4.6 `coordination/scenarios/<scenario-id>/manifest.lock.yaml`

A scenario lock is a **reviewable local validation snapshot**, not a full environment reproduction artifact.

```yaml
version: 1

scenario:
  id: schema-rollout
  change: CHG-2026-04-19-001
  context: schema-rollout
  generated_at: 2026-04-19T12:00:00Z
  generated_by:
    tool: wkit
    version: 0.y.z
  semantics: reviewable-local-validation-snapshot
  notes:
    - v0.x scenarios pin revisions and local checks but do not guarantee full environment replay.

tool_versions:
  wkit: 0.y.z
  git: 2.49.0
  extra: {}

repos:
  - repo: shared-schema
    revision:
      commit: 2e5cc2d6f6a8f1f6d7f8a1b2c3d4e5f60718293a
      short: 2e5cc2d6
      branch: main
    worktree:
      clean: true
      dirty_files: 0
      untracked_files: 0
      dirty_paths: []
      untracked_paths: []
    dependency_hints:
      lockfiles: []

  - repo: app-web
    revision:
      commit: b5f92f2f093ab0a1b2c3d4e5f60718293acdd991
      short: b5f92f2f
      branch: feature/payload-v3
    worktree:
      clean: false
      dirty_files: 2
      untracked_files: 1
      dirty_paths: []
      untracked_paths: []
    dependency_hints:
      lockfiles:
        - path: package-lock.json
          sha256: null

checks:
  - id: shared-schema:test
    repo: shared-schema
    cwd: .
    run: bin/test
    timeout_seconds: 600
    env_profile: default
    env_requirements: []
    expected_artifacts: []
    requires_clean_worktree: true
    status: planned

  - id: app-web:test
    repo: app-web
    cwd: .
    run: bin/test
    timeout_seconds: 900
    env_profile: default
    env_requirements: []
    expected_artifacts: []
    requires_clean_worktree: false
    status: planned
```

#### Scenario lock notes

In v0.x:

- `dirty_paths`, `untracked_paths`, and `dependency_hints.lockfiles` are **optional but recommended** when they are cheap to collect;
- `tool_versions.extra` may record relevant toolchain versions when known;
- `env_requirements` is descriptive, not a full environment capture system;
- `env_profile` is descriptive metadata only and does not imply automatic environment loading.

### 4.7 `local/reports/<scenario-id>/<run-id>.yaml`

Reports are derived artifacts. They are reviewable evidence, not canonical inputs.

```yaml
version: 1

report:
  scenario: schema-rollout
  generated_at: 2026-04-19T12:10:00Z
  report_kind: local-validation-run

results:
  - check: shared-schema:test
    status: passed
    duration_seconds: 41
    env_profile: default
    stdout_path: logs/20260419T121000Z/shared-schema-test.stdout.txt
    stderr_path: logs/20260419T121000Z/shared-schema-test.stderr.txt
    artifacts: []

  - check: app-web:test
    status: blocked
    reason: pinned ref drift: current HEAD does not match scenario lock
    env_profile: default
    stdout_path: null
    stderr_path: null
    artifacts: []
```

`run-id` should start with a UTC timestamp such as `20260419T121000Z`. If a report for that timestamp already exists, implementations must not overwrite it and should append a stable numeric suffix such as `.001`, `.002`, and so on.

Report file paths are relative to the report directory unless an implementation explicitly documents otherwise.

Implementations may also write a paired text summary report such as:

```text
local/reports/<scenario-id>/<run-id>.txt
```

The text report is a derived review aid. The YAML report remains the structured report artifact.

### 4.8 `local/vscode/workspace.code-workspace`

The VS Code workspace file is a local derived artifact generated from:

- `coordination/workspace.yaml`;
- `repos/*/repo.yaml`;
- `local/bindings.yaml`;
- `coordination/scenarios/*/manifest.lock.yaml` where present.

It is intended to make a `wkit` workspace usable as a VS Code multi-root
workspace without turning VS Code metadata into canonical state.

The generated file should include:

- one folder for the `wkit` workspace root;
- one folder for each bound repo checkout, named by repo id;
- workspace tasks for `wkit overview`, `wkit validate`, `wkit doctor`, and
  `wkit status`;
- workspace tasks for pinned scenario status/run commands where scenario locks
  exist;
- workspace tasks for repo entrypoints from `repos/<repo-id>/repo.yaml`;
- conservative workspace settings that improve multi-root readability without
  changing repository behavior.

The VS Code export must not write `.vscode/*` files into bound repositories by
default. It must remain disposable and regenerable from canonical `wkit` state.

## 5. Relation semantics

Relations are directional.

- `from` = consumer / depender / initiator side
- `to` = provider / dependency / upstream side

Allowed `kind` values in v0.x:

- `runtime`
- `build`
- `contract`
- `release`
- `docs`

Relation kinds mean:

- `runtime` — `from` consumes runtime behavior or published package output from `to`;
- `build` — `from` depends on build-time outputs, generation, or templates from `to`;
- `contract` — `from` consumes or implements a schema, API, protocol, or compatibility surface defined by `to`;
- `release` — rollout order or release readiness of `from` depends on `to`;
- `docs` — docs, examples, or reference material in `from` track changes originating in `to`.

Relations are used for:

- bounded context expansion;
- rollout reasoning;
- scenario scoping;
- validation warnings.

In v0.x:

- relation expansion should stay bounded and explicit;
- no automatic full-graph traversal is implied;
- rollout rules may refer to relation kind, but relations alone do not encode every rollout policy.

## 6. Portable guidance model

### 6.1 Canonical source

```text
guidance/
  rules/
    always-on.md
  skills/
    <skill-name>/
      SKILL.md
```

### 6.2 Portable outputs

Portable outputs are limited to:

- `AGENTS.md`
- `.agents/skills/*`

Portable outputs are derived artifacts, not source-of-truth files.

### 6.3 Adapter outputs

Adapter outputs are tool-specific, non-canonical derived artifacts.

#### Initial v0.x target surface

The following target paths are the intended v0.x adapter surface. They should be treated as `docs-backed` until an empirical compatibility pass records tool version, probe date, target path, and observed behavior.

- **portable**
  - verification status: docs-backed
  - repo scope: `AGENTS.md`, `.agents/skills/*`
  - user scope: `.agents/skills/*`
- **codex**
  - verification status: docs-backed
  - repo scope: same as portable
- **opencode**
  - verification status: docs-backed
  - repo scope: same as portable
- **copilot**
  - verification status: docs-backed
  - repo scope: `.github/copilot-instructions.md`
- **claude**
  - verification status: docs-backed
  - repo scope: `CLAUDE.md`, `.claude/skills/*`

#### Candidate / unverified targets

The following are **not** part of the v0.x compatibility guarantee unless separately validated and documented:

- Codex user-scope instruction files;
- OpenCode-specific config or command directories;
- OpenCode user-scope installs beyond portable `.agents/skills/*`;
- Copilot user-scope installs;
- Claude user-scope installs;
- `.github/skills/*` or other Copilot surfaces beyond the initial target surface above;
- any `.codex/*` or `.opencode/*` path not explicitly marked in the initial target surface;
- any target supported only by docs-backed inference and not by current compatibility notes.

In v0.x, the only user-scope target in the initial target surface is portable `.agents/skills/*`.
User-scope targets for Codex, OpenCode, Copilot, and Claude remain candidate / unverified unless separately validated.

## 7. Resolution policy

### 7.1 Lookup order

1. Resolve target repo from explicit input, current working directory, or active `change`.
2. Read repo-local descriptors and curated `read_first` docs.
3. Expand cross-repo only by declared `relation`, active `change`, or explicit task signal.
4. Expand into a named `context`, not the whole workspace.

### 7.2 Authority order

1. repo-local executable truth;
2. repo-local manifests and docs;
3. workspace coordination files;
4. shared guidance;
5. local notes only by explicit opt-in.

## 8. Validation contract

The validator must check:

- workspace file presence and shape;
- workspace rule references point to existing `coordination/rules/<rule-id>.yaml` files;
- coordination rule file `rule.id` matches the referenced id and file stem;
- coordination rule `kind` is in the allowed v0.x vocabulary;
- coordination rule `applies_to` references valid relation kinds, repo ids, or contexts where used;
- coordination rule `policy` shape is valid for its `kind`;
- repo manifest presence and id consistency;
- binding path existence and repo id coverage;
- relation endpoint validity and allowed kinds;
- context repo references;
- change context and repo references;
- scenario refs and checks;
- report schema where report files exist.

It may additionally warn on:

- missing `test` or other expected entrypoints;
- stale bindings;
- generated adapter outputs that appear manually edited;
- conflicting generated guidance in the same repo.

## 9. CLI contract

### 9.1 Core commands

- `wkit init <path>`
- `wkit repo register <repo-id> --kind <kind>`
- `wkit bind set <repo-id> <path>`
- `wkit context list`
- `wkit context show <context-id>`
- `wkit info`
- `wkit overview`
- `wkit status [--context <context-id>]`
- `wkit doctor`
- `wkit validate`
- `wkit version`
- `wkit --version`
- `wkit change new <context> --title <title>`
- `wkit change show <change-id>`
- `wkit scenario pin <scenario-id> --change <change-id>`
- `wkit scenario show <scenario-id>`
- `wkit scenario status <scenario-id>`
- `wkit scenario run <scenario-id>`
- `wkit vscode plan`
- `wkit vscode diff`
- `wkit vscode apply`
- `wkit vscode open`

### 9.2 Orientation and diagnostics commands

Orientation and diagnostics commands are read-only local inspection. They must not
clone, fetch, pull, push, switch branches, commit, execute scenario checks, or
mutate repository checkouts.

- `wkit context list` prints known context ids sorted by id with repo counts.
- `wkit context show <context-id>` prints repo ids in manifest order. Unknown
  contexts fail with exit code `1`.
- `wkit info` and its alias `wkit overview` print workspace id/root, repo counts
  by kind, relation counts by kind, context summaries, change/scenario
  counts/latest ids, binding coverage, guidance counts, and next likely local
  commands.
- `wkit status [--context <context-id>]` prints local checkout state for all
  workspace repos or the repos in the named context. It reports repo id, binding
  state, branch or detached state, short commit, dirty count, untracked count,
  upstream, and ahead/behind counts when a local upstream ref exists. It must not
  run `git fetch`; missing upstream is reported as `upstream=none`.
- `wkit scenario status <scenario-id>` compares current local checkouts with the
  pinned scenario lock without running checks. Per pinned repo, it reports pinned
  commit, current commit, branch label, and `ok`, `drift`, `missing`, or
  `blocked`. Drift, missing commit data, missing bindings, inaccessible paths,
  and non-git checkouts fail with exit code `4`.
- `wkit doctor` combines manifest validation with actionable local diagnostics
  for missing bindings, inaccessible checkouts, non-git checkouts, invalid or
  missing entrypoint `cwd` paths, and stale scenario locks. It exits `0` when
  there are no errors and `2` when there are errors; warnings alone do not fail.
- `wkit version` prints local build metadata: version, commit, build date, dirty
  state, and builder. `wkit --version` prints a compact single-line variant.
  Neither form inspects a workspace, and both exit `0`.

### 9.3 Install commands

- `wkit install plan <tool> [repo-id]`
- `wkit install diff <tool> [repo-id]`
- `wkit install show-targets <tool> [repo-id]`
- `wkit install apply <tool> [repo-id]`

### 9.4 Install flags

- `--scope repo|user`
- `--user-root <path>`
- `--dry-run`
- `--yes`
- `--force`
- `--backup`
- tool-specific optional flags where needed

## 10. Scenario behavior

### 10.1 Scenario intent

For v0.x, a scenario is a **reviewable local validation snapshot**.

It is intended to support:

- pinned local review;
- bounded drift detection;
- normalized execution of repo-local entrypoints;
- handoff and evidence.

It is **not** intended to guarantee:

- complete dependency replay;
- machine-independent environment reproduction;
- CI-level orchestration.

### 10.2 `scenario pin`

The implementation must:

- resolve the referenced change;
- resolve bound local repo checkouts;
- read the current git commit for each repo in the change;
- record full commit and short ref;
- record current branch where available;
- record worktree cleanliness and counts for dirty/untracked files;
- record path lists for dirty/untracked files when cheap and available;
- derive checks from repo entrypoints, preferring `test`;
- normalize each check to include `cwd`, `run`, `timeout_seconds`, and `env_profile`;
- write the lock manifest.

Scenario-specific check lists are deferred for v0.x. In v0.x, checks are derived from repo entrypoints only.

### 10.3 `scenario run`

The implementation must:

- load the pinned manifest;
- resolve local bindings for every referenced repo;
- compare current refs with pinned refs;
- compare worktree cleanliness when `requires_clean_worktree` is set;
- run the declared commands in the declared `cwd`;
- apply timeout behavior where declared;
- include `env_profile` in reports when present, without loading environment state automatically;
- emit a text-first report suitable for review;
- record per-check status;
- optionally record `stdout_path`, `stderr_path`, and artifact paths.

In v0.x, command execution is intentionally narrow: commands are split into an executable plus arguments and run without implicit shell expansion. Quoted arguments are not parsed as shell syntax; repo-local scripts should be used when a workflow needs grouped arguments, shell features, environment setup, or complex command composition.

### 10.4 Check statuses

Allowed check statuses in v0.x:

- `planned`
- `passed`
- `failed`
- `skipped`
- `blocked`

`blocked` means the check could not safely start because a prerequisite was not met, for example missing binding, missing command, or disallowed worktree state.

## 11. VS Code workspace behavior

### 11.1 Intent

The VS Code export is an IDE orientation surface, not a new adapter truth model.
It should make bound polyrepo checkouts easier to open, inspect, search, and run
through declared entrypoints in VS Code.

### 11.2 Generated target

The initial v0.x target is:

```text
local/vscode/workspace.code-workspace
```

The fixed target path avoids using `workspace.id` as a filesystem component.

### 11.3 Commands

- `wkit vscode plan`
- `wkit vscode diff`
- `wkit vscode apply`
- `wkit vscode open`

`plan`, `diff`, and `apply` follow the same preview-before-write posture as
install commands. `apply` requires `--yes` before writing and supports
`--dry-run`, `--force`, and `--backup`.

`open` runs `code <workspace-file>` for the generated file. If the workspace
file is missing or stale, `open` must not write unless the user passes `--yes`.

### 11.4 Task generation

Generated VS Code tasks should use `process` tasks where possible, with command
and arguments separated. Repo entrypoint tasks must preserve repo-local
executable truth from `repo.yaml`; they must not invent centralized commands.

Entrypoint commands with quoted arguments remain unsupported in v0.x. Users
should move shell-sensitive workflows into repo-local wrapper scripts.

Repo entrypoint task `cwd` values must be relative to the bound repo and must
not escape the repo checkout. Generated tasks should use scoped
`${workspaceFolder:<repo-id>}` variables for bound repo folders.

### 11.5 Safety

The VS Code export must:

- require bindings for every declared repo before rendering a complete
  workspace file;
- keep generated output inside the `wkit` workspace root both lexically and
  after resolving existing parent directories;
- block symlinked target files or symlinked parent paths that escape the
  workspace boundary;
- refuse to overwrite changed files unless `--force` or `--backup` is explicit;
- write backups with the same `<original-path>.bak.<UTC timestamp>` convention
  used by install targets.

## 12. Adapter contract

Adapters are installation strategies for real tool discovery scopes.

### 12.1 Supported adapters in v0.x

- `portable`
- `codex`
- `opencode`
- `copilot`
- `claude`

### 12.2 Scope model

- **Repo scope** is the default and primary installation mode.
- **User scope** is explicit and may differ by tool.
- User-scope installs should prefer **skills-first** behavior where global instruction paths are inconsistent across tools.

### 12.3 Compatibility model

Adapter behavior is a versioned compatibility assumption, not timeless truth.

The project should maintain:

- a docs-backed compatibility table;
- an empirical compatibility matrix;
- adapter tests or probes where feasible.

Public docs must distinguish:

- the initial v0.x target surface;
- verification status (`docs-backed`, `empirically verified`, or `candidate / unverified`);
- tool/version-specific caveats.

## 13. Install safety contract

### 13.1 Plan target record

Each plan target must include:

- `tool`
- `scope`
- `path`
- `kind` (`instructions`, `skill`, `command`, or other adapter-specific kind)
- `source`
- `status`
- `ownership` (`wkit-owned`, `unmarked`, `foreign`, or `unknown`)
- optional `backup_path`
- optional `notes`

### 13.2 Plan statuses

Allowed plan statuses in v0.x:

- `new` — target path does not exist.
- `unchanged` — target exists and content already matches.
- `blocked` — target exists and would be overwritten, overwrite is not allowed under current flags, or the target/source path is unsafe.
- `overwrite` — target exists and will be replaced because overwrite is explicitly allowed.
- `backup+overwrite` — target exists, a backup will be written, and then the target will be replaced.

### 13.3 Ownership markers

Where file format permits, adapter outputs should include a short marker indicating:

- generated by `wkit`;
- canonical source location;
- non-canonical / do not edit by hand notice where appropriate.

If file format does not naturally support comments, ownership may remain `unknown` and overwrite rules must stay conservative.

### 13.4 Unmarked existing files

If a target exists and content differs:

- without `--force` or `--backup`, status is `blocked`;
- with `--force`, status is `overwrite`;
- with `--backup`, status is `backup+overwrite`.

The implementation should not assume that an unmarked file is safe to replace.

### 13.5 Backup naming

Backups should use:

`<original-path>.bak.<UTC timestamp>`

If that backup path already exists, implementations must not overwrite it. They should append a stable numeric suffix such as `.001`, `.002`, and so on until a free backup path is found.

Backup creation should stage content before finalizing the backup where possible. If backup creation fails after creating the final backup path, implementations should clean up that newly created path so a valid-looking partial backup artifact is not left behind.

### 13.6 Install path containment

Installer planning, diffing, and applying are safety-sensitive because both source
guidance and existing target files may be supplied by an untrusted workspace or
checkout.

Implementations must not follow symlinks in a way that reads or writes outside
the intended boundary:

- repo-scope targets must remain inside the bound repo checkout both lexically
  and after resolving existing parent directories;
- user-scope targets must remain inside the selected user root both lexically
  and after resolving existing parent directories;
- existing target files that are symlinks must be blocked before diff, backup,
  or overwrite logic reads them;
- guidance source files used for rules and skills must be regular files inside
  their canonical `guidance/*` roots; symlinked source files must be rejected or
  blocked before their contents are read.

Unsafe source or target paths should be reported as blocked plan targets with a
short explanatory note where practical.

### 13.7 Exit codes

Suggested v0.x exit codes:

- `0` — success
- `2` — validation, doctor, or argument error
- `3` — blocked install or VS Code workspace targets
- `4` — scenario drift, missing binding, or blocked `scenario status` / `scenario run` prerequisite
- `5` — command failure during scenario run

If one scenario run contains both blocked/drift checks and command failures, command failure takes precedence for the process exit code.

## 14. Proof-oriented invariants

The following are invariants for the proof stage:

- the core model must remain useful even if no adapter is installed;
- adapter outputs must remain derived and disposable;
- portable guidance must stay minimal and high-signal;
- no new canonical entity should be introduced without repeated pilot evidence.

## 14. Out of core for v0.x

The following are out of core for v0.x:

- universal custom-agent schema;
- universal command model;
- plugin registry or marketplace;
- pack-first architecture;
- graph auto-discovery as canonical truth;
- ownership catalog;
- hosted runtime or policy layer;
- generic project memory as canonical workspace state.
