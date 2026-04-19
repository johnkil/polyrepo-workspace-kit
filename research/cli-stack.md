# CLI Stack Research

Date: 2026-04-19
Scope: language and implementation stack choice for the first `wkit` CLI.

## Question

What language and stack should `wkit` use for the first production CLI implementation?

## Short Answer

Use **Go** for the first production CLI.

Recommended stack:

- Language: Go
- CLI framework: Cobra
- YAML: `go.yaml.in/yaml/v3`
- JSON/report encoding: Go standard `encoding/json`
- Git integration: shell out to `git` via `os/exec`, do not use an embedded Git implementation in v0
- Path handling: Go standard `path/filepath`
- Release: GitHub Releases plus GoReleaser when distribution hardens
- Install path for early adopters: `go install .../cmd/wkit@latest`, then prebuilt binaries/Homebrew later

This is the most boring stack, and that is the point. `wkit` should feel like a small trustworthy local tool, not a runtime ecosystem decision.

## Decision Criteria

The CLI needs to optimize for:

1. Low installation friction.
2. Cross-platform behavior across macOS, Linux, and Windows.
3. Safe file writes and path handling.
4. Subprocess execution for repo-local entrypoints.
5. Git state capture from real local checkouts.
6. YAML/JSON manifest parsing and validation.
7. Fast startup.
8. Small dependency surface.
9. Easy release of native binaries.
10. Maintainability by a small team.

It does **not** need:

- embedded daemon runtime;
- plugin runtime;
- web server;
- TUI;
- remote service;
- rich async networking;
- language-specific integration with agent tools.

## Stack Comparison

| Option | Fit | Why |
| --- | --- | --- |
| Go | Best default | Native binaries, simple cross-build story, strong standard library for files/paths/processes, mature CLI ecosystem, good enough type system. |
| Rust | Strong but heavier | Excellent CLI and binary story, strongest correctness story, but slower iteration and a less settled YAML ecosystem. |
| Python | Great prototype, weaker product CLI | Fastest to build validators and workflow logic, but runtime and tool-environment management add adoption friction. |
| TypeScript/Node | Good npm CLI, weaker local-tool default | Strong developer familiarity and npm distribution, but Node runtime/dependency surface is heavier. Single executable support exists but is still marked active development in Node docs. |
| Deno/Bun | Interesting, not default | Standalone TypeScript binaries are attractive, but the ecosystem is younger and less boring for an infrastructure-style CLI. |

## Why Go Wins For `wkit`

### 1. Distribution Matches The Product

`wkit` should be easy to drop into any local workspace. Go produces native executable binaries with the standard `go build`/`go install` flow. The Go docs explicitly separate `go build` for compiling a binary from `go install` for installing it on the user's PATH.

Go also has a first-class cross-compilation model through `GOOS` and `GOARCH`, covering the operating systems and architectures relevant for a local developer CLI.

### 2. The Standard Library Covers The Risky Parts

The risky parts of `wkit` are not exotic algorithms. They are local filesystem behavior, path portability, subprocess execution, and JSON output.

Go has standard packages for:

- `os/exec` for running external commands without invoking a shell by default;
- `path/filepath` for OS-compatible path handling;
- `encoding/json` for machine-readable reports.

That maps directly to `wkit` needs: run repo-local entrypoints, avoid shell injection, handle Windows/macOS/Linux paths, and emit stable reports.

### 3. Cobra Fits The Command Shape

The planned CLI has nested command groups:

```text
wkit repo register
wkit bind set
wkit change new
wkit scenario pin
wkit install plan
wkit install apply
```

Cobra is built for subcommand CLIs and provides help generation, POSIX-style flags, nested subcommands, shell completion, and man page generation. It is used by large Go CLIs such as Kubernetes, Hugo, and GitHub CLI, which is useful category evidence for this style of tool.

### 4. YAML Is Less Awkward In Go Than It Currently Is In Rust

`wkit` already chose YAML for human-authored manifests. Go now has `go.yaml.in/yaml/v3`, maintained by the official YAML organization after the old `go-yaml` project became unmaintained. The package supports decoding into structs and has a stable v3 API.

Rust has excellent `serde`, but the historical `serde_yaml` path is weaker now. There are alternatives and forks, but for this project the YAML layer should be boring and low-risk.

### 5. Go Encourages The Right Amount Of Abstraction

The biggest product risk is overbuilding. Go is a good constraint for this project:

- explicit structs for manifests;
- explicit validators;
- explicit command execution;
- explicit error handling;
- small internal packages.

That nudges the implementation toward the product's philosophy: narrow control, inspectable behavior, no hidden platform.

## Why Not Rust First

Rust is a credible second choice.

Use Rust if the primary goal becomes:

- maximal type-level correctness;
- smallest standalone binary;
- strict memory safety story;
- long-term infrastructure credibility for a technically deep audience.

Do not pick Rust first if the main objective is quickly proving the `change` + `scenario` loop with pilots. Build time, contributor ramp, and YAML ecosystem friction are real costs.

If Rust is chosen anyway, the likely stack is:

- `clap` for CLI parsing;
- `serde` for data modeling;
- a maintained YAML parser/fork selected after a separate YAML parser decision;
- `assert_cmd`/`insta` style snapshot tests;
- `cargo-dist` or GoReleaser for binary distribution.

## Why Not Python First

Python is the best language for a throwaway prototype or research harness.

A Python prototype stack would be:

- Typer or Click for CLI;
- Pydantic for manifest models and JSON Schema output;
- PyYAML or ruamel.yaml for YAML;
- `subprocess` for commands;
- `uv tool install` or `pipx` for installation.

