# Minimal Workspace Example

This example creates a temporary local workspace with two repositories:

- `shared-schema`
- `app-web`

It demonstrates:

- workspace validation;
- local repo bindings;
- change creation;
- scenario pin and run;
- portable install into a repo checkout.
- local VS Code multi-root workspace export.

A committed sample of the scenario evidence is available under `artifacts/schema-rollout/`.

Run from the repository root:

```bash
sh examples/minimal-workspace/run-demo.sh
```

By default the script uses:

```bash
go run ./cmd/wkit
```

To test a built binary instead:

```bash
WKIT_BIN=/path/to/wkit sh examples/minimal-workspace/run-demo.sh
```

The script writes only to a temporary directory and prints the workspace path at the end.
