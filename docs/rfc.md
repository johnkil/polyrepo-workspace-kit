# Request for Comments
## RFC-0001: core model and layering for Polyrepo Workspace Kit

Status: Accepted for v0.x baseline
Authors: Project maintainers
Last updated: 2026-04-19

## 1. Context

Polyrepo teams need a minimal coordination model for local multi-repo work. At the same time, many teams now use coding agents across several ecosystems.

Audit and research work sharpened the boundary:

- the strongest product wedge is **local multi-repo coordination**, not generic agent-file generation;
- there is real convergence around **small instruction files, skills, and tool access**, but meaningful fragmentation around **custom agents, commands, plugins, precedence rules, and discovery behavior**;
- adjacent tools validate the category, but they do not eliminate the need for a local coordination substrate.

This RFC defines the accepted v0.x layering and the boundaries that should remain frozen during the proof stage.

## 2. Decision

The project adopts four layers:

1. **Core workspace**
2. **Portable guidance**
3. **Adapters**
4. **Packs** (future, deferred)

Only the first layer is the true canonical coordination model. The second layer is canonical content but not the coordination model. The third and fourth layers are derived or optional.

## 3. Accepted thesis

Polyrepo Workspace Kit should be treated as:

- a local, file-based coordination substrate;
- for repeated cross-repo workflows;
- that can derive bounded guidance for supported coding-agent tools;
- without pretending adapter outputs are the source of truth.

The project should **not** be treated as:

- a universal agent abstraction framework;
- a project memory platform;
- a retrieval engine;
- an IDE product;
- a batch-change or PR-campaign engine;
- a hosted operating framework.

## 4. Layering rationale

### 4.1 Core workspace

The core exists to answer:

- which repositories belong to this workspace?
- how are they related?
- which repositories matter for this task or live change?
- what evidence do we need to consider this coordinated change safe?

### 4.2 Portable guidance

Portable guidance exists because there is enough convergence around:

- small project instructions;
- reusable `SKILL.md`-style workflows;
- minimal repo-scope installs.

Portable guidance remains narrow by design.

### 4.3 Adapters

Adapters translate canonical guidance into real discovery scopes for specific tools.

Adapters are **versioned compatibility layers**, not timeless truth.

### 4.4 Packs

Packs may later become a distribution format for reusable installs or tool bundles, but they are deferred until the coordination model is proven.

## 5. Canonical state model

### Canonical shared state

- `coordination/workspace.yaml`
- `coordination/contexts.yaml`
- `coordination/changes/*`
- `coordination/scenarios/*/manifest.lock.yaml`
- `coordination/rules/*`
- `repos/*/repo.yaml`
- `guidance/rules/*`
- `guidance/skills/*`

### Canonical machine-local state

- `local/bindings.yaml`

`binding` is canonical as a logical entity, but its storage is machine-local by design because checkout paths are local facts.

### Coordination rules vs guidance rules

- `coordination/rules/*` are canonical coordination constraints, such as rollout order or validation policy.
- `guidance/rules/*` are canonical portable agent guidance, such as short always-on instructions.
- `workspace.yaml rules:` references coordination rule ids, not guidance rule files.

### Derived state

Files produced from canonical state:

- scenario run reports;
- adapter outputs;
- copied or linked skills in tool directories;
- generated instruction files.

Derived state is disposable and must never become source of truth.

## 6. Why `scenario` stays in core

`change` describes the live cross-repo unit of work.
`scenario` captures a **reviewable local validation snapshot** for that work.

The project keeps `scenario` in core because it is the strongest proof-of-value wedge. Without it, the project risks collapsing into “one more agent-file generator.”

For v0.x, `scenario` is **not** a CI platform and does **not** claim full environment replay. It optimizes for:

- pinned revisions;
- bounded worktree observation;
- normalized local checks;
- drift detection;
- reviewable reports.

It may later grow stronger lock/report details, but v0.x wording should stay honest about current guarantees.

In v0.x, `env_profile` is descriptive metadata only. It may be carried through locks and reports, but it does not imply automatic environment loading, shell activation, secret loading, or toolchain management.

## 7. Why `binding` stays in core

Without `binding`, logical repository ids cannot be resolved to real local checkouts. This breaks validation, scenario execution, and adapter installation.

`binding` therefore remains canonical as a concept, but its storage is explicitly local.

## 8. Why portable guidance stays narrow

Portable guidance exists because there is enough ecosystem convergence around:

- file-based instructions;
- `SKILL.md`-style reusable skills;
- repo-scope and user-scope installs.

Portable guidance does **not** include tool-specific files such as:

- `CLAUDE.md`;
- `.github/copilot-instructions.md`;
- `.opencode/*`;
- `.codex/*`.

Those belong to adapters.

## 9. Why adapters are not canonical

Adapter outputs vary by:

- file paths;
- precedence rules;
- repo-scope vs user-scope behavior;
- support for commands, skills, or custom agents.

Therefore adapters are treated as compatibility layers, not as product truth.

For v0.x, public docs should distinguish between:

- **initial v0.x target surface** — the supported paths the project intends to cover in v0.x;
- **verification status** — whether a target is docs-backed, empirically verified, or still candidate / unverified;
- **candidate / unverified targets** — mentioned only as future or experimental compatibility surfaces.

The project may maintain compatibility notes and empirical probes, but it does not elevate adapter files into canonical state.

For v0.x, the only user-scope path in the initial target surface is portable `.agents/skills/*`. User-scope targets for Codex, OpenCode, Copilot, and Claude remain candidate / unverified unless empirical compatibility passes validate them.

## 10. Relation semantics

Relations are directional. For v0.x:

- `from` is the **consumer / depender / initiator** side;
- `to` is the **provider / dependency / upstream** side;
- relation kinds are a constrained vocabulary, not free text.

Relations support:

- bounded context expansion;
- rollout reasoning;
- scenario scoping;
- validation warnings.

Relations do not imply full graph traversal, ownership inference, or automatic truth discovery.

In v0.x, coordination rules are advisory constraints for validation and rollout reasoning. They do not mutate repository state or execute commands by themselves.

## 11. What remains out of core

The following are out of core for v0.x unless pilot evidence proves otherwise:

- universal custom-agent schema;
- universal command model;
- plugin registry or marketplace;
- pack-first architecture;
- graph auto-discovery as canonical truth;
- ownership catalog;
- hosted runtime or policy layer;
- generic project memory as canonical workspace state.

## 12. Consequences

### Positive consequences

- the core remains small enough to understand and validate;
- the strongest differentiation (`change` + `scenario`) stays in the center;
- adapter churn does not destabilize canonical state;
- YAGNI is enforceable.

### Negative consequences

- some tool-specific convenience features stay deferred;
- compatibility notes and empirical probes become important maintenance work;
- v0.x language must stay disciplined and avoid stronger claims than the schema supports.

## 13. Status

This RFC is accepted for the v0.x proof stage.

The current task is **not** to redesign the model. The current task is to harden truthfulness, complete pilot evidence, and validate the coordination wedge in real usage.
