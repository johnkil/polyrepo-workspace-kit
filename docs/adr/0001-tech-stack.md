# ADR 0001: CLI Tech Stack

Status: Accepted
Date: 2026-04-19

## Context

`wkit` is a local coordination CLI for polyrepo workspaces. It needs safe filesystem writes, YAML manifests, Git/subprocess interaction, fast startup, and low installation friction.

The project should avoid turning the CLI into a runtime ecosystem decision. Python is excellent for prototypes, but a production CLI should be easy to distribute as a native binary.

## Decision

Build the production CLI in Go.

Initial stack:

- Go for implementation.
- Cobra for command structure.
- `go.yaml.in/yaml/v3` for YAML parsing and writing.
- Go standard library for path handling, JSON, filesystem work, and subprocess execution.
- The real `git` binary for v0 Git state capture.

## Consequences

- The CLI can ship as a native binary later.
- Local development uses `go run ./cmd/wkit ...`.
- Release packaging can use GoReleaser when public distribution matters.
- Runtime plugin systems, embedded databases, and daemon behavior stay out of scope for v0.x.

## Notes

The reference Python implementation remains useful as a behavior prototype, but it is not the production stack decision.
