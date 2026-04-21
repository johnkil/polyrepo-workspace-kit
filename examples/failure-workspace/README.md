# Failure Workspace Example

This example reuses the minimal two-repo workspace and then intentionally makes
the pinned scenario go stale:

- `shared-schema` gets a new commit after `scenario pin`, producing pinned ref
  drift;
- `app-web` gets a failing local test command after `scenario pin`, producing a
  command failure with stderr evidence.

It demonstrates why scenario reports are useful beyond a happy-path validation
run.

A committed sample of the failure evidence is available under
`artifacts/schema-rollout/`.

Run from the repository root:

```bash
sh examples/failure-workspace/run-demo.sh
```

By default the script uses:

```bash
go run ./cmd/wkit
```

To test a built binary instead:

```bash
WKIT_BIN=/path/to/wkit sh examples/failure-workspace/run-demo.sh
```

The script writes only to a temporary directory and prints the workspace path at
the end.
