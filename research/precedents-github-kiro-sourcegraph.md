# Precedents: GitHub Well-Architected, Kiro, Sourcegraph

Date: 2026-04-19
Scope: narrow addendum for positioning and product framing.

## Question

Do GitHub Well-Architected, Kiro, and Sourcegraph change the `wkit` thesis?

## Short Answer

No. They do not change the thesis, but they make the positioning sharper.

Each precedent validates a different part of the market:

- GitHub Well-Architected validates the need for an operating-model lens, not just tool configuration.
- Kiro validates multi-root workspace UX and per-root agent artifacts.
- Sourcegraph validates context retrieval as a major battleground for coding agents.

Together they make one thing clearer: `wkit` should not claim to be an IDE, retrieval engine, enterprise operating framework, or batch-change platform. Its stronger position is a local coordination substrate for multi-repo changes.

## GitHub Well-Architected

Source: [GitHub Well-Architected](https://wellarchitected.github.com/) and [GitHub Well-Architected overview](https://wellarchitected.github.com/library/overview/).

GitHub Well-Architected is organized around pillars such as productivity, collaboration, application security, governance, and architecture. It is framed as an opinionated guide and assessment model for deploying and operating GitHub effectively at scale.

Why it matters:

- It is an operating-model precedent: teams need repeatable assessments, design principles, and governance language around development platforms.
- It shows that "how teams operate" is a product surface, not only a docs surface.
- It validates the idea that tools can ship checklists, anti-patterns, and decision frameworks without pretending to own every workflow.

Implication for `wkit`:

- `wkit` can use a lightweight readiness/health vocabulary without becoming a GitHub Well-Architected clone.
- Examples: workspace health, stale bindings, missing scenario checks, unsafe user-scope installs, missing repo entrypoints, conflicting generated guidance.
- Avoid: broad enterprise framework language before the local CLI proves value.

Positioning lesson:

> `wkit` is a small operational layer for multi-repo work, not a complete SDLC operating model.

## Kiro Multi-Root Workspaces

Source: [Kiro multi-root workspaces](https://kiro.dev/docs/editor/multi-root-workspaces/).

Kiro supports workspaces with multiple root folders. Its docs describe multi-root behavior for file paths, codebase indexing, repository maps, specs, steering files, hooks, and MCP servers.

Important details:

- Codebase indexing and repository maps include all roots.
- Ambiguous file references show matching paths so the user can choose.
- Specs from each root are shown as one unified list with the containing root displayed.
- Steering files are retrieved from each root's `.kiro` folder.
- Always-included steering files are always loaded; conditional steering files apply only within their root and matching pattern.
- Hooks trigger only for files in the same root where the hook is defined.
- MCP server behavior has explicit multi-root conflict behavior.

Why it matters:

- This is a strong UX precedent for multi-root agent work.
- It proves that modern agentic IDEs are already treating "workspace with several roots" as a first-class scenario.
- It also shows how quickly precedence and scope rules become subtle.

Implication for `wkit`:

- `wkit` should explicitly model root/repo identity rather than rely on folder names.
- Ambiguity should be surfaced, not guessed.
- Generated guidance should preserve repo identity in multi-root contexts.
- Hooks, MCP, and custom agent behavior should remain out of core until scope rules are proven.

Positioning lesson:

> Kiro owns IDE multi-root UX. `wkit` should own portable multi-repo coordination outside one IDE.

## Sourcegraph Context And Retrieval

Sources:

- [Sourcegraph Cody Context](https://sourcegraph.com/docs/cody/core-concepts/context)
- [Sourcegraph Agentic Context Fetching](https://sourcegraph.com/docs/cody/capabilities/agentic-context-fetching)
- [Sourcegraph Batch Changes](https://sourcegraph.com/docs/batch-changes)

Sourcegraph is relevant in two ways:

1. Context retrieval: Cody uses sources such as keyword search, Sourcegraph Search, and code graph context. Cody supports repo-based context, including multiple repositories in supported clients.
2. Cross-repo changes: Batch Changes automates large-scale code changes across many repositories and code hosts.

Why it matters:

- Sourcegraph validates that context quality is a central AI-coding problem.
- It validates multi-repo context as a serious enterprise need.
- It also validates that cross-repo change orchestration is already a category.

Implication for `wkit`:

- `wkit` should not compete as a retrieval engine.
- It should not claim to replace code search, embeddings, or code graph systems.
- It can complement retrieval by declaring the intended workspace scope, relationships, active change, and validation scenario.

Positioning lesson:

> Sourcegraph retrieves context. `wkit` declares coordination context.

## Combined Product Implication

These precedents suggest a clean product boundary:

- GitHub Well-Architected: broad operating model.
- Kiro: IDE-level multi-root agent UX.
- Sourcegraph: codebase retrieval and cross-repo change platform.
- `wkit`: local, file-based coordination model for polyrepo changes.

The sharper `wkit` wedge:

- not "index all code";
- not "run all PR campaigns";
- not "be the agent IDE";
- not "be the enterprise SDLC framework";
- yes: "tell humans and agents which repos matter, why they matter, and how this cross-repo change was validated."
