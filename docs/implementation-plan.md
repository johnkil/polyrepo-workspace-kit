# Implementation Plan
## Polyrepo Workspace Kit

Status: Working engineering plan
Last updated: 2026-04-20

## Purpose

This document tracks implementation order.

It is different from `docs/plan.md`:

- `docs/plan.md` explains how to prove product value.
- `docs/implementation-plan.md` explains how to build the CLI in a disciplined order.

## Current Shipped Surface

The first Go implementation slice exists and is tested.

Implemented commands:

- `wkit init <path>`
- `wkit repo register <repo-id> --kind <kind>`
- `wkit bind set <repo-id> <path>`
- `wkit context list`
- `wkit context show <context-id>`
- `wkit info`
- `wkit overview`
- `wkit status [--context <context-id>]`
- `wkit doctor`
- `wkit version`
- `wkit --version`
- `wkit change new <context> --title <title>`
- `wkit change show <change-id>`
- `wkit scenario pin <scenario-id> --change <change-id>`
- `wkit scenario show <scenario-id>`
- `wkit scenario status <scenario-id>`
- `wkit scenario run <scenario-id>`
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
- `internal/fsutil`
- `internal/gitstate`
- `internal/install`
- `internal/manifest`
- `internal/model`
- `internal/orient`
- `internal/scenario`
- `internal/validate`
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
- CI runs Go tests, race tests, coverage, fuzz smoke checks, vulnerability scanning, Windows test/build coverage, builds the CLI, and smoke-tests the minimal example.

Current packaging posture:

- source installs and local builds are documented;
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
- `.github/workflows/test.yml` validates formatting, module tidiness, vet, lint, tests, race tests, coverage, fuzz smoke checks, vulnerability scanning, Windows test/build coverage, build, and the minimal example without publishing artifacts.
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
