# Research Index

Date: 2026-04-19
Project: Polyrepo Workspace Kit (`wkit`)
Status: research snapshot

## Purpose

This folder is the evidence base for deciding whether `wkit` is a good project idea and how it should be shaped.

The research is split by question, not by source type:

- [market.md](market.md) - why this problem matters now, who has it, and what adoption signals support the timing.
- [competitors.md](competitors.md) - direct and adjacent tools, what they already solve, and where `wkit` must differentiate.
- [agent-standards.md](agent-standards.md) - current fragmentation and convergence around `AGENTS.md`, `SKILL.md`, `CLAUDE.md`, Copilot instructions, OpenCode, and MCP.
- [tech-options.md](tech-options.md) - technical architecture options, data model choices, install safety, scenario reproducibility, and MVP recommendations.
- [cli-stack.md](cli-stack.md) - language and implementation stack choice for the first `wkit` CLI.

Narrow addenda:

- [precedents-github-kiro-sourcegraph.md](precedents-github-kiro-sourcegraph.md) - GitHub Well-Architected, Kiro multi-root workspaces, and Sourcegraph context as positioning precedents.
- [agents-ge-teardown.md](agents-ge-teardown.md) - nearest direct-adjacent teardown for agents.ge.
- [polyrepo-feature-inspiration.md](polyrepo-feature-inspiration.md) - feature-level research from bulk polyrepo Git managers and adjacent tools.
- [empirical-agent-compatibility-matrix.md](empirical-agent-compatibility-matrix.md) - local probe results and remaining compatibility tests for Codex, Claude, Copilot, and OpenCode.
- [primary-research-plan.md](primary-research-plan.md) - user interview, pilot workspace, measured workflow, and cold-start validation plan.

## Executive Verdict

The idea is worth pursuing, but the center of gravity should be precise.

`wkit` should not be positioned as a generic AI instructions generator. That layer is useful, but it is becoming standardized and easier to copy. The durable opportunity is a local multi-repo coordination model that helps humans and agents understand:

- which repositories belong together;
- how those repositories depend on each other;
- what a cross-repo change is trying to accomplish;
- which checks prove the change is safe;
- which generated agent files came from canonical workspace context.

The strongest product thesis:

> A local multi-repo workspace kit for humans and AI agents: map repo relationships, generate portable agent guidance, and validate cross-repo changes without pretending your polyrepo is a monorepo.

## Recommended Structure

The proposed 3-file split was close, but one extra file is useful:

```text
research/
  README.md
  market.md
  competitors.md
  agent-standards.md
  tech-options.md
  cli-stack.md
  precedents-github-kiro-sourcegraph.md
  agents-ge-teardown.md
  polyrepo-feature-inspiration.md
  empirical-agent-compatibility-matrix.md
  primary-research-plan.md
```

Why `agent-standards.md` deserves its own file:

- It is not only a competitor topic.
- It is not only a technical format topic.
- It is the main strategic risk for the project: if agent guidance converges into common standards, simple adapter generation becomes table stakes.
- It also validates the architecture decision in `docs/RFC.md`: canonical guidance should live above tool-specific outputs.

## Core Decision

Build the MVP around `change` + `scenario`.

Adapters should exist, but as supporting features:

- `AGENTS.md`
- `.agents/skills/*`
- `CLAUDE.md`
- `.github/copilot-instructions.md`
- `.github/instructions/*.instructions.md`
- OpenCode-compatible skill locations

The more important capability is the workspace truth behind those files:

- repo registry;
- repo bindings;
- cross-repo relations;
- context sets;
- change objects;
- scenario locks;
- validation reports.

## Confidence

Confidence: medium-high.

Why not higher:

- There is clear demand for multi-repo and AI-agent context, but willingness to maintain yet another manifest must be validated with users.
- Agent standards are moving fast.
- The `scenario` abstraction is promising but must become concrete enough to avoid being just "run shell commands in several repos."

## Next Validation Step

The next real test should not be more docs. It should be a small end-to-end prototype with 3-5 repositories:

1. Register repos.
2. Define relations.
3. Define one cross-repo context.
4. Create one change.
5. Pin one scenario.
6. Generate agent guidance.
7. Run validation.
8. Produce a report.

If users say "I would keep this manifest current because it saves coordination time", the project has a strong path. If they say "this is just another config layer", the project should narrow or pivot.
