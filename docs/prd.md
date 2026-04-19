# Product Requirements Document
## Polyrepo Workspace Kit

Version: 1.6
Status: Active v0.x baseline after final doc hardening pass

## 1. Product sentence

**Polyrepo Workspace Kit is a thin coordination layer for local polyrepo workspaces, helping humans and coding agents coordinate repeated cross-repo work without pretending a polyrepo is a monorepo and without turning tool-specific agent files into the source of truth.**

## 2. Summary

Polyrepo Workspace Kit exists for teams and individual developers who work across multiple repositories locally and need a small, explicit way to coordinate repeated cross-repo workflows.

The product is intentionally layered:

1. **Core workspace** — canonical coordination truth
2. **Portable guidance** — small rules and reusable skills
3. **Adapters** — repo-scope and user-scope installs into real tool discovery locations
4. **Packs** — future distribution layer, explicitly deferred

The product does **not** attempt to be:

- a monorepo manager;
- a developer portal;
- a retrieval engine;
- an AI IDE;
- a universal subagent framework;
- a command abstraction framework;
- a project-memory platform.

## 3. Problem

Teams operating across many repositories repeatedly face the same issues:

- cross-repo changes are real, but the coordination model is implicit;
- repo relationships are known socially rather than structurally;
- rollout order and validation evidence are reconstructed each time;
- broad search is often used where bounded context would be safer;
- agent context is fragmented across tool-specific files and user habits;
- generated agent files are easily mistaken for source of truth;
- large metadata systems create new maintenance burden before proving value.

The product should make these workflows more explicit without demanding monorepo adoption or a heavy internal platform.

## 4. Primary opportunity

The durable opportunity is **not** “generate more agent files.”

The durable opportunity is:

- declare which repositories belong to a workspace;
- declare which repositories matter for a task family or live change;
- make a live cross-repo `change` a first-class object;
- pin a `scenario` to current repository state and derived checks;
- produce a **reviewable local validation snapshot** for a coordinated change;
- derive bounded guidance from the same coordination model.

The strongest wedge is therefore:

**cross-repo coordination and scenario validation in a local multi-repo workspace.**

## 5. Product boundary

The project should be positioned as a **local coordination substrate**, not as a larger category winner in adjacent spaces.

It is deliberately **not**:

- an enterprise SDLC operating framework;
- a code search or retrieval system;
- an IDE-level multi-root UX layer;
- a batch-change or PR-campaign engine;
- a single source of truth for all agent instructions;
- a long-term project-memory system.

## 6. Target users

### First audience

- developers who actively work across **5+ repositories**;
- platform, tooling, infra, SDK, or API maintainers dealing with repeated cross-repo workflows;
- teams whose workflows span service + schema + SDK + docs + examples, or shared library + multiple consumers;
- teams already using one or more coding-agent tools and wanting bounded, inspectable guidance installs;
- teams for whom moving to a monorepo is unrealistic, undesirable, or premature.

### Not first audience

- single-repo teams;
- teams whose main problem is code search rather than coordination;
- organizations primarily looking for a service catalog or developer portal;
- users who only want an `AGENTS.md` generator;
- users who expect universal custom-agent, command, or plugin abstraction on day one.

## 7. Jobs to be done

### Job 1 — Coordinate a real cross-repo change

A user wants to declare which repositories are involved, what the change is trying to accomplish, and what checks prove it is safe.

### Job 2 — Keep repo-local execution authoritative

A user wants build, test, and other operational commands to remain defined inside repositories, not duplicated in a central system.

### Job 3 — Reuse minimal guidance across tools

A user wants one canonical source for small always-on rules and reusable `SKILL.md`-style skills, then wants that guidance installed into supported tool locations.

### Job 4 — Make tool-specific writes inspectable and reversible

A user wants to plan, diff, confirm, force, or back up tool-specific writes rather than letting generated files silently overwrite local state.

### Job 5 — Hand off bounded context to humans and agents

A user wants to pass a change and its affected repositories to another person or agent without re-explaining the workspace from scratch.

## 8. Product goals

### Goal A — Stay a coordination layer

The project must remain a small coordination model, not a platform.

### Goal B — Make cross-repo work explicit

The project must support named contexts, live changes, and pinned scenarios as first-class concepts.

### Goal C — Preserve repo-local truth

Repo-local commands and docs must remain more authoritative than workspace-level abstractions.

### Goal D — Use portability only where convergence is real

Portable guidance should focus on small rules and skills, not on universalizing every agent feature.

