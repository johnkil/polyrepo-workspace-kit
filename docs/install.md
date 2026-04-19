# Install and Development
## Polyrepo Workspace Kit

Status: v0.x source-first install instructions

## Requirements

- Go 1.25 or newer
- Git

`wkit` uses the real `git` binary for local repository state capture.

## Local Development

From the repository root:

```bash
make tools
make check
make coverage
make fuzz
go run ./cmd/wkit --help
go test ./...
```

`make tools` installs the local Go developer tools used by `make check`, including the pinned `golangci-lint` version.
`make check` runs formatting, module tidiness, vet, lint, unit tests, race tests, vulnerability scanning, build, and the minimal example.
`make coverage` writes `coverage.out` and prints per-function coverage.
`make fuzz` runs a short fuzz pass for packages that define `Fuzz*` targets; override the duration with `FUZZTIME=30s`.

Run the example workspace:

```bash
sh examples/minimal-workspace/run-demo.sh
```

## Build a Local Binary

```bash
go build -o bin/wkit ./cmd/wkit
./bin/wkit --help
```

Use the built binary with the example:

```bash
WKIT_BIN="$(pwd)/bin/wkit" sh examples/minimal-workspace/run-demo.sh
```

## Source Install

When this repository is available from a Git remote, early adopters can install from source with:

```bash
go install github.com/johnkil/polyrepo-workspace-kit/cmd/wkit@latest
```

## Not Yet Provided

The project does not yet ship:

- prebuilt release binaries;
- Homebrew formula;
- signed or notarized macOS binaries;
- GoReleaser artifacts;
- compatibility guarantees for tool-specific user-scope installs.
