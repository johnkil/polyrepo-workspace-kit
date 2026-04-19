# Proof and Pilot Plan
## Polyrepo Workspace Kit

Version: 1.7
Status: Proof-oriented plan after final doc hardening pass

## 1. Plan philosophy

The project no longer needs broad redesign.
It needs proof.

The current plan therefore prioritizes:

- keeping the baseline frozen;
- fixing doc/spec truthfulness gaps;
- running repeatable pilots;
- collecting evidence;
- validating adapter assumptions empirically;
- resisting pressure to broaden the model too early.

## 2. Hardening pass completed

The focused doc/spec hardening pass is complete enough to enter prototype and pilot preparation. The items below should now be treated as release checks, not as a separate redesign phase.

### Release checks

1. make public docs truthful about shipped vs planned surfaces;
2. align all public file names and links to lower-case docs;
3. define canonical machine-local binding storage;
4. strengthen `scenario` lock and report schema;
5. clean up portable-vs-adapter boundaries in spec;
6. specify install safety tightly enough to implement;
7. distinguish the initial v0.x target surface from candidate adapter targets and record verification status.

Completion criteria:

- public docs do not promise missing binaries or package surfaces;
- `binding`, `scenario`, `relation`, and install safety are fully specified for v0.x;
- lower-case doc names are consistent everywhere;
- supported adapter targets are marked as part of the initial v0.x target surface, or as candidate / unverified.

## 3. What is already considered built enough for the proof stage

The proof stage assumes the following baseline already exists or is close enough to treat as stable:

- core workspace model;
- changes and scenarios;
- portable rules and skills;
- thin adapters;
- safe install UX.

The plan is not primarily about adding more surfaces. It is about validating that the existing surfaces are useful.

## 4. Workstreams

### Workstream A — Freeze the baseline

Goals:

- keep core nouns unchanged through the proof window;
- keep packs and universal custom-agent models deferred;
- keep adapter outputs non-canonical.

Exit criteria:

- no new canonical entity is added during the pilot window;
- forever-deferred list remains intact.

### Workstream B — Pilot the coordination model

Goals:

- validate that `change` and `scenario` are worth maintaining;
- prove that the workspace model reduces coordination ambiguity.

Exit criteria:

- at least 2 independent workspace pilots;
- at least 3 repeated workflows executed end-to-end.

### Workstream C — Validate install ergonomics

Goals:

- validate repo-scope install flows for every initial v0.x target for each non-portable tool adapter;
- validate the portable user-scope install flow;
- keep tool-specific user-scope installs candidate / unverified until empirical passes justify them;
- confirm that plan/diff/apply/force/backup are sufficient for real usage.

Exit criteria:

- users can safely preview and apply initial-surface installs without maintainer intervention;
- the portable user-scope install flow is understandable and predictable;
- overwrite behavior is predictable in real workflows.

### Workstream D — Validate cold-start understanding

Goals:

- prove that the project can be adopted without author context;
- identify where documentation still assumes insider knowledge.

Exit criteria:

- at least one cold-start onboarding succeeds;
- blockers are documented and categorized.

### Workstream E — Validate adapter compatibility empirically

Goals:

- confirm actual discovery behavior for repo-scope installs and any explicitly evaluated user-scope targets;
- document tool/version-specific compatibility assumptions for each non-portable tool adapter in v0.x scope;
- run a separate portable output smoke test for `AGENTS.md` and `.agents/skills/*`;
- avoid over-generalizing from documentation alone.

Exit criteria:

- one compatibility pass completed and published for each non-portable tool adapter in v0.x scope;
- one portable output smoke test completed for `AGENTS.md` and `.agents/skills/*`;
- compatibility matrix updated with probe dates, tool versions, target paths, and confidence levels;
- candidate targets either move to empirically verified with evidence or stay explicitly unverified.

## 5. Pilot matrix

### Pilot A — API / SDK / Docs

Representative workflow:

- change an API or schema field;
- update the API/service repo;
- update generated SDK or client repo;
- update docs and example repos;
- pin and run a scenario;
- review the scenario report.

Success signal:

- the scenario artifact is useful for coordination and review;
- users prefer the bounded workflow over ad hoc search and memory.

### Pilot B — Shared library change

Representative workflow:

- update a shared library used by several repos;
- scope the relevant repos with a `context`;
- create a `change` for the migration;
- run a pinned validation flow.

Success signal:

- the context and scenario reduce wrong-repo exploration;
- the report is useful for handoff or review.

### Pilot C — Onboarding

Representative workflow:

- a new user reads the docs;
- initializes or opens a workspace;
- understands where truth lives;
- runs one scenario-driven workflow.

Success signal:

- the user completes the flow without author intervention;
- the user does not mistake generated adapter files for canonical state.

## 6. Measured workflows

The proof stage should capture 3–5 concrete workflows:

1. baseline workflow without Polyrepo Workspace Kit;
2. workspace setup;
3. change creation;
4. scenario pin and run;
5. agent handoff using generated guidance.

For each workflow, capture:

- previous method;
- steps taken;
- time spent;
- mistakes or ambiguity;
- whether the manifest felt worth maintaining.

## 7. Metrics

Track a small proof-oriented set:

- time to first successful cross-repo workflow;
- wrong-repo exploration count;
- search-everywhere episodes;
- install overwrite/conflict rate;
- scenario drift frequency;
- onboarding completion without maintainer help.

## 8. Evidence artifacts

For each pilot or workflow, capture:

- workspace topology summary;
- workflow summary;
- `change` artifact;
- `scenario` lock;
- scenario report;
- generated adapter outputs where relevant;
- participant feedback;
- what stayed frozen vs what users asked to broaden.

## 9. Stop conditions

Stop adding new core nouns if any of the following are true:

- pilots have not yet validated the existing core;
- the request is primarily for convenience, not repeated workflow pain;
- the same problem can be solved by a better adapter, better docs, or better reports;
- the addition would turn portable guidance into a tool abstraction layer.

## 10. Acceptance gates for “MVP proven”

The MVP should be called proven only if all of the following are true:

- 2 independent pilots completed;
- 3–5 measured workflows captured;
- 1 cold-start onboarding succeeds;
- 1 compatibility pass completed and published for **each non-portable tool adapter in v0.x scope**, plus **1 portable output smoke test** for `AGENTS.md` and `.agents/skills/*`;
- at least 1 non-author pilot participant reports that keeping the manifests current is worth the coordination savings for repeated workflows;
- the core workflow remains useful even when adapter install is not the primary user value;
- no new canonical entities added during the proof window.

## 11. What not to do during this plan

Do **not**:

- broaden core into universal agent abstractions;
- add packs or plugins before pilots justify them;
- add retrieval features as if this were a search product;
- turn scenario into a CI platform;
- let public docs run ahead of shipped behavior.

## 12. Next release discipline

Each public release during the proof stage should include:

- truthful README/install instructions;
- updated compatibility notes;
- migration notes for spec changes;
- one example workspace or scenario artifact;
- clear changelog entries describing what changed in core vs adapters.
