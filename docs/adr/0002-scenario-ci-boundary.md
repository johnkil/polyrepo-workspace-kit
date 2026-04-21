# ADR 0002: Scenario Is Not CI

Status: Accepted
Date: 2026-04-21

## Context

`wkit scenario` exists to create reviewable local evidence for coordinated
polyrepo work. It pins local repository revisions, runs declared repo-local
entrypoints, detects drift, and writes reports under the local workspace.

This overlaps with some CI vocabulary, but the product boundary is different.
If `wkit` starts owning remote events, hosted execution, or run orchestration, it
stops being a thin local coordination layer.

## Decision

`wkit` scenarios are local validation snapshots, not a CI platform.

For v0.x, `wkit` will not:

- run a daemon or background scheduler;
- listen for webhooks or remote repository events;
- register itself as a GitHub Actions, GitLab CI, Buildkite, or other CI
  integration owner;
- own remote workers, hosted runners, queues, retries, or notifications;
- store canonical run history outside the local workspace;
- infer scenario lifecycle from PR, issue, branch, merge, or release state.

The only supported way to produce scenario evidence is an explicit local command
such as:

```bash
wkit scenario run <scenario-id>
```

Teams may call that command from their own CI scripts if it is useful, but
`wkit` does not own that integration or become the CI control plane.

## Consequences

- Scenario reports remain derived local artifacts under `local/reports/*`.
- Repo-local entrypoints stay authoritative for build/test behavior.
- CI-specific features should be documented as user-owned integration examples,
  not promoted into core workflow ownership.
- Requests for daemon mode, webhooks, hosted run history, or PR-status ownership
  should be rejected or deferred unless the product boundary is explicitly
  revisited in a new ADR.
