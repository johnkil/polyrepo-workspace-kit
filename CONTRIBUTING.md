# Contributing

Thank you for helping improve Polyrepo Workspace Kit.

This project is still in v0.x, so the most useful contributions are small, evidence-backed changes that keep the CLI narrow and trustworthy.

## Development Setup

Requirements:

- Go 1.25 or newer
- Git

From the repository root:

```bash
make tools
make check
```

Optional checks:

```bash
make coverage
make fuzz
```

Use `make fmt` before sending changes.

## Contribution Guidelines

- Keep changes focused and easy to review.
- Prefer repo-local commands and explicit manifests over hidden discovery.
- Do not add new canonical workspace entities without updating the RFC, spec, docs, examples, and tests.
- Keep adapter outputs derived from canonical workspace state.
- Add or update tests for behavior changes.
- Update `CHANGELOG.md` for user-visible behavior, checks, docs, or packaging changes.

## Pull Request Checklist

Before opening a PR:

- Run `make check`.
- Run `make coverage` if tests changed.
- Run `make fuzz` if YAML, id, path, scenario, or report safety changed.
- Update docs when command behavior, install behavior, or compatibility claims change.
- Avoid claiming adapter compatibility without empirical evidence in `research/empirical-agent-compatibility-matrix.md`.

## Scope Discipline

For v0.x, avoid:

- plugin runtimes;
- hosted services;
- universal command abstractions;
- graph auto-discovery as canonical truth;
- tool-specific user-scope installs without compatibility evidence.

## License

By contributing, you agree that your contribution is submitted under the Apache License 2.0, as described in `LICENSE`.
