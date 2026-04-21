# Implementation Plan
## Polyrepo Workspace Kit

Status: Working engineering plan
Last updated: 2026-04-21

## Purpose

This document tracks implementation order.

It is different from `docs/plan.md`:

- `docs/plan.md` explains how to prove product value.
- `docs/implementation-plan.md` explains how to build the CLI in a disciplined order.

## Current Shipped Surface

The first Go implementation slice exists and is tested.

Implemented commands:

- `wkit init <path> [--repo <id=path> ...]`
- `wkit demo [minimal|failure]`
- `wkit repo register <repo-id> --kind <kind>`
- `wkit bind set <repo-id> <path>`
- `wkit context list`
- `wkit context show <context-id>`
- `wkit relations suggest [--context <context-id>]`
- `wkit info`
- `wkit overview`
- `wkit status [--context <context-id>]`
- `wkit doctor`
- `wkit version`
- `wkit --version`
- `wkit telemetry enable`
- `wkit telemetry disable`
- `wkit telemetry status`
- `wkit telemetry export`
- `wkit change new <context> --title <title>`
- `wkit change show <change-id>`
- `wkit handoff <change-id>`
- `wkit scenario pin <scenario-id> --change <change-id>`
- `wkit scenario show <scenario-id>`
- `wkit scenario status <scenario-id>`
- `wkit scenario run <scenario-id>`
- `wkit vscode plan`
- `wkit vscode diff`
- `wkit vscode apply`
- `wkit vscode open`
- `wkit install show-targets <tool> [repo-id]`
- `wkit install plan <tool> [repo-id]`
- `wkit install diff <tool> [repo-id]`
- `wkit install apply <tool> [repo-id]`
- `wkit validate`

Implemented install tools:

- `portable`
- `codex`
- `opencode`
- `copilot`
- `claude`

Implemented packages:

- `cmd/wkit`
- `internal/buildinfo`
- `internal/cli`
- `internal/demo`
- `internal/fsutil`
- `internal/gitstate`
- `internal/handoff`
- `internal/install`
- `internal/manifest`
- `internal/model`
- `internal/orient`
- `internal/relations`
- `internal/scaffold`
- `internal/scenario`
- `internal/telemetry`
- `internal/validate`
- `internal/vscode`
- `internal/workspace`

## Implementation Principles

- Current spec wins over the reference implementation.
- Core coordination comes before adapters.
- Repo scope comes before user scope.
- Portable outputs come before tool-specific outputs.
- Generated adapter outputs are never canonical truth.
- Tool-specific user-scope targets stay deferred until compatibility evidence exists.
- Do not add new canonical entities without passing the YAGNI gate in the spec.
- Every milestone needs focused tests before the next milestone expands scope.

## Milestone 1: Core CLI Foundation

Status: done

Scope:

- workspace initialization;
- repo registration;
- local bindings;
- change creation and display;
- scenario pin/show/run;
- workspace validation;
- Go module and CLI entrypoint;
- ADR for CLI tech stack.

Acceptance:

- `go test ./...` passes;
- `go test -race ./...` passes;
- `go run ./cmd/wkit --help` works;
- README describes shipped vs planned behavior truthfully;
- scenario reports are written under `local/reports/*`.

## Milestone 2: Portable Install Layer

Status: done

Scope:

- `wkit install show-targets portable [repo-id]`
- `wkit install plan portable [repo-id]`
- `wkit install diff portable [repo-id]`
- `wkit install apply portable [repo-id]`

Repo-scope outputs:

- `AGENTS.md`
- `.agents/skills/*`

User-scope output:

- `.agents/skills/*`

Acceptance:

- plan reports target path, kind, source, status, and backup path when relevant;
- diff shows textual changes for instruction files and skill files where practical;
- apply refuses to overwrite changed files without `--force` or `--backup`;
- `--backup` writes `<original-path>.bak.<UTC timestamp>`, with a numeric suffix when that path already exists;
- generated files include ownership markers where the file format permits;
- tests cover new, unchanged, blocked, force overwrite, and backup overwrite.

## Milestone 3: Repo-Scope Tool Adapters

Status: done

Scope:

- `codex` repo scope, same portable baseline;
- `opencode` repo scope, same portable baseline;
- `copilot` repo scope, `.github/copilot-instructions.md`;
- `claude` repo scope, `CLAUDE.md` and `.claude/skills/*`.

