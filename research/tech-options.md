# Technical Options Research

Date: 2026-04-19
Scope: architecture, file formats, data model, scenario reproducibility, adapter safety, and MVP implementation choices.

## Question

What is the most technically reasonable way to build `wkit` without turning it into a monorepo tool, build system, or generic agent platform?

## Short Answer

Use a small file-based workspace model, keep repo-local commands authoritative, and make generated agent files derived artifacts. Build the MVP around `change` + `scenario`, because that is the part with the strongest differentiated value.

## Existing Local Baseline

The current docs already point in a sound direction:

- `docs/RFC.md` defines four layers: core workspace, portable guidance, adapters, and future packs.
- `docs/SPEC.md` defines the canonical entities: workspace, repo, relation, rule, context, change, scenario, binding, and entrypoint.
- `docs/SPEC.md` uses YAML examples for `workspace.yaml`, `contexts.yaml`, change files, and scenario lock manifests.
- `docs/PLAN.md` correctly defers packs, MCP, subagents, and commands.

The main technical task is not inventing a larger model. It is tightening the small model enough that it can be validated and trusted.

## Architecture Recommendation

Use these layers:

1. Core workspace

Canonical, validated, agent-independent:

- workspace;
- repo;
- relation;
- context;
- rule;
- change;
- scenario;
- binding;
- entrypoint.

2. Portable guidance

Canonical enough to share:

- short always-on rules;
- `SKILL.md` skill directories;
- references and scripts used by skills.

3. Adapters

Generated, non-canonical outputs:

- `AGENTS.md`;
- `.agents/skills/*`;
- `CLAUDE.md`;
- `.claude/skills/*`;
- `.github/copilot-instructions.md`;
- `.github/instructions/*.instructions.md`;
- OpenCode-compatible skill directories.

4. Future packs

Deferred:

- tool-native custom agents;
- subagents;
- commands;
- MCP configs;
- reusable bundles.

## File Format Options

### YAML

Pros:

- already used in local docs;
- familiar for manifests;
- good for hand-written files;
- common in developer tooling.

Cons:

- complex specification;
- indentation and implicit typing pitfalls;
- parser differences can surprise users.

