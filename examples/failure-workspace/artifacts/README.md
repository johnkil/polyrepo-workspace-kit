# Failure Scenario Artifact Snapshot

This directory contains a stable example of failure evidence produced by:

```bash
wkit scenario pin schema-rollout --change CHG-2026-04-19-001
wkit scenario run schema-rollout
```

The represented run has two outcomes:

- `shared-schema:test` is blocked by pinned ref drift.
- `app-web:test` fails and records stdout/stderr logs.

Files:

- `schema-rollout/manifest.lock.yaml` - pinned scenario snapshot.
- `schema-rollout/20260419T122000Z.yaml` - structured failure run report.
- `schema-rollout/20260419T122000Z.txt` - human-readable failure run summary.
- `schema-rollout/20260419T122000Z.md` - reviewer-friendly markdown summary
  with a stderr excerpt.
- `schema-rollout/logs/20260419T122000Z/*` - stdout/stderr logs referenced by
  the failed check.

This is a committed sample artifact, not live output. Commit hashes,
timestamps, durations, and tool versions are deterministic example values so the
artifact stays reviewable in documentation.
