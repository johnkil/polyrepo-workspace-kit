# Install and Development
## Polyrepo Workspace Kit

Status: v0.x source install plus tagged release archive instructions

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
`make release-tools` installs the pinned GoReleaser version used by local release checks.
`make release-check` validates `.goreleaser.yaml`.
`make release-snapshot` builds local, unpublished release artifacts under `dist/`.

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

Install from source with:

```bash
go install github.com/johnkil/polyrepo-workspace-kit/cmd/wkit@latest
```

Source installs may report `wkit version` as `dev` when the Go toolchain does not embed module version metadata. Tagged release archives embed release version, commit, date, dirty state, and builder metadata.

## Release Archive Install

Tagged releases produced after release automation was added publish prebuilt archives for:

- Linux amd64 and arm64
- macOS amd64 and arm64
- Windows amd64 and arm64

Download the matching archive and `checksums.txt` from the GitHub Releases page:

```bash
version=0.y.z
os=darwin
arch=arm64
base="https://github.com/johnkil/polyrepo-workspace-kit/releases/download/v${version}"
curl -L -O "${base}/wkit_${version}_${os}_${arch}.tar.gz"
curl -L -O "${base}/checksums.txt"
grep " wkit_${version}_${os}_${arch}.tar.gz$" checksums.txt | shasum -a 256 -c -
tar -xzf "wkit_${version}_${os}_${arch}.tar.gz"
./wkit version
```

Windows archives use `.zip` instead of `.tar.gz`.
Use `arch=x86_64` for amd64 archives and `arch=arm64` for arm64 archives.

Release archives are not signed or notarized in this phase. On macOS, prefer source install or Homebrew once available if your local security policy requires signed/notarized CLI binaries.

## Release Tooling

Maintainers can validate release configuration locally:

```bash
make release-tools
make release-check
make release-snapshot
```

Pushing a `v*` tag runs the GitHub release workflow, builds release archives, writes `checksums.txt`, creates artifact attestations from that checksum file, and opens a draft GitHub Release for maintainer review.

## Not Yet Provided

The project does not yet ship:

- Homebrew formula;
- signed or notarized macOS binaries;
- compatibility guarantees for tool-specific user-scope installs.
