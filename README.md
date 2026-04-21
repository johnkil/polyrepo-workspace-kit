# Polyrepo Workspace Kit

[![Test](https://github.com/johnkil/polyrepo-workspace-kit/actions/workflows/test.yml/badge.svg)](https://github.com/johnkil/polyrepo-workspace-kit/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/johnkil/polyrepo-workspace-kit.svg)](https://pkg.go.dev/github.com/johnkil/polyrepo-workspace-kit)
[![Release](https://img.shields.io/github/v/release/johnkil/polyrepo-workspace-kit?display_name=tag&sort=semver)](https://github.com/johnkil/polyrepo-workspace-kit/releases)
[![License](https://img.shields.io/github/license/johnkil/polyrepo-workspace-kit)](LICENSE)

**`wkit` turns a cross-repo change into reviewable local evidence.**

Pin revisions across several local checkouts. Run each repo's own test command.
Get one report - YAML, plain text, and markdown - that says what passed, what
failed, and what drifted. Hand it to a reviewer or a coding agent as evidence
for a coordinated change.

## See the artifact first

After `wkit demo failure`, `local/reports/schema-rollout/<run-id>.txt` looks
like this:

```text
Scenario: schema-rollout
Results: passed=0 failed=1 blocked=1 skipped=0

- blocked: shared-schema:test (pinned ref drift: current HEAD 333333333333 does not match scenario lock 111111111111)
  env_profile: default
- failed: app-web:test (exit status 7)
  stdout: logs/20260419T122000Z/app-web-test.stdout.txt
  stderr: logs/20260419T122000Z/app-web-test.stderr.txt
  env_profile: default
```

Two different failure modes are caught in one run: `shared-schema` moved past
its pinned commit, and `app-web` ran but exited non-zero with stderr captured. A
committed sample is in
[`examples/failure-workspace/artifacts/`](examples/failure-workspace/artifacts/README.md),
so you can inspect the evidence without installing anything.

This is the center of the product. Not another `AGENTS.md` generator.

## Install and try

Install a prebuilt binary on macOS or Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/johnkil/polyrepo-workspace-kit/main/install.sh | sh
```

Install from source with Go:

```bash
go install github.com/johnkil/polyrepo-workspace-kit/cmd/wkit@latest
```

Then run the bundled demos. Both write only to a temporary directory:

```bash
wkit demo          # happy path: two repos, both tests pass
wkit demo failure  # drift in one repo, failed test in the other
```

## When `wkit` fits

Use `wkit` when several repositories change together and the current process is
Slack messages, tribal knowledge, ad hoc checklists, and rediscovered breakage
after handoff. Typical shapes:

- shared schema + service + generated SDK + docs + examples
- shared library + several consumers
- API contract + mobile app + web app + backend

You want to declare explicitly which repos are involved in this change, produce
evidence that they work together, and hand that evidence to a reviewer or an
agent without adopting a monorepo build system, a developer portal, or a hosted
PR platform.

## When `wkit` does not fit

Stated plainly so you do not waste time:

- **One repo.** Your existing tools are probably enough.
- **Cross-repo code search.** Use Sourcegraph, ripgrep, or your IDE's multi-root search.
- **Mechanical PR campaigns at scale.** Use [Sourcegraph Batch Changes](https://sourcegraph.com/docs/batch-changes) or [multi-gitter](https://github.com/lindell/multi-gitter).
- **Service catalog, ownership, on-call, or scorecards.** Use [Backstage](https://backstage.io/).
- **Just an AGENTS.md generator.** Use [agents.ge](https://agents.ge/).
- **A CI replacement.** `wkit` [explicitly is not one](docs/adr/0002-scenario-ci-boundary.md): no daemon, no webhooks, no hosted runners, no remote run history.

See [competitive research](research/competitors.md) for a more detailed
comparison.

## The model, in one paragraph

A **workspace** contains **repos** linked by **relations** (`contract`, `build`,
`runtime`, `release`, `docs`). A **context** names the subset of repos relevant
to a task. A **change** is a live cross-repo unit of work tied to a context. A
**scenario** pins involved repos to specific commits and runs each repo's
declared **entrypoint**. Commands stay repo-local; the workspace does not
centralize arbitrary build or test truth. **Bindings** map logical repo ids to
local checkout paths and live in `local/bindings.yaml` because paths are local
facts, not shared state.

Nine nouns total, [fully specified here](docs/spec.md). You do not need to learn
them all before starting; the scaffold flags below handle the first workspace.

## End-to-end workflow

```bash
# One command scaffolds a workspace, repos, relations, context, and first change.
INIT_OUTPUT="$(wkit init ./ws \
  --repo app-web=../app-web \
  --repo shared-schema=../shared-schema \
  --repo-kind shared-schema=contract \
  --relation app-web:shared-schema:contract \
  --context schema-rollout \
  --change-title "Payload field rollout")"
printf '%s\n' "$INIT_OUTPUT"
CHANGE_ID="$(printf '%s\n' "$INIT_OUTPUT" | awk '/^change:/ { print $2 }')"

# Suggest more relations from go.mod, build.gradle, package.json, and Cargo.toml.
# Read-only: you approve suggestions before they become canonical workspace state.
wkit --workspace ./ws relations suggest

# Pin and run the scenario.
wkit --workspace ./ws scenario pin schema-rollout --change "$CHANGE_ID"
wkit --workspace ./ws scenario run schema-rollout

# Produce a single markdown handoff with change + scenario + latest report.
wkit --workspace ./ws handoff "$CHANGE_ID" --scenario schema-rollout

# Optional: generate a VS Code multi-root workspace for the bound repos.
wkit --workspace ./ws vscode apply --yes
```

Reports, VS Code workspace files, and adapter outputs are derived artifacts.
Canonical truth lives in the shared manifests under `coordination/`, `repos/`,
and `guidance/`, plus machine-local bindings under `local/bindings.yaml`.

## Agent guidance (optional)

`wkit install plan|diff|apply <tool>` derives agent-readable files from
`guidance/rules/*` and `guidance/skills/*`. Supported in repo scope:

| tool | output |
| --- | --- |
| `portable` | `AGENTS.md`, `.agents/skills/*` |
| `codex` | same as portable |
| `opencode` | same as portable |
| `copilot` | `.github/copilot-instructions.md` |
| `claude` | `CLAUDE.md`, `.claude/skills/*` |

Every write is previewable with `install plan`, diffable with `install diff`,
and requires `install apply --yes`. Changed unmarked files are blocked unless
you pass `--force` or `--backup`. See
[empirical compatibility notes](research/empirical-agent-compatibility-matrix.md)
for which tool versions have been probed against these targets.

Tool-specific user-scope targets are deliberately deferred. Portable user scope
is skills-only: `.agents/skills/*`.

## Safety and scope

- `wkit scenario run` does not clone, fetch, install packages, upload data,
  start daemons, or call remote services as part of its own orchestration.
- Repo-local entrypoints are still your commands; if a script reaches the
  network, that is repo behavior, not `wkit` orchestration.
- No daemon, background scheduler, webhook listener, hosted runner, or remote
  run history.
- `bind set` rejects missing paths before writing.
- Scenario commands with shell quoting are rejected; use a repo-local wrapper
  script.
- Scenario `cwd` values and symlinks that escape a repo checkout are rejected
  before command execution.
- `install apply` and `vscode apply` will not write without explicit `--yes`.
- Pilot telemetry is opt-in, local-only, and never exported unless you run
  `wkit telemetry export`.

More in [`SECURITY.md`](SECURITY.md).

## Status

**v0.3.0 - proof stage.** The CLI is usable, tested, and release-packaged, but
the project is not MVP-proven until independent pilots produce evidence.

- Core workspace model, scenarios, reports, portable install, repo-scope
  adapters, VS Code export, demos, handoff, relation suggestions, and local
  telemetry are shipped.
- Two example workspaces, minimal and failure, are exercised in CI.
- Empirical compatibility probes: Codex 0.117.0 and Claude Code 2.1.90 done.
  OpenCode and Copilot are pending.
- Non-author pilots: **actively recruiting**. If you run a repeated polyrepo
  workflow and want to try `wkit` end-to-end, see
  [`docs/pilot-kit.md`](docs/pilot-kit.md).
- Homebrew packaging, signing/notarization, OS packages, and tool-specific
  user-scope installs are deferred.

During v0.x, minor releases may change command, manifest, adapter, or validation
behavior. Breaking changes are called out in [`CHANGELOG.md`](CHANGELOG.md).

## Documentation

- [Product Requirements](docs/prd.md)
- [RFC: Core Model and Layering](docs/rfc.md)
- [Technical Specification](docs/spec.md)
- [Proof and Pilot Plan](docs/plan.md)
- [Pilot Kit](docs/pilot-kit.md)
- [Install and Development](docs/install.md)
- [VS Code Workspace Export](docs/vscode.md)
- [Release and Versioning](docs/release.md)
- [ADR 0001: CLI Tech Stack](docs/adr/0001-tech-stack.md)
- [ADR 0002: Scenario Is Not CI](docs/adr/0002-scenario-ci-boundary.md)
- [Competitive Research](research/competitors.md)
- [Empirical Compatibility Matrix](research/empirical-agent-compatibility-matrix.md)

## Contributing

Issues, PRs, and pilot reports are welcome. See [`CONTRIBUTING.md`](CONTRIBUTING.md)
and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md). For security issues, see
[`SECURITY.md`](SECURITY.md).

## License

Apache 2.0. See [`LICENSE`](LICENSE).
