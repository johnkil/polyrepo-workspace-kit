# Agent Instructions

This file is hand-maintained repo-local guidance for coding agents working in
`polyrepo-workspace-kit`. It is not a generated adapter output and is not the
source of truth for the product model.

## Read First

- `README.md` for the product thesis, shipped CLI surface, and current status.
- `docs/spec.md` for the canonical model, file formats, CLI contracts, and safety rules.
- `docs/implementation-plan.md` for shipped milestones, deferred scope, and build posture.
- `docs/install.md` for local development commands.
- `CHANGELOG.md` before changing shipped behavior.

## Project Boundary

Keep the project a thin local coordination layer for polyrepo workspaces.

Do not expand the core model into:

- a monorepo manager;
- a developer portal;
- a retrieval/code graph engine;
- a universal agent, command, plugin, or pack system;
- a hosted runtime or policy platform.

Adapter outputs are derived artifacts. Canonical truth lives in workspace
manifests, repo manifests, guidance sources, and the spec.

## Coding Rules

- Prefer existing package boundaries under `internal/*`.
- Keep repo-local executable truth authoritative; do not centralize arbitrary repo commands beyond declared entrypoints.
- Validate any id before using it in a filesystem path.
- Treat installs and backups as safety-sensitive: avoid silent overwrite, path escape, partial backup, and mismatched plan/apply behavior.
- Keep scenario execution honest: local validation snapshot, not full environment reproduction.
- Preserve strict YAML decoding for manifest-driven inputs.
- Add focused regression tests for every safety fix.

## Commands

Useful local checks:

```bash
make tools
make check
make coverage
make fuzz
```

Fast focused checks while iterating:

```bash
go test ./...
go vet ./...
go build -o /tmp/wkit-review ./cmd/wkit
WKIT_BIN=/tmp/wkit-review sh examples/minimal-workspace/run-demo.sh
```

Use `make check` before considering a broad change ready. If tool installation
is unavailable, run the focused Go checks and state what could not be run.

## Documentation Rules

- Keep README, spec, implementation plan, release notes, and changelog aligned with actual behavior.
- Label planned, deferred, docs-backed, and empirically verified behavior clearly.
- Do not claim public binary distribution, Homebrew, GoReleaser, or tool-specific user-scope compatibility until implemented and evidenced.
- When changing report paths, backup naming, install targets, or exit codes, update both code tests and spec text.

## Review Posture

When reviewing changes, prioritize:

- path traversal and symlink escape;
- overwrite, backup, and partial-write safety;
- scenario lock/report correctness;
- stale docs that overclaim shipped behavior;
- missing regression tests for fixed bugs.