### Goal E — Make writes safe by default

All tool-specific writes must be plannable, diffable, and controllable.

### Goal F — Keep public docs truthful

Public documentation must describe shipped behavior or clearly label future behavior as planned, experimental, or unverified.

### Goal G — Prove value before broadening

The project must prove that users will maintain the manifests because they reduce coordination cost, not because they generate more files.

## 9. Non-goals

The project is explicitly not trying to be:

- a monorepo manager;
- a build graph or distributed task execution platform;
- a software catalog or developer portal;
- a retrieval engine or code graph system;
- a general AI platform;
- a universal custom-agent schema;
- a universal command model;
- a plugin marketplace;
- a hosted policy system;
- a second source of truth for repository code, tests, or release logic;
- a global memory or accumulated project-knowledge system.

## 10. Frozen v0.x scope

### In scope

- workspace initialization;
- repo registration and bindings;
- relations and contexts;
- live change objects;
- pinned scenario manifests and reviewable reports;
- scenario execution through repo-local entrypoints;
- canonical rules and skills;
- thin adapters for portable, Codex, OpenCode, Copilot, and Claude;
- safe install UX.

### Explicitly deferred

- universal custom-agent profiles;
- universal command abstractions;
- plugin or pack implementation;
- MCP server or MCP bundle generation;
- ownership catalogs;
- graph auto-discovery as canonical truth;
- runtime hosting or CI platform behavior;
- captured project memory as canonical workspace state.

## 11. Functional requirements

### Core workspace

The product must:

- initialize a canonical local workspace layout;
- register repositories and their kinds;
- maintain local checkout bindings;
- support directional relations with defined semantics;
- support bounded contexts for cross-repo lookup;
- support live cross-repo change objects;
- pin scenarios to current repository state and derived checks;
- run scenarios against local checkouts.

### Validation

The product must:

- validate workspace structure and file shape;
- validate repository ids, references, relations, and bindings;
- validate context, change, and scenario references;
- fail fast on broken references or unusable bindings.

### Portable guidance

The product must:

- support canonical always-on rules;
- support canonical skills built around `SKILL.md`;
- keep the portable surface minimal.

### Adapters

The product must:

- install into real repo-scope discovery locations;
- install into user-scope locations only when explicitly requested;
- support preview, target listing, diffing, confirmation, force, and backup;
- keep adapter outputs clearly marked as generated and non-canonical;
- distinguish the **initial v0.x target surface** from **candidate / unverified** targets in public docs, and record verification status in compatibility notes.

## 12. Quality requirements

- small, explainable model;
- predictable CLI behavior;
- low setup cost;
- idempotent writes when content is unchanged;
- text-first manifests and reports;
- strong YAGNI discipline;
- adapter outputs remain disposable and non-canonical;
- generated guidance stays short, scoped, and conflict-aware;
- scenario claims remain truthful about bounded local review vs full environment reproduction.

## 13. Success criteria

### MVP build criteria

The MVP is built when:

- one example workspace works end-to-end;
- bindings, relations, contexts, changes, and scenarios are all representable and validated;
- scenarios can be pinned and run as **reviewable local validation snapshots** that are useful in practice;
- supported adapters can install canonical guidance into the **initial v0.x target surface** for each non-portable tool adapter in v0.x scope, and portable outputs can be generated and smoke-checked;
- generated writes can be previewed and safely applied.

### MVP proof criteria

The MVP is **proven** only when:

- at least **2 independent workspace pilots** are completed;
- at least **3 repeated workflows** are executed end-to-end;
- at least **1 cold-start onboarding** succeeds without maintainer hand-holding;
- at least **1 compatibility pass is completed and published for each non-portable tool adapter in v0.x scope**, plus **1 portable output smoke test** for `AGENTS.md` and `.agents/skills/*`;
- at least **1 non-author pilot participant reports that keeping the manifests current is worth the coordination savings for repeated workflows**;
- the core remains useful even when adapters are not the primary focus;
- no new canonical entities are added during the proof window.

## 14. Proof metrics

The project should track a small number of proof-oriented metrics:

- time to first successful cross-repo workflow;
- wrong-repo exploration count;
- broad-search episodes during a bounded workflow;
- install overwrite/conflict rate;
- onboarding completion without maintainer help.

## 15. YAGNI gate

No new canonical entity should be added unless all of the following are true:

1. it has a clear writer;
2. it has a clear reader;
3. it has a validator or other enforcement point;
4. it removes repeated workflow pain;
5. the same workflow cannot be explained or reproduced without it.
