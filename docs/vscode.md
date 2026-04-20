# VS Code Workspace Export

`wkit` can generate a local VS Code multi-root workspace file from canonical
workspace manifests and local bindings.

The generated file is:

```text
local/vscode/workspace.code-workspace
```

It is a derived artifact. Do not edit it by hand and do not treat it as
canonical state.

## Why this exists

VS Code multi-root workspaces are a natural fit for polyrepo work:

- the Explorer can show each bound repository as a separate root;
- global search spans all opened roots and groups results by folder;
- Source Control can show multiple Git repositories in one window;
- workspace tasks can expose `wkit` diagnostics and repo-local entrypoints.

This keeps `wkit` as a thin coordination layer. It does not turn the workspace
into a monorepo, a build graph, or an IDE product.

Official VS Code references:

- [Multi-root Workspaces](https://code.visualstudio.com/docs/editing/workspaces/multi-root-workspaces)
- [Tasks](https://code.visualstudio.com/docs/debugtest/tasks)
- [Working with repositories](https://code.visualstudio.com/docs/sourcecontrol/repos-remotes#_working-with-repositories)
- [Workspace Trust](https://code.visualstudio.com/docs/editing/workspaces/workspace-trust)

## Workflow

Preview the generated target:

```bash
wkit --workspace /path/to/workspace vscode plan
```

Inspect the textual diff:

```bash
wkit --workspace /path/to/workspace vscode diff
```

Write the generated workspace file:

```bash
wkit --workspace /path/to/workspace vscode apply --yes
```

Open it in VS Code:

```bash
wkit --workspace /path/to/workspace vscode open
```

If the file is missing or stale, `open` refuses to write unless explicitly
confirmed:

```bash
wkit --workspace /path/to/workspace vscode open --yes
```

`vscode open` uses the `code` command-line launcher. If `code` is not available
on `PATH`, install it from VS Code with the command palette action:

```text
Shell Command: Install 'code' command in PATH
```

## Generated content

The generated workspace includes:

- a root folder for the `wkit` workspace;
- one root folder for each declared repo binding;
- `wkit` tasks for `overview`, `validate`, `doctor`, and `status`;
- scenario `status` and `run` tasks for pinned scenario locks;
- repo entrypoint tasks from `repos/<repo-id>/repo.yaml`;
- small workspace settings that improve multi-root readability, such as showing
  folder names in editor labels and the repositories view in Source Control.

Repo entrypoint tasks use VS Code scoped variables such as:

```json
"cwd": "${workspaceFolder:app-web}"
```

Entrypoint commands keep the same v0.x command policy as scenario execution:
quoted shell fragments are not parsed. Put shell-sensitive behavior in
repo-local wrapper scripts and reference those scripts from `repo.yaml`.

## Safety behavior

The export follows the same preview-first posture as install targets:

- missing repo bindings block rendering;
- changed existing workspace files are blocked by default;
- `--force` overwrites changed files;
- `--backup` writes a backup before overwriting;
- `--dry-run` previews apply without writing;
- generated output stays under the `wkit` workspace root;
- symlinked target files or parent paths that escape the workspace are blocked.

The export does not write `.vscode/*` files into bound repositories by default.

## Smoke test

From a local checkout:

```bash
make build
WKIT_BIN="$(pwd)/bin/wkit" sh examples/minimal-workspace/run-demo.sh
```

The demo prints a temporary workspace path. Use it to inspect the generated
workspace:

```bash
bin/wkit --workspace <demo-workspace> vscode plan
bin/wkit --workspace <demo-workspace> vscode diff
bin/wkit --workspace <demo-workspace> vscode open
```

Inside VS Code, check:

- roots include `wkit: minimal-workspace`, `app-web`, and `shared-schema`;
- Source Control shows both demo repositories;
- tasks include `wkit: validate`, `wkit: doctor`, `app-web: test`,
  `shared-schema: test`, and `scenario: schema-rollout: run`;
- running tasks produces the same local evidence as the CLI workflow.