But the product CLI should not require users to think about Python environments. The `uv` tool model is strong, but it still introduces tool environments, Python version selection, executable directory setup, and version caching rules. That is acceptable for Python developers, less ideal for a workspace coordination CLI that should be boring for everyone.

## Why Not TypeScript/Node First

Node/TypeScript is good when:

- npm distribution is the main install path;
- the product is mostly adapter generation for agent ecosystems;
- rich string/template tooling matters more than native binary distribution.

For `wkit`, those are secondary. The core product is local coordination and scenario validation. Requiring Node is extra friction for Go, Rust, Swift, Java, or mobile teams.

Node single executable applications are real, but the official Node docs still mark the feature as active development. That makes it less attractive as the default for a trust-oriented local infrastructure CLI.

If TypeScript is chosen anyway, prefer:

- Commander or oclif;
- Zod or JSON Schema validation;
- npm package `bin` for early distribution;
- Node SEA, Deno compile, or Bun compile only after separate packaging validation.

## Proposed Go Architecture

```text
cmd/
  wkit/
    main.go
internal/
  cli/
  workspace/
  manifest/
  validate/
  gitstate/
  scenario/
  install/
  adapters/
  report/
  fsutil/
  testutil/
schemas/
  workspace.schema.json
  repo.schema.json
  scenario-lock.schema.json
testdata/
  workspaces/
```

Package responsibilities:

- `cli`: Cobra command tree and flag wiring only.
- `workspace`: workspace discovery and initialization.
- `manifest`: load/save YAML and JSON.
- `validate`: typed validation with high-quality errors.
- `gitstate`: `git -C <path>` wrappers for commit, branch, status, and dirty files.
- `scenario`: pin/run behavior.
- `install`: plan/diff/apply, ownership markers, backup logic.
- `adapters`: portable/Codex/OpenCode/Copilot/Claude target generation.
- `report`: text-first and machine-readable reports.
- `fsutil`: atomic writes, path checks, backup naming, safe relative paths.

## Dependency Policy

Start narrow:

- `github.com/spf13/cobra`
- `go.yaml.in/yaml/v3`
- optional small diff library only if a simple internal unified diff becomes painful

Avoid in v0:

- Viper, unless there is a real multi-source config need;
- TUI frameworks;
- embedded database;
- embedded Git implementation;
- plugin runtime;
- template engines beyond the standard library unless adapters prove they need more.

## Git Strategy

Use the real `git` binary in v0.

Reasoning:

- `wkit` needs the same behavior developers see in their checkout;
- repo-local Git config, submodules, worktrees, and ignored files are real local facts;
- `git status --porcelain=v1` and `git rev-parse` are stable enough for the first implementation;
- shelling out through `os/exec` avoids invoking a shell by default.

Avoid `go-git` in v0. It is useful for pure library workflows, but `wkit` is coordinating real local checkouts, not replacing the Git CLI.

## Validation Strategy

Use Go structs plus explicit validators for v0. Keep JSON Schema files as documentation and compatibility artifacts.

For example:

- parse YAML into typed structs;
- reject unknown fields where feasible;
- validate cross-references in Go;
- produce deterministic error codes/messages;
- use `schemas/*.schema.json` in tests and docs, not as the only runtime validation mechanism.

This avoids making JSON Schema tooling the center of the product while still preserving a portable schema story.

## Release Strategy

Phase 1:

- source install: `go install .../cmd/wkit@latest`;
- local dev: `go run ./cmd/wkit ...`;
- CI: Go test on macOS, Linux, Windows.

Phase 2:

- GoReleaser builds for darwin/linux/windows and amd64/arm64;
- GitHub Release artifacts;
- checksums;
- Homebrew tap after the first pilot users need it.

Phase 3:

- signed/notarized macOS binaries if public adoption justifies it;
- package managers only where demand is proven.

## Recommendation

Pick **Go + Cobra** for the production CLI.

Use Python only for quick throwaway experiments, if needed. Reconsider Rust only if the first pilot proves that strict correctness and binary polish matter more than iteration speed. Reconsider TypeScript only if the product pivots from coordination into agent adapter distribution.

## Sources

- Go compile/install tutorial: https://go.dev/doc/tutorial/compile-install
- Go cross-compilation environment: https://go.dev/doc/install/source#environment
- Go `os/exec`: https://pkg.go.dev/os/exec
- Go `path/filepath`: https://pkg.go.dev/path/filepath
- Go `encoding/json`: https://pkg.go.dev/encoding/json
- Cobra docs: https://cobra.dev/docs/
- Cobra package docs: https://pkg.go.dev/github.com/spf13/cobra
- Go YAML v3 package: https://pkg.go.dev/go.yaml.in/yaml/v3
- GoReleaser Go build docs: https://goreleaser.com/customization/builds/builders/go/
- Rust clap docs: https://docs.rs/clap/latest/clap/
- Cargo install docs: https://doc.rust-lang.org/cargo/commands/cargo-install.html
- Python argparse docs: https://docs.python.org/3/library/argparse.html
- Python entry points spec: https://packaging.python.org/en/latest/specifications/entry-points/
- uv tools docs: https://docs.astral.sh/uv/concepts/tools/
- Typer docs: https://typer.tiangolo.com/
- Click docs: https://pocoo-click.readthedocs.io/
- npm package `bin` docs: https://docs.npmjs.com/cli/v10/configuring-npm/package-json#bin
- oclif docs: https://oclif.io/docs/introduction/
- Node single executable applications: https://nodejs.org/api/single-executable-applications.html
- Deno compile docs: https://docs.deno.com/runtime/reference/cli/compile/
- Bun single-file executable docs: https://bun.sh/docs/bundler/executables
