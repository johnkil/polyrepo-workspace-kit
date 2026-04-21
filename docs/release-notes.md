# Release Notes

## Unreleased

This development line adds a local VS Code multi-root workspace export and
proof-hardening artifacts for scenario evidence.

### Highlights

- Added `wkit vscode plan`, `wkit vscode diff`, `wkit vscode apply`, and
  `wkit vscode open`.
- Generated `local/vscode/workspace.code-workspace` as a disposable local
  artifact with workspace folders and VS Code tasks derived from canonical
  `wkit` manifests.
- Added `docs/vscode.md` with the workflow, safety behavior, and smoke-test
  checklist for the VS Code export.
- Added a failure/drift example that shows scenario evidence for pinned ref
  drift and a failed repo-local check.
- Added ADR 0002 to keep `wkit scenario` bounded as local evidence rather than
  CI platform ownership.
- Added `wkit handoff <change-id>` to render a markdown handoff summary from a
  change, its context, a scenario lock, and the latest local report when
  present.
- Added markdown scenario run reports alongside the existing structured YAML and
  text summaries.
- Added `wkit demo [minimal|failure]` for a self-contained first-run demo from
  an installed binary.
- Added scaffold flags to `wkit init` so a first real workspace can register
  and bind repos, add explicit relations, create a context, and create an
  initial change without hand-writing every manifest.
- Added `wkit relations suggest` to inspect local dependency manifests and print
  missing relation candidates without writing canonical graph state.
- Added local opt-in pilot telemetry commands for workspace-local JSONL command
  event logs.
- Added `docs/pilot-kit.md` with the participant checklist, run sheet, evidence
  bundle template, and pilot pass/fail rubric.

### Behavior Notes

- The VS Code export requires local bindings for all declared repos before it
  renders a complete workspace file.
- Generated repo tasks use repo entrypoints from `repos/<repo-id>/repo.yaml`;
  `wkit` does not invent central repo commands.
- The export follows conservative overwrite behavior with `--force`,
  `--backup`, `--dry-run`, and `--yes`.
- `wkit scenario` remains local and explicit; teams may call it from their own
  CI scripts, but `wkit` does not own webhooks, daemons, hosted runners, or
  remote run history.
- YAML scenario reports remain the structured artifact. Text and markdown
  reports are derived review aids.
- `wkit demo` writes to a temporary directory and keeps that directory available
  for inspection; the printed paths point to disposable demo state.
- `wkit init` scaffold flags require explicit repo paths and relation
  declarations; they do not discover repositories, infer graph truth, pin
  scenarios, or run checks.
- `wkit relations suggest` is read-only and suggestion-only. It does not clone,
  fetch, run package managers, write manifests, or make discovered dependencies
  canonical by itself.
- `wkit telemetry enable` is disabled by default and writes only to
  `local/telemetry/*`; export is explicit and no network upload is performed.

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
