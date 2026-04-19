# Release Notes
## Unreleased v0.x

These notes describe the current unreleased implementation state.

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

Tool-specific user-scope installs are not implemented in this release state.

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

Prebuilt binaries, Homebrew, signing, notarization, and GoReleaser packaging are deferred until public distribution is justified.

Release tags and readiness checks are documented in `docs/release.md`.
