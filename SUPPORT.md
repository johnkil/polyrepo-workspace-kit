# Support

Polyrepo Workspace Kit is currently in a v0.x source-first phase.

## Getting Help

Start with:

- `README.md` for the product overview and CLI contract;
- `docs/install.md` for local development;
- `docs/spec.md` for behavior details;
- `examples/minimal-workspace/README.md` for a runnable example.

Useful local commands:

```bash
make tools
make check
go run ./cmd/wkit --help
```

## Support Channel

Use GitHub Issues for public support requests and bug reports:

https://github.com/johnkil/polyrepo-workspace-kit/issues

Search existing issues before opening a new one.

## What To Include

For bug reports or support requests, include:

- `wkit` command and flags;
- operating system;
- Go version if building from source;
- relevant workspace manifest snippets;
- expected and actual behavior;
- whether the issue affects canonical state, local state, or derived adapter outputs.

Do not include secrets, private repository contents, or unpublished vulnerability details in public support issues.
