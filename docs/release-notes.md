# Release Notes

## Unreleased

- Added read-only orientation and diagnostics commands:
  `wkit context list`, `wkit context show`, `wkit info`/`wkit overview`,
  `wkit status`, `wkit scenario status`, and `wkit doctor`.
- `wkit status`, `wkit scenario status`, and `wkit doctor` inspect local truth
  without fetching remotes, mutating checkouts, or running scenario checks.
- Added `wkit version` and release build metadata.
- Added GoReleaser-based tagged release automation for draft GitHub Releases
  with Linux/macOS/Windows archives, checksums, and checksum-based artifact
  attestations.

## v0.1.0 - 2026-04-19

These notes describe the first source-first public release.

## Highlights

- `wkit` now has a Go CLI implementation.
- The core local workflow is runnable end to end:
  - initialize or open a workspace;
  - register repositories;
  - bind local checkouts;
  - create a change;
  - pin and run a scenario;
  - inspect scenario evidence under `local/reports/*`.
- Portable and repo-scope adapter install flows support plan, diff, and apply safety.
- A minimal example workspace can be run with:

```bash
sh examples/minimal-workspace/run-demo.sh
```

- A stable scenario artifact snapshot is committed under `examples/minimal-workspace/artifacts/schema-rollout/`.

## Compatibility Notes

Adapter target paths are still docs-backed unless a compatibility probe records tool version, probe date, target path, and observed behavior.

Tool-specific user-scope installs are not implemented in this release.

## Distribution Notes

This is still source-first.
The project is licensed under Apache License 2.0. Support uses GitHub Issues, security reports use GitHub private vulnerability reporting, and non-sensitive conduct reports use GitHub Issues.

Supported local commands:

```bash
make tools
make check
make coverage
make fuzz
go run ./cmd/wkit --help
go install github.com/johnkil/polyrepo-workspace-kit/cmd/wkit@latest
go test ./...
go build -o bin/wkit ./cmd/wkit
```

CI currently runs the Go hygiene suite on supported Go release lines, plus race tests, coverage, fuzz smoke checks, Windows test/build coverage, and `govulncheck`.

Homebrew, signing, notarization, and OS package managers remain deferred. Tagged releases produced after release automation was added can publish draft GitHub Releases with prebuilt archives and checksums.

Release tags and readiness checks are documented in `docs/release.md`.