Acceptance:

- adapter outputs remain derived;
- target paths match `docs/spec.md`;
- user-scope tool-specific installs remain out of scope unless compatibility evidence is added;
- tests cover plan/diff/apply behavior for each adapter.

## Milestone 4: Scenario Hardening

Status: done

Scope:

- polish drift and blocked-run reporting;
- improve report readability;
- review command execution policy;
- improve timeout and stdout/stderr handling;
- validate scenario lock and report files more deeply.

Acceptance:

- scenario failures produce useful review evidence;
- drift exits with the documented code;
- command failure exits with the documented code;
- reports remain derived artifacts under `local/reports/*`.
- text summaries are written alongside structured YAML reports;
- `wkit validate` checks scenario report shape where report files exist.

## Milestone 5: Examples and Packaging

Status: done

Scope:

- example workspace;
- install instructions;
- changelog;
- release notes;
- CI for tests;
- release automation posture.

Acceptance:

- a new user can run a local example without author help;
- README contains truthful install/development instructions;
- release artifacts do not imply unproven adapter compatibility.
- CI runs Go tests, race tests, coverage, fuzz smoke checks, vulnerability scanning, Windows test/build coverage, builds the CLI, and smoke-tests the minimal and failure/drift examples.

Current packaging posture:

- source installs and local builds are documented;
- macOS/Linux release archive installs are supported by `install.sh`;
- Go module path is `github.com/johnkil/polyrepo-workspace-kit`;
- the minimal example includes a committed scenario artifact snapshot;
- GitHub Issues are documented for public support and non-sensitive conduct reports;
- GitHub private vulnerability reporting is documented for security reports;
- private conduct reports are directed to the maintainer email in `CODE_OF_CONDUCT.md`;
- Apache License 2.0 is recorded in `LICENSE`;
- release and versioning policy is documented in `docs/release.md`;
- clean-repo empirical Codex and Claude Code adapter notes are recorded in `research/empirical-agent-compatibility-matrix.md`;
- public binary release automation is configured for tagged draft GitHub Releases;
- GoReleaser builds Linux, macOS, and Windows archives for amd64 and arm64, plus checksums and build metadata;
- `install.sh` downloads tagged release archives, verifies `checksums.txt`, and installs into an existing `PATH` directory without requiring Go;
- `.github/workflows/test.yml` validates formatting, module tidiness, vet, lint, tests, race tests, coverage, fuzz smoke checks, vulnerability scanning, Windows test/build coverage, build, and the minimal and failure/drift examples without publishing artifacts.
- `.github/dependabot.yml` keeps Go module and GitHub Actions updates visible on a weekly cadence.

## Milestone 6: Orientation and Diagnostics

Status: done

Scope:

- `wkit context list`
- `wkit context show <context-id>`
- `wkit info` with alias `wkit overview`
- `wkit status [--context <context-id>]`
- `wkit scenario status <scenario-id>`
- `wkit doctor`

Acceptance:

- commands are read-only and do not clone, fetch, pull, push, switch branches, commit, or run scenario checks;
- status reports binding state, branch/detached state, short commit, dirty/untracked counts, upstream, and local ahead/behind counts when an upstream ref exists;
- scenario status reports `ok`, `drift`, `missing`, and `blocked` against the pinned lock without executing entrypoints;
- doctor combines manifest validation with local diagnostics for bindings, git checkouts, entrypoint `cwd` paths, and stale scenario locks;
- tests cover context listing/showing, info counts, status git states, scenario status drift/blocked states, and doctor exits/diagnostics.

## Milestone 7: Release Foundations

Status: done

Scope:

- `wkit version`
- `wkit --version`
- GoReleaser archive and checksum configuration
- tag-triggered GitHub release workflow
- local release validation targets

Acceptance:

- release builds embed version, commit, date, dirty state, and builder metadata;
- tagged `v*` workflow builds Linux, macOS, and Windows archives for amd64 and arm64;
- release workflow uploads `checksums.txt` and creates artifact attestations from it;
- GitHub Releases are draft by default for maintainer review before publishing;
- Homebrew, signing, notarization, OS packages, Scoop, and Winget remain deferred and documented honestly.

## Milestone 8: Binary Install UX

Status: done

Scope:

