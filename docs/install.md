# Install and Development
## Polyrepo Workspace Kit

Status: v0.x release archive installer, source install, and local development instructions

## Requirements

- Git for normal `wkit` workspace operations.
- macOS or Linux, plus `curl` or `wget`, `tar`, and `sha256sum` or `shasum`
  for the release archive installer.
- Go 1.25 or newer for source installs and local development.

`wkit` uses the real `git` binary for local repository state capture.

## Install a Prebuilt Binary

On macOS or Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/johnkil/polyrepo-workspace-kit/main/install.sh | sh
```

The installer:

- resolves the latest GitHub Release by default;
- downloads the matching macOS or Linux archive for `amd64`/`arm64`;
- verifies the archive against `checksums.txt`;
- installs `wkit` into the first writable absolute directory already on
  `PATH`;
- refuses to overwrite a symlink target.

No shell startup file changes are needed when the selected install directory is
already on `PATH`.

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/johnkil/polyrepo-workspace-kit/main/install.sh | sh -s -- --version v0.3.0
```

Install into an explicit directory:

```bash
curl -fsSL -o /tmp/wkit-install.sh https://raw.githubusercontent.com/johnkil/polyrepo-workspace-kit/main/install.sh
sudo WKIT_INSTALL_DIR=/usr/local/bin sh /tmp/wkit-install.sh
```

Use `WKIT_INSTALL_DIR` only for an absolute directory. If that directory is not
on `PATH`, the installer prints a warning and `wkit` will not be discoverable
until your shell can find that directory.

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
`make check` runs formatting, module tidiness, vet, lint, unit tests, race tests, vulnerability scanning, build, the minimal example, and the failure/drift example.
`make coverage` writes `coverage.out` and prints per-function coverage.
`make fuzz` runs a short fuzz pass for packages that define `Fuzz*` targets; override the duration with `FUZZTIME=30s`.
`make release-tools` installs the pinned GoReleaser version used by local release checks.
`make release-check` validates `.goreleaser.yaml`.

After installing or building `wkit`, run `wkit demo` or `wkit demo failure` to
create a temporary self-contained workspace and print the generated scenario
markdown report.
For a real workspace, use scaffold flags to avoid hand-writing the first set of
manifests:

```bash
wkit init ./workspace \
  --repo app-web=../app-web \
  --repo shared-schema=../shared-schema \
  --repo-kind shared-schema=contract \
  --relation app-web:shared-schema:contract \
  --context schema-rollout \
  --change-title "Payload field rollout"
```

Then inspect dependency manifests for relation candidates without writing the
canonical graph:

```bash
wkit --workspace ./workspace relations suggest
```

For pilot runs, optionally enable local command event logging. This writes only
inside the workspace under `local/telemetry/*` and exports only when you ask:

```bash
wkit --workspace ./workspace telemetry enable
wkit --workspace ./workspace telemetry export
```

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

Tagged releases publish prebuilt archives for:

- Linux amd64 and arm64
- macOS amd64 and arm64
- Windows amd64 and arm64

Download the matching archive and `checksums.txt` from the GitHub Releases page:

```bash
version=0.3.0
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
