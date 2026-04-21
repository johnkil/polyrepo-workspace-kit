# Release Notes

## v0.3.0 - 2026-04-21

This release adds a local VS Code multi-root workspace export, a no-Go release
archive installer, and proof-hardening artifacts for scenario evidence and
pilot onboarding.

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
- Hardened scaffold validation, Gradle relation suggestions, telemetry coverage
  for invalid invocations, and handoff selection when same-second artifacts are
  present.

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
- `install.sh` downloads tagged release archives on macOS and Linux, verifies
  `checksums.txt`, and installs into an existing writable `PATH` directory.

### Distribution Notes

Supported install paths:

```bash
curl -fsSL https://raw.githubusercontent.com/johnkil/polyrepo-workspace-kit/main/install.sh | sh
```

```bash
go install github.com/johnkil/polyrepo-workspace-kit/cmd/wkit@v0.3.0
```

Tagged releases publish prebuilt archives and `checksums.txt` on the GitHub
Releases page.

Homebrew, signing, notarization, OS package managers, and tool-specific
user-scope installs remain deferred.

Release tags and readiness checks are documented in `docs/release.md`.
