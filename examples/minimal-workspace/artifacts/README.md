# Scenario Artifact Snapshot

This directory contains a stable example of the evidence produced by:

```bash
wkit scenario pin schema-rollout --change CHG-2026-04-19-001
wkit scenario run schema-rollout
```

Files:

- `schema-rollout/manifest.lock.yaml` - pinned scenario snapshot.
- `schema-rollout/20260419T121000Z.yaml` - structured run report.
- `schema-rollout/20260419T121000Z.txt` - human-readable run summary.
- `schema-rollout/logs/20260419T121000Z/*` - stdout/stderr logs referenced by the report.

This is a committed sample artifact, not live output. Commit hashes, timestamps, durations, and tool versions are deterministic example values so the artifact stays reviewable in documentation.
