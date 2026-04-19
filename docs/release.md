# Release and Versioning

Status: v0.x source-first policy

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

The current distribution mode is source-first.

Supported install path after the repository is published:

```bash
go install github.com/johnkil/polyrepo-workspace-kit/cmd/wkit@latest
```

Prebuilt binaries, Homebrew packaging, signing, notarization, SBOMs, and provenance attestations are deferred until public distribution is justified.

## Release Readiness Checklist

Before tagging a release:

- Run `make tools`.
- Run `make check`.
- Run `make coverage`.
- Run `make fuzz` when YAML, id, path, scenario, or report safety changed.
- Confirm `CHANGELOG.md` describes user-visible changes.
- Confirm `docs/release-notes.md` matches shipped behavior.
- Confirm `README.md` install and status sections are truthful.
- Confirm compatibility claims are backed by `research/empirical-agent-compatibility-matrix.md`.

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

Do not create binary artifacts manually unless the release process explicitly documents them.

## Deferred Release Automation

When public distribution needs prebuilt artifacts, add GoReleaser with:

- Linux, macOS, and Windows builds for amd64 and arm64 where practical;
- checksums;
- changelog generation from committed release notes;
- SBOM and provenance/attestation support;
- Homebrew tap only after user demand is proven.
