# Release and Versioning

Status: v0.x release archive installer plus source install policy

## Versioning

Polyrepo Workspace Kit uses SemVer tags with a `v` prefix:

```text
v0.y.z
```

During `v0.x`, minor versions may include command, manifest, adapter, or validation changes. Patch versions should be bug fixes, documentation corrections, compatibility-note updates, or release metadata fixes.

Before `v1.0.0`, compatibility promises are intentionally narrow:

- canonical workspace state should remain explicit and reviewable;
- adapter outputs remain derived artifacts;
- breaking changes must be called out in `CHANGELOG.md`;
- empirical adapter compatibility claims require recorded probe evidence.

## Current Distribution

The lowest-friction macOS and Linux install path is the release archive
installer:

```bash
curl -fsSL https://raw.githubusercontent.com/johnkil/polyrepo-workspace-kit/main/install.sh | sh
```

It downloads the latest GitHub Release archive, verifies `checksums.txt`, and
installs `wkit` into a writable directory already on `PATH`.

Source install remains available when Go is already present:

```bash
go install github.com/johnkil/polyrepo-workspace-kit/cmd/wkit@latest
```

For tags produced after release automation was introduced, the release workflow
creates draft GitHub Releases with:

- Linux, macOS, and Windows archives for amd64 and arm64;
- a `checksums.txt` file using SHA-256;
- artifact attestations generated from `dist/checksums.txt`;
- embedded `wkit version` metadata for version, commit, date, dirty state, and builder.

Homebrew packaging, signing, notarization, SBOM files, deb/rpm/apk packages, Scoop, and Winget are deferred.

## Release Readiness Checklist

Before tagging a release:

- Run `make tools`.
- Run `make check`.
- Run `make coverage`.
- Run `make fuzz` when YAML, id, path, scenario, or report safety changed.
- Confirm `CHANGELOG.md` describes user-visible changes.
- Confirm `docs/release-notes.md` matches shipped behavior.
- Confirm `README.md` install and status sections are truthful.
- Confirm `install.sh --help` and release snapshot checks pass through `make check` and `make release-snapshot`.
- Confirm compatibility claims are backed by `research/empirical-agent-compatibility-matrix.md`.
- Run `make release-tools` if GoReleaser is not installed.
- Run `make release-check`.
- Run `make release-snapshot`.

## Tagging Flow

1. Choose the next SemVer tag.
2. Move relevant `CHANGELOG.md` entries from `Unreleased` into the release section.
3. Update `docs/release-notes.md` for the release.
4. Run the release readiness checklist.
5. Create an annotated tag:

```bash
git tag -a v0.y.z -m "v0.y.z"
git push origin v0.y.z
```

The tag workflow creates a draft GitHub Release. Review the draft release notes, archives, checksums, and workflow result before publishing it.

Do not create binary artifacts manually unless the release process explicitly documents them.

## Homebrew Posture

Homebrew is useful for user experience, but it is intentionally not shipped in this release automation pass.

The preferred next step is a dedicated tap repository with a source-build formula such as `johnkil/homebrew-tap`. GoReleaser's binary Homebrew cask support should wait until the project has a signing/notarization decision or a deliberately documented unsigned-binary caveat.
