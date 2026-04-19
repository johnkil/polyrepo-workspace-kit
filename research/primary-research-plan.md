# Primary Research Plan

Date: 2026-04-19
Scope: user validation plan for proving whether `wkit` creates real value.

## Question

What evidence is needed beyond web research and competitor analysis?

## Short Answer

The main unresolved risk is not market awareness. It is proof of value.

The next best evidence is primary research:

- 5-10 user conversations;
- 2 independent pilot workspaces;
- 3-5 measured workflows;
- 1 cold-start onboarding test without author help.

## Core Hypotheses

H1. Target users already coordinate changes across multiple repositories often enough to feel pain.

H2. The pain is not just checkout management; it includes context, ownership, rollout order, validation, and handoff.

H3. Users will maintain a small workspace manifest if it produces immediate value.

H4. `change` + `scenario` is more valuable than generated agent files alone.

H5. Agent guidance generated from workspace topology is useful, but only as a supporting feature.

H6. The MVP can stay local and file-based without needing a hosted service.

## Target Participants

Prioritize:

- platform engineers;
- SDK/API maintainers;
- infra/tooling engineers;
- maintainers of related OSS packages;
- teams with service + SDK + docs + examples split across repos;
- developers already using Codex, Claude Code, Copilot, OpenCode, Cursor, or Kiro.

Avoid for the first round:

- single-repo app teams;
- teams fully standardized on one monorepo platform;
- users who only want a generic `AGENTS.md` generator;
- people who cannot discuss real workflows concretely.

## User Conversation Guide

Ask for a recent real workflow, not abstract opinions.

Questions:

1. Tell me about the last change you made that touched more than one repository.
2. Which repos were involved?
3. How did you know those repos were involved?
4. What order did the changes need to happen in?
5. What broke or almost broke?
6. What commands or CI jobs proved the change was safe?
7. Where was the coordination knowledge stored?
8. Did an AI coding tool help? If yes, what context did you have to provide?
9. What would have made the task easier?
10. Would you maintain a small manifest if it generated validation reports and agent guidance?

Watch for:

- concrete pain;
- repeated workflows;
- manual checklists;
- stale docs;
- Slack/Linear/Jira as de facto coordination memory;
- repeated agent prompting;
- fear of agents touching the wrong repo;
- inability to prove cross-repo safety locally.

## Pilot Workspaces

Run two independent pilots:

### Pilot A: API / SDK / Docs

Repos:

- API service;
- shared schema or OpenAPI spec;
- generated SDK;
- docs site;
- example app.

Workflow:

- add/rename one API field;
- update schema;
- update SDK;
- update docs;
- run validation scenario.

### Pilot B: Platform Migration

Repos:

- shared library;
- two consumer apps;
- CI template or repo policy;
- docs or onboarding repo.

Workflow:

- migrate a build/test command or shared package;
- update consumers;
- validate all affected repos.

## Measured Workflows

Measure 3-5 workflows:

1. Baseline without `wkit`

Participant explains how they would discover repos, commands, and validation steps today.

2. Workspace setup

Time how long it takes to register repos, bindings, relations, and one context.

3. Change creation

Create a `change` object for a real or realistic cross-repo task.

4. Scenario validation

Pin and run a scenario. Capture whether the report is useful.

5. Agent handoff

Generate agent guidance and ask an agent to summarize the workspace/change.

## Metrics

Quantitative:

- setup time;
- number of repos correctly identified;
- number of missed repos before/after;
- number of manual prompts needed for an agent;
- time to produce validation evidence;
- number of commands discovered automatically;
- number of stale/missing bindings or entrypoints found.

Qualitative:

- "I would keep this file current" / "I would not";
- "This report would help in review" / "it would not";
- "This reduces re-explaining context to agents" / "it does not";
- "This is lighter than our current approach" / "it is extra process."

## Cold-Start Test

Give a participant:

- a sample workspace;
- a README;
- a task;
- no author help.

Ask them to:

1. understand the workspace;
2. identify affected repos;
3. create or inspect a change;
4. pin/run a scenario;
5. generate agent guidance;
6. explain whether the output is useful.

Success means:

- they can complete the flow without author intervention;
- they understand what `change` and `scenario` mean;
- they can explain why generated agent files are not the canonical source;
- they can name one real workflow where they would use it.

Failure means:

- terminology is confusing;
- manifest maintenance feels unjustified;
- scenario output is not useful;
- generated agent files feel like the only valuable part.

## Decision Criteria

Proceed if:

- at least 5 target users report recurring cross-repo coordination pain;
- at least 3 say they would maintain a manifest for this;
- at least 2 pilot workspaces produce useful scenario reports;
- cold-start user can complete the core workflow;
- agent handoff becomes easier with generated context.

Narrow or pivot if:

- users only want agent instruction generation;
- users see no value in scenario reports;
- setup feels heavier than the pain;
- repo relationships are too fluid to model;
- existing tools already solve the workflow for the target segment.

## Research Artifacts To Capture

For each conversation:

- participant role;
- repo topology;
- recent cross-repo workflow;
- current tools;
- pain score 1-5;
- manifest willingness 1-5;
- scenario report usefulness 1-5;
- agent-context usefulness 1-5;
- objections;
- exact quotes.

For each pilot:

- before workflow;
- `wkit` setup;
- scenario report;
- generated agent files;
- missed assumptions;
- changes recommended to product model.

## Practical Next Step

Do not build a broad platform demo.

Build one narrow prototype flow:

```text
init -> register repos -> bind paths -> define relations -> create change -> pin scenario -> run scenario -> generate AGENTS.md -> produce report
```

Then test that flow with real users.
