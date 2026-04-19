# agents.ge Teardown

Date: 2026-04-19
Scope: nearest direct-adjacent competitor for the agent-guidance story.

## Question

How dangerous is agents.ge to the `wkit` positioning, and where is `wkit` still meaningfully different?

## Short Answer

agents.ge is the closest direct-adjacent project found so far. It is a serious positioning risk if `wkit` is framed as "one source of truth for agent instructions."

But it does not appear to close the strongest `wkit` wedge: explicit polyrepo topology, cross-repo `change` objects, pinned `scenario` validation, and repo-local execution evidence.

Source: [agents.ge](https://agents.ge/).

## What agents.ge Claims

agents.ge positions itself around making a project agent-ready through a versioned `.agents/` directory.

Its stated model:

- `AGENTS.md` is the entrypoint.
- `.agents/` is the durable source of truth.
- `.agents/config.yaml` stores stack/capabilities/metadata.
- `.agents/rules/` stores mandatory agent instructions.
- `.agents/knowledge/` stores accumulated project intelligence.
- `.agents/skills/` stores reusable workflows.
- `.agents/mcp/` stores MCP definitions.
- Generated/synced files support Claude Code, Cursor, Codex, GitHub Copilot, Gemini CLI, OpenCode, and other markdown-reading agents.

Its strongest narrative:

> Static instructions rot; project memory should compound over sessions.

That is a good narrative and very close to the general "agent guidance source of truth" story.

## Where agents.ge Is Stronger

1. Clear agent-readiness story

agents.ge has a very legible pitch: one `.agents` directory makes a repository agent-ready.

2. Project memory

It focuses on accumulated knowledge, architecture decisions, patterns, lessons, conventions, and dependency notes.

3. Capture loop

It claims capture hooks that extract decisions and lessons after sessions, with review before entering the knowledge base.

4. MCP sync

It explicitly includes MCP config sync as part of the product story.

5. Low-friction onboarding

The advertised flow is simple:

- run init;
- agent reads AGENTS.md;
- project knowledge fills in over time.

6. Strong positioning against file drift

It names a real pain: `CLAUDE.md`, `.cursorrules`, `GEMINI.md`, Copilot instructions, and `AGENTS.md` drifting apart.

## Where agents.ge Does Not Close The `wkit` Wedge

Based on public positioning, agents.ge appears repo-centric or project-centric, not polyrepo-change-centric.

Open space for `wkit`:

- logical repo registry;
- local checkout bindings;
- cross-repo relations;
- named cross-repo contexts;
- active cross-repo change objects;
- scenario lock manifests;
- repo-local entrypoints;
- validation reports across repos;
- cross-repo rollout order;
- "which repos prove this change safe?"

agents.ge answers:

> How does this project remember knowledge for future agents?

`wkit` should answer:

> Which repositories are involved in this change, how do they relate, and what evidence proves the coordinated change is safe?

These are different enough if `wkit` keeps its center of gravity.

## Positioning Risk

High risk phrases for `wkit`:

- "one source of truth for agent instructions";
- "make any project agent-ready";
- "sync rules to every AI coding agent";
- "project memory for agents";
- "MCP config sync";
- "AGENTS.md plus skills."

Those phrases put `wkit` too close to agents.ge.

Safer phrases:

- "multi-repo workspace coordination";
- "cross-repo change validation";
- "polyrepo scenario locks";
- "repo relationships for humans and agents";
- "agent guidance generated from workspace topology";
- "validation evidence across repos."

## Competitive Response

| agents.ge strength | `wkit` response |
| --- | --- |
| Durable `.agents` directory | Treat `.agents` as an adapter target or optional interop layer, not the core |
| Project memory | Keep memory out of v0 core unless tied to change/scenario evidence |
| MCP sync | Defer MCP to packs/adapters; do not lead with it |
| Multi-agent support | Support adapters, but anchor in workspace semantics |
| Knowledge capture | Consider later as scenario/report capture, not general chat memory |
| Simple onboarding | `wkit` needs an equally concrete first-run flow around one cross-repo change |

## Product Boundary

Do not try to out-agents.ge agents.ge.

Instead:

- generate `AGENTS.md` and skills as outputs;
- optionally interoperate with `.agents/skills` later;
- keep canonical truth in `coordination/` and `guidance/`;
- make `change` and `scenario` the hero workflow.

## Next Teardown Step

This is a desk teardown from public material. A deeper teardown should install agents.ge in a disposable repository and compare:

- generated file tree;
- overwrite behavior;
- adapter targets;
- MCP sync behavior;
- how knowledge capture works;
- whether multi-repo workspaces are modeled or only multiple independent projects.

That hands-on teardown belongs after the `wkit` MVP shape is clearer, because otherwise it can pull the project toward agent-memory features too early.
