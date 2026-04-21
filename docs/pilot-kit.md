# Pilot Kit
## Polyrepo Workspace Kit

Status: proof-stage runbook

This kit is for non-author pilots. It should help a participant run one bounded
polyrepo workflow and leave behind enough evidence to decide whether `wkit`
reduced coordination ambiguity.

## 1. Participant Checklist

Before the pilot:

- choose one repeated polyrepo workflow;
- identify the involved local checkouts;
- install or build `wkit`;
- decide whether local telemetry is acceptable for this run;
- keep one baseline note for how the same workflow is done today.

During the pilot:

- initialize or open a workspace;
- register and bind the involved repos;
- inspect relation candidates with `wkit relations suggest`;
- create a context and change;
- pin and run a scenario;
- generate a handoff summary;
- record surprises, wrong-repo jumps, and search-everywhere episodes.

After the pilot:

- export telemetry if it was enabled;
- collect the evidence bundle below;
- answer the pass/fail questions;
- note whether keeping manifests current felt worth it.

## 2. Suggested Command Flow

```bash
wkit demo failure

wkit init ./workspace \
  --repo app-web=../app-web \
  --repo shared-schema=../shared-schema \
  --repo-kind shared-schema=contract \
  --context schema-rollout \
  --change-title "Payload field rollout"

wkit --workspace ./workspace relations suggest
wkit --workspace ./workspace telemetry enable
wkit --workspace ./workspace scenario pin schema-rollout --change <change-id>
wkit --workspace ./workspace scenario run schema-rollout
wkit --workspace ./workspace handoff <change-id> --scenario schema-rollout
wkit --workspace ./workspace telemetry export > pilot-telemetry.jsonl
```

Telemetry is optional. It is disabled by default, writes only under
`local/telemetry/*`, and exports only when the participant runs
`wkit telemetry export`.

## 3. Run Sheet

Pilot:

- participant:
- date:
- workspace:
- workflow name:
- previous method:
- repos involved:
- coding agent or editor used:
- telemetry enabled: yes / no

Workflow notes:

- start time:
- first successful bounded workflow time:
- end time:
- commands run:
- where the participant hesitated:
- where the participant searched broadly:
- where the participant opened or edited the wrong repo:
- where the scenario report helped:
- where the handoff helped:
- where the model felt too heavy:

Metrics:

- time to first successful cross-repo workflow:
- wrong-repo exploration count:
- search-everywhere episodes:
- install overwrite/conflict events:
- scenario drift events:
- onboarding completed without maintainer help: yes / no

Participant answer:

- Would you keep these manifests current for this repeated workflow?
- What would you delete from the model?
- What would you add only after another repeated workflow proves it?

## 4. Evidence Bundle

Create one local folder per pilot run and include:

```text
pilot-evidence/
  run-sheet.md
  workspace-summary.txt
  change.yaml
  manifest.lock.yaml
  scenario-report.yaml
  scenario-report.md
  handoff.md
  telemetry.jsonl
  notes.md
```

Minimum evidence:

- `coordination/changes/<change-id>.yaml`;
- `coordination/scenarios/<scenario-id>/manifest.lock.yaml`;
- latest `local/reports/<scenario-id>/<run-id>.yaml`;
- latest `local/reports/<scenario-id>/<run-id>.md`;
- `wkit handoff <change-id> --scenario <scenario-id>` output;
- run sheet answers.

Optional evidence:

- `wkit telemetry export` output;
- `wkit status` before and after;
- generated VS Code workspace file;
- install plan/diff/apply output where adapter install was part of the workflow.

## 5. Pass / Fail Criteria

A pilot counts as passed only if:

- the participant completed one bounded cross-repo workflow;
- the scenario lock and report were understandable without maintainer
  translation;
- the participant can explain where canonical truth lives;
- the participant says whether manifest maintenance is worth the coordination
  savings;
- no new canonical entity was required to finish the workflow.

A pilot is inconclusive if:

- the workflow was too artificial;
- the participant relied on maintainer explanation for every step;
- evidence artifacts are missing;
- the workspace never reached scenario pin/run.

A pilot should be treated as failed if:

- the model caused more wrong-repo exploration than the baseline;
- the participant confused generated adapter outputs with canonical state;
- handoff/report artifacts were not useful to another human or agent;
- the user asked for a hosted, CI, graph database, or portal feature as the only
  way to make the workflow useful.
