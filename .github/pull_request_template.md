## Summary

Describe the change and the workflow it improves.

## Checks

- [ ] `make check`
- [ ] `make coverage` if tests changed
- [ ] `make fuzz` if YAML, id, path, scenario, report, install, or backup safety changed
- [ ] Docs updated for user-visible behavior changes
- [ ] `CHANGELOG.md` updated for user-visible behavior, checks, docs, or packaging changes

## Scope

- [ ] The change keeps Polyrepo Workspace Kit a thin local coordination layer.
- [ ] Adapter outputs remain derived from canonical workspace state.
- [ ] New compatibility claims are backed by evidence in `research/empirical-agent-compatibility-matrix.md`.
