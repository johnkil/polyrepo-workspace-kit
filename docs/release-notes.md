# Release Notes

## v0.2.0 - 2026-04-20

This release adds the first read-only orientation layer and the first tagged
release archive automation for `wkit`.

## Highlights

- Added workspace orientation commands:
  `wkit context list`, `wkit context show`, `wkit info`, and
  `wkit overview`.
- Added local diagnostics commands:
  `wkit status`, `wkit scenario status`, and `wkit doctor`.
- Added release identity commands:
  `wkit version` and `wkit --version`.
- Added GoReleaser-based tagged release automation for draft GitHub Releases
  with Linux/macOS/Windows archives, SHA-256 checksums, and checksum-based
  artifact attestations.
- Added GitHub project hygiene improvements: issue forms, PR template, status
  badges, social preview, repository discoverability notes, and GitHub Pages
  landing page.

## Behavior Notes

- `wkit status`, `wkit scenario status`, and `wkit doctor` inspect local truth
  without fetching remotes, mutating checkouts, or running scenario checks.
- Release archives embed `wkit` version, commit, build date, dirty state, and
  builder metadata.
- GitHub Releases created by the tag workflow are drafts by default so
  maintainers can inspect artifacts before publishing.

## Distribution Notes

Supported install paths:

```bash
go install github.com/johnkil/polyrepo-workspace-kit/cmd/wkit@latest
```

Tagged releases produced after this release automation was added can publish
prebuilt archives and `checksums.txt` on the GitHub Releases page.

Homebrew, signing, notarization, OS package managers, and tool-specific
user-scope installs remain deferred.

Release tags and readiness checks are documented in `docs/release.md`.
