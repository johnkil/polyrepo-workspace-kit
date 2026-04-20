# Changelog

All notable changes to this project will be documented here.

The format is intentionally lightweight during v0.x. Public releases should keep this file truthful about shipped behavior.

## Unreleased

### Added

- Added a macOS/Linux `install.sh` for checksum-verified release archive installs without requiring Go.

### Fixed

- Allowed custom release notes passed to GoReleaser to populate future draft GitHub Releases.

## v0.2.0 - 2026-04-20

### Added

- Added GitHub issue forms and a pull request template for clearer community input.
- Added README status badges and promoted the source-first install command in Quick Start.
- Added a GitHub social preview asset and repository discoverability checklist.
- Added a lightweight GitHub Pages landing page with basic SEO metadata, robots.txt, and sitemap.xml.
- Added a dedicated GitHub Pages workflow for publishing the static docs landing page.
- Switched the GitHub Pages landing page and social preview asset to a dark-first visual design with a CLI/workspace preview.
- Added a dedicated README hero image using the dark CLI/workspace visual system.
- Added a competitive research note distinguishing `wkit` from bulk polyrepo Git managers such as `polyrepopro/polyrepo`.
- Added feature-level research on which bulk polyrepo Git manager ideas should inspire future `wkit` status, doctor, overview, and scenario drift workflows.
- Added read-only orientation commands: `context list`, `context show`, `info`/`overview`, `status`, `scenario status`, and `doctor`.
- Added local checkout diagnostics for git state, dirty/untracked counts, upstream/ahead/behind state, stale scenario locks, missing bindings, non-git checkouts, and invalid entrypoint `cwd` paths.
- Added `wkit version` with build metadata for release archives and source builds.
- Added GoReleaser release configuration, local release-check targets, and a tag-triggered draft GitHub Release workflow with archives, checksums, and checksum-based artifact attestations.

### Changed

- Updated GitHub Actions workflow dependencies to Node.js 24-compatible `actions/checkout@v6` and `actions/setup-go@v6`, and replaced `golangci-lint` and `govulncheck` workflow actions with direct Go commands to avoid stale action runtimes.

### Fixed

- Blocked installer source and target symlink escapes before planning diffs, backups, or derived guidance writes.

## v0.1.0 - 2026-04-19

### Added

- Go CLI entrypoint: `cmd/wkit`.
- Core workspace commands: `init`, `repo register`, `bind set`, `validate`.
- Change commands: `change new`, `change show`.
- Scenario commands: `scenario pin`, `scenario show`, `scenario run`.
- Scenario YAML reports and text summaries under `local/reports/*`.
- Portable install commands: `install show-targets`, `install plan`, `install diff`, `install apply`.
- Repo-scope adapter targets for `codex`, `opencode`, `copilot`, and `claude`.
- Install safety behavior for blocked targets, `--force`, `--backup`, `--dry-run`, and `--yes`.
- Minimal runnable example workspace.
- ADR for the CLI tech stack.
- Go hygiene checks for formatting, module tidiness, vet, linting, tests, build, and the minimal example.
- Race-detector and `govulncheck` checks for the Go CLI.
- Windows CI coverage for Go tests and CLI build.
- Coverage target and CI coverage smoke check.
- Fuzz targets and CI fuzz smoke check for YAML decoding, ids, and path safety.
- Dependabot configuration for Go modules and GitHub Actions.
- Contributing, security, support, and code of conduct documentation.
- Apache License 2.0 project license.
- Public Go module path `github.com/johnkil/polyrepo-workspace-kit`.
- Minimal scenario artifact snapshot for documentation and review examples.
- GitHub Issues support channel and GitHub private vulnerability reporting guidance.
- Release and versioning policy for source-first v0.x releases.
- Clean-repo empirical compatibility pass for Codex and Claude Code adapter behavior.
- Conduct reporting policy without publishing a personal email address.

### Fixed

- Validated all coordination rule files, including orphan files not referenced from `workspace.yaml`.
- Replaced fragile Git status parsing with NUL-delimited porcelain v2 parsing for paths with spaces and renames.
- Made `bind set` reject missing paths and file paths before saving a local binding.
- Rejected quoted scenario commands during validation, pinning, and run preflight so users move shell-sensitive commands into repo-local scripts.
- Hardened file-backed ids against path traversal before constructing workspace paths.
- Blocked scenario `cwd` values that escape the bound repo checkout.
- Prevented `change new` from reusing an existing change id when daily numbering has gaps.
- Avoided backup filename collisions during `install apply --backup`.
- Made backup creation reject existing backup destinations with exclusive writes.
- Blocked scenario `cwd` symlinks that resolve outside the bound repo checkout.
- Avoided same-second report/log collisions during `scenario run`.
- Made file installs use atomic copy semantics while preserving source file mode.
- Made exclusive file writes and directory backups stage through temporary paths and clean up newly created backup paths on failure.
- Made scenario command failures take precedence over blocked/drift state for CLI exit code `5`.
- Strengthened scenario validation and custom YAML field strictness.
- Handled CLI output write errors and read-close cleanup found by `errcheck`.
- Added a test that decodes the committed scenario artifact snapshot.
- Aligned CI and lint `goimports` local-prefix configuration with the public Go module path.

### Not Yet Included

- Public release binaries.
- Tool-specific user-scope installs.
- GoReleaser packaging.
- Homebrew packaging.
- Empirically verified adapter compatibility claims.
