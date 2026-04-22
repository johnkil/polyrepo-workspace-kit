# GitHub Stacked PRs Teardown

Date: 2026-04-22
Scope: adjacent review-layering workflow for single-repo pull request stacks.

## Question

Does GitHub Stacked PRs strengthen the `wkit` thesis, or does it pull the
project toward a blurrier PR orchestration product?

## Short Answer

GitHub Stacked PRs strengthens the thesis if `wkit` treats it as an adjacent
review workflow, not as a feature model to copy.

`gh-stack` makes a large change reviewable by splitting it into ordered PR
layers inside one repository. `wkit` should make a cross-repo change reviewable
by declaring the involved repositories, their relationships, the active change,
and the local validation evidence.

The clean positioning:

> GitHub Stacked PRs makes single-repo changes reviewable. `wkit` makes
> cross-repo changes reviewable.

Sources:

- [GitHub Stacked PRs overview](https://github.github.com/gh-stack/)
- [Quick Start](https://github.github.com/gh-stack/getting-started/quick-start/)
- [CLI reference](https://github.github.com/gh-stack/reference/cli/)
- [FAQ](https://github.github.com/gh-stack/faq/)

## What GitHub Stacked PRs Claims

GitHub Stacked PRs is a native GitHub workflow, currently in private preview,
for arranging pull requests into an ordered stack. A stack is a chain of PRs in
the same repository where each PR targets the branch below it, ultimately
landing on the trunk branch.

The workflow includes:

- GitHub UI stack navigation from each PR;
- PRs that show focused diffs for each layer;
- branch protection and required checks evaluated against the stack base;
- bottom-up or partial contiguous merges;
- automatic rebase behavior after merge;
- a `gh stack` CLI extension for local stack management;
- an optional `github/gh-stack` skill for AI coding agents.

The CLI surface includes `init`, `add`, `view`, `checkout`, `submit`, `sync`,
`rebase`, `push`, `unstack`, `link`, and navigation commands such as `up`,
`down`, `top`, and `bottom`.

## Where It Is Strong

1. Review ergonomics

It makes a large single-repo change easier to review by splitting it into
focused diffs with visible ordering.

2. Native host integration

Because GitHub owns the PR UI, branch protection behavior, merge queue
integration, and stack merge behavior, this is stronger than a third-party
wrapper for GitHub-only teams.

3. Local workflow assistance

The CLI handles the tedious parts of branch chains: creating layers, pushing,
submitting PRs, cascading rebases, and navigation.

4. Agent workflow signal

The documented agent skill suggests that stack-aware workflows are moving into
agent instructions. This validates `wkit` keeping skills as a supporting
surface, but not as the product center.

## Where It Does Not Close The `wkit` Wedge

Based on the public documentation, GitHub Stacked PRs is intentionally
repository-local:

- a stack is a chain of PRs in one repository;
- all branches in a stack must live in the same repository;
- cross-fork stacks are not supported;
- the model is PR and branch based, not workspace topology based.

Open space for `wkit`:

- local workspace registry;
- repo bindings to multiple checkouts;
- cross-repo relations;
- named contexts;
- cross-repo change manifests;
- scenario locks across repositories;
- repo-local entrypoint execution;
- local reports that show pass, fail, block, skip, and drift across repos;
- derived handoff context for humans and agents.

`gh-stack` answers:

> How should one repository split a large change into reviewable PR layers?

`wkit` should answer:

> Which repositories are involved in this coordinated change, how do they
> relate, and what local evidence proves the set is safe enough to review?

## Product Risk

High-risk directions for `wkit`:

- `wkit stack`;
- PR chain creation;
- branch base management;
- cascading rebase orchestration;
- GitHub Stack API ownership;
- merge queue integration;
- scenario lifecycle inferred from PR, issue, branch, merge, or CI state.

Those directions conflict with the accepted v0.x boundary in which `change`
state is local and declarative, `handoff` does not inspect code-host state, and
`scenario` is local validation evidence rather than CI or PR orchestration.

## Product Response

Use `gh-stack` as category validation:

- reviewability matters;
- large changes benefit from smaller layers;
- reviewers need both focused diffs and the larger context;
- agents need workflow-specific instructions.

Do not copy the product surface.

The useful integration posture is documentation and handoff awareness, not
control-plane ownership:

- mention stacked PRs as compatible with repo-local review workflows;
- let each repo choose its own branch and PR workflow;
- keep `wkit` centered on cross-repo scope and scenario evidence;
- if a future handoff includes PR links, treat them as optional derived context.

## Boundary Rule

During v0.x, a proposed `gh-stack` inspired feature should pass this gate:

> Does this help describe, validate, or hand off a cross-repo change without
> taking ownership of PR lifecycle, branch mutation, or remote GitHub state?

If yes, consider it as derived reporting or documentation. If no, defer it.