Source: [YAML 1.2.2 specification](https://spec.yaml.io/main/spec/1.2.2/).

Recommendation: use YAML for human-authored workspace files, but keep the schema conservative. Avoid clever YAML features.

### TOML

Pros:

- good for configuration;
- less ambiguous than YAML;
- easier to read for simple key/value structures.

Cons:

- less natural for nested repo/context/change structures;
- current docs already use YAML.

Source: [TOML v1.1.0](https://toml.io/en/v1.1.0).

Recommendation: do not switch now. Consider TOML only if YAML complexity becomes a real user complaint.

### JSON

Pros:

- excellent for machine-generated lockfiles and reports;
- strong compatibility with JSON Schema;
- deterministic formatting is easier.

Cons:

- less pleasant for hand-authored manifests;
- comments are not standard JSON.

Source: [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12).

Recommendation: use JSON Schema to define/validate the data model. YAML files can be parsed into data and validated against equivalent schemas. Reports may be JSON plus Markdown summaries.

## Data Model Recommendations

Keep the core model small:

- `workspace`: coordination boundary.
- `repo`: logical repository identity.
- `binding`: local checkout path for a repo.
- `relation`: conservative connection between repos.
- `context`: named subset of repos for a task family.
- `change`: live cross-repo operation.
- `scenario`: pinned validation snapshot.
- `entrypoint`: repo-local executable truth such as test/build/lint.
- `rule`: coordination constraint or policy.

Do not add these to core in v0:

- universal subagent schema;
- universal command schema;
- hosted policy model;
- PR orchestration model;
- CI provider abstraction;
- MCP server model;
- catalog ownership model.

## Scenario Reproducibility

This is the most important technical area to strengthen.

Current baseline from `docs/SPEC.md`:

- `scenario pin` resolves a change;
- reads current Git `HEAD` for each repo;
- derives checks from repo entrypoints;
- writes a lock manifest;
- `scenario run` loads the manifest and runs checks.

That is a good start, but not reproducible enough.

Recommended `scenario` state:

- repo id;
- absolute or workspace-relative binding path;
- current branch;
- `HEAD` commit;
- dirty tracked file list;
- untracked file list, optionally ignored by policy;
- submodule state if present;
- lockfile hashes for known package managers;
- toolchain hints if available, such as Node, Python, Go, Rust, Java, Swift versions;
- environment variables required by entrypoints, without recording secrets;
- command list;
- command working directory;
- command timeout;
- expected output artifacts;
- timestamp;
- `wkit` version;
- schema version.

Recommended behavior:

- default `scenario pin` should warn on dirty working trees;
- `--allow-dirty` should be explicit;
- scenario lock should be stable and diffable;
- scenario run should report skipped, failed, passed, and blocked states separately;
- validation reports should be human-readable Markdown and optionally machine-readable JSON.

## Install Safety

The current install model is good:

- `plan`;
- `diff`;
- `show-targets`;
- `apply`;
- `--dry-run`;
- `--yes`;
- `--force`;
- `--backup`.

This is necessary because adapters may write into:

- repo roots;
- `.github/`;
- `.claude/`;
- `.opencode/`;
- `.agents/`;
- user-level agent directories.

Recommended additional rules:

- Every generated file should include an ownership marker.
- Never overwrite an unmarked file without `--force`.
- If overwriting with `--force`, offer `--backup`.
- Store generated-file checksums in a local manifest so `wkit` can detect drift.
- Treat user-scope writes as high-risk and require exact target listing.
- Never write secrets into agent guidance.

## CLI Shape

MVP commands should stay close to the current docs:

```text
wkit init <path>
wkit repo register <repo-id> --kind <kind>
wkit bind set <repo-id> <path>
wkit validate
wkit change new <context> --title <title>
wkit change show <change-id>
wkit scenario pin <scenario-id> --change <change-id>
wkit scenario run <scenario-id>
wkit install plan <tool> [repo-id]
wkit install diff <tool> [repo-id]
wkit install show-targets <tool> [repo-id]
wkit install apply <tool> [repo-id]
```

Avoid adding in v0:

- `wkit agent create`;
- `wkit mcp serve`;
- `wkit pr create`;
- `wkit ci watch`;
- `wkit catalog sync`;
- `wkit command generate`.

These may become useful later, but they blur the category too early.

## Adapter Options

Recommended v0 adapter priority:

1. Portable `AGENTS.md`.
2. Portable `.agents/skills/*`.
3. Codex-compatible project guidance and skills.
4. Claude `CLAUDE.md` that imports portable guidance where possible.
5. Copilot `.github/copilot-instructions.md`.
6. OpenCode skill locations.

Defer:

- custom agents;
- subagents;
- commands;
- MCP config;
- marketplace/packs.

Reasoning: instructions and skills have enough convergence. Custom agents, subagents, commands, permissions, and hooks differ too much between tools.

## MCP Option

MCP is strategically important, but should be later.

Potential future MCP server resources/tools:

- list workspaces;
- list repos;
- show repo relations;
- show active change;
- show scenario state;
- run safe read-only validation queries;
- expose generated research/reports.

Why defer:

- MCP adds runtime complexity;
- security and permissions become more serious;
- many users can get value from files and CLI first;
- the MVP needs to validate the core model before serving it over a protocol.

Source: [Model Context Protocol architecture](https://modelcontextprotocol.io/docs/learn/architecture).

## MVP Technical Slice

Build the thinnest useful loop:

1. File parser and validator.
2. Repo registry and bindings.
3. Context definitions.
4. Change creation.
5. Scenario pinning with Git state.
6. Scenario run with repo-local entrypoints.
7. Markdown validation report.
8. `AGENTS.md` generation.
9. One skill output path.
10. Safe install plan/diff/apply.

This would prove the product's real thesis without building a platform.

## Technical Conclusion

The current architecture is reasonable because it is small and layered.

The main improvement is to make `scenario` more exact. If `scenario` becomes a reliable, inspectable record of cross-repo validation, `wkit` has a strong technical reason to exist. If `scenario` remains just a command runner, the product collapses toward existing tools.