- root `install.sh` for macOS/Linux release archive installs;
- README and install docs that lead with the no-Go binary install path;
- `make check` coverage for installer syntax and help output.

Acceptance:

- installer resolves the latest release by default and supports explicit versions;
- installer detects macOS/Linux and amd64/arm64 archive names;
- installer verifies `checksums.txt` before writing;
- installer writes only into an absolute install directory and refuses symlink targets;
- the default install path uses an existing writable `PATH` directory, so shell startup edits are not required when such a directory exists.

## Milestone 9: VS Code Workspace Pilot

Status: done

Scope:

- `wkit vscode plan`
- `wkit vscode diff`
- `wkit vscode apply`
- `wkit vscode open`
- generated local VS Code multi-root workspace at
  `local/vscode/workspace.code-workspace`

Acceptance:

- generated output remains a local derived artifact and does not become
  canonical state;
- generated folders include the workspace root and every bound repo checkout;
- generated tasks include `wkit` orientation/diagnostic tasks, pinned scenario
  status/run tasks where locks exist, and repo entrypoint tasks from
  `repos/<repo-id>/repo.yaml`;
- repo entrypoint tasks use scoped `${workspaceFolder:<repo-id>}` `cwd`
  variables and preserve repo-local executable truth;
- missing bindings block rendering rather than silently producing an incomplete
  workspace;
- apply refuses to overwrite changed workspace files without `--force` or
  `--backup`;
- symlinked target files or parent paths that escape the workspace boundary are
  blocked before diff or write behavior reads or mutates them;
- `open` runs `code <workspace-file>` and only updates the generated file when
  explicitly confirmed.

## Milestone 10: Proof Hardening Artifacts

Status: in progress

Scope:

- committed failure/drift example artifacts that show why scenario evidence is
  useful beyond the happy path;
- spec text for `change` lifecycle boundaries;
- ADR for the scenario/CI boundary;
- `wkit handoff` command as a derived artifact, not a new canonical entity;
- reviewer-friendly markdown scenario reports;
- `wkit demo` first-run workflow that creates a temporary self-contained
  workspace without requiring a source checkout;
- scaffolded init flags for first real workspaces without hand-writing every
  manifest.

Acceptance:

- example artifacts include at least one drift-blocked check and one command
  failure with referenced stderr logs;
- example artifacts are decoded by tests so documentation cannot silently drift
  away from runtime report structs;
- `docs/spec.md` explicitly says that v0.x `change` state is local and
  declarative, with no PR/backend lifecycle tracking;
- the CI boundary ADR states that `wkit` does not own webhooks, daemons, remote
  schedulers, or hosted run history;
- `wkit handoff` remains a report/export command derived from existing
  `change`, `context`, `scenario`, and report data.
- scenario runs write a markdown report suitable for PR descriptions or chat,
  while YAML remains the structured report artifact.
- `wkit demo` can show both a passing scenario and failure/drift evidence from
  an installed binary.
- `wkit init` can register and bind repos, add explicit relations, create a
  context, and create an initial change from explicit flags without discovery or
  scenario execution.
- `wkit relations suggest` can inspect local dependency manifests and print
  missing relation candidates without writing canonical state.

## Milestone 11: Pilot Instrumentation

Status: done

Scope:

- local opt-in command event logging for pilot runs;
- telemetry config under `local/telemetry/config.yaml`;
- JSONL events under `local/telemetry/events.jsonl`;
- explicit status and export commands.

Acceptance:

- telemetry is disabled by default;
- enabling telemetry is a workspace-local action;
- recorded events include command path, captured args, exit code, timestamp, and
  duration;
- telemetry does not send data over the network, run a daemon, collect command
  output, or change command behavior when recording fails;
- export is explicit and prints the local JSONL event file.

### Proof Hardening Backlog

Implement in this order unless pilot evidence changes the priority:

1. done: failure/drift demo artifacts;
2. done: thin scaffolded init path;
3. done: `relations suggest` as an explicit suggestion-only workflow.

## Deferred

Do not implement in the near term:

- universal custom-agent schema;
- universal command model;
- plugin registry or marketplace;
- pack-first architecture;
- MCP bundle distribution;
- graph auto-discovery as canonical truth;
- tool-specific user-scope installs without empirical compatibility evidence;
- hosted runtime or policy layer.
