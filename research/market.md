# Market Research

Date: 2026-04-19
Scope: AI coding-agent adoption, multi-repo pain, ICP, positioning, and validation criteria.

## Question

Is now a good time to build a local polyrepo coordination tool for humans and AI coding agents?

## Short Answer

Yes, timing is favorable. AI-assisted development is mainstreaming, but trust, accuracy, security, and workflow integration remain unresolved. That creates demand for explicit, inspectable project context.

The market gap is not "developers need another config file." The market gap is that agents are increasingly asked to perform larger changes, while many real organizations still operate across multiple repositories with weak shared context.

## Adoption Signals

AI coding assistance is no longer niche:

- Gartner predicted that by 2028, 75% of enterprise software engineers will use AI code assistants, up from less than 10% in early 2023. Gartner also reported that 63% of surveyed organizations were piloting, deploying, or had already deployed AI code assistants in Q3 2023. Source: [Gartner](https://www.gartner.com/en/newsroom/press-releases/2024-04-11-gartner-says-75-percent-of-enterprise-software-engineers-will-use-ai-code-assistants-by-2028).
- Google DORA 2025 surveyed nearly 5,000 technology professionals and reported 90% AI adoption at work, with a median of two hours per day spent working with AI. It also reported a trust paradox: only 24% said they trust AI a lot or a great deal, while 30% trust it a little or not at all. Source: [Google DORA 2025](https://blog.google/innovation-and-ai/technology/developers-tools/dora-report-2025/).
- Stack Overflow's 2025 survey reports strong usage of out-of-the-box AI assistants among developers using or developing agents: ChatGPT at 81.7%, GitHub Copilot at 67.9%, Google Gemini at 47.4%, and Claude Code at 40.8%. The same survey reports major concerns around agent accuracy, security, and privacy. Source: [Stack Overflow Developer Survey 2025: AI](https://survey.stackoverflow.co/2025/ai).
- GitHub Octoverse 2025 shows AI-related development changing language and repository trends, including strong Python, TypeScript, and JavaScript activity around AI projects. Source: [GitHub Octoverse 2025](https://github.blog/news-insights/octoverse/octoverse-a-new-developer-joins-github-every-second-as-ai-leads-typescript-to-1/).

Interpretation for `wkit`: adoption creates a broad tailwind, but the trust gap matters more. A useful product should make agent behavior more grounded, scoped, and auditable.

## The Polyrepo Pain

The relevant pain is not simply "many Git repos exist." The pain is cross-repo coordination:

- A backend API changes and SDKs, docs, examples, tests, and clients must move together.
- A schema repo drives generated code in several application repos.
- A platform team owns shared policies, but repo teams own local execution.
- A migration affects dozens of repos, but each repo has different test commands and ownership.
- A human or agent needs to know "what else must I check before this change is safe?"

This pain survives even when teams have strong tooling:

- Monorepo tools help when the codebase is actually one coordinated graph.
- Developer portals help with ownership and discovery, but often do not operate as local change workflows.
- Batch-change tools help apply broad mechanical diffs, but not necessarily explain a local multi-repo workspace to an agent.

## Target Users

Best initial ICP:

- platform or infrastructure engineers coordinating related repositories;
- SDK/API teams with service, schema, generated client, examples, and docs split across repos;
- engineering leads maintaining several related OSS packages;
- internal tools teams maintaining many repo templates;
- AI-forward developers who already use Codex, Claude Code, Copilot, OpenCode, Cursor, or similar tools.

Weak initial ICP:

- single-repo application teams;
- teams already fully committed to one monorepo build platform;
- teams that want a hosted enterprise catalog before a local CLI;
- teams that only want one generated `AGENTS.md` file.

## Job To Be Done

When I need to make or review a change that spans multiple repositories, help me understand the relevant repo set, the dependency relationships, the right repo-local commands, and the evidence that the whole change is safe.

Agent-specific version:

When I ask an AI coding agent to work in one repo, make sure it can see the relevant cross-repo context without giving it an unbounded workspace or stale tribal knowledge.

## Why Existing Behavior Is Not Enough

Current agent workflows are usually repo-local. They can read the files in front of them, but cross-repo knowledge is often implicit:

- local folder layout;
- human memory;
- Slack threads;
- issue descriptions;
- README links;
- CI failures after the fact;
- manually pasted context.

This is workable for one-off tasks, but brittle for repeated cross-repo changes.

`wkit` can be valuable if it turns this into explicit context:

- which repos are in scope;
- which repos are related;
- which docs should be read first;
- which checks matter;
- which scenario state was validated;
- which generated agent instructions are current.

## Positioning

Recommended positioning:

> A local multi-repo workspace kit for humans and AI agents: map repo relationships, generate portable agent guidance, and validate cross-repo changes without pretending your polyrepo is a monorepo.

Important phrases for discoverability:

- multi-repo workspace
- polyrepo coordination
- cross-repo changes
- repo relationships
- AI coding agent instructions
- AGENTS.md generator
- SKILL.md generator
- workspace manifest
- scenario validation
- cross-repo validation

Avoid leading with:

- "AI config generator"
- "monorepo alternative"
- "developer portal"
- "MCP platform"
- "universal agent schema"

Those categories either undersell the project or put it against much heavier incumbents.

## Market Risks

1. Maintenance burden

Users may not want another manifest unless the value is immediate.

Mitigation: first-run experience must create value fast: register repos, generate guidance, run validation, produce a report.

2. Standards commoditization

If `AGENTS.md` and `SKILL.md` become widely supported, file generation alone is weak.

Mitigation: make generated files a delivery mechanism, not the product core.

3. Confusing category

The project can sound like several things at once: manifest tool, agent config manager, cross-repo runner, developer portal, or monorepo alternative.

Mitigation: lead with one concrete workflow: "validate a cross-repo change."

4. Enterprise buyers may want hosted controls

Large teams may ask for policy, RBAC, audit logs, dashboards, PR automation, and central catalogs.

Mitigation: keep MVP local and open-source. Treat hosted/platform features as later proof of demand.

## Go / No-Go Criteria

Go if 5-10 target users say:

- "I already lose time coordinating this across repos."
- "I would maintain this manifest because it saves time."
- "I would let an agent read this before it changes code."
- "The scenario report is useful for review or handoff."
- "This is lighter than moving to a monorepo or adopting a developer portal."

No-go or pivot if users say:

- "We only need `AGENTS.md` generation."
- "Our existing monorepo tooling already solves this."
- "The manifest would be stale immediately."
- "The scenario report does not change our workflow."
- "This feels like another metadata layer without enforcement."

## Market Conclusion

The market is real, but the wedge must be sharp.

Good wedge:

- local;
- multi-repo;
- agent-aware;
- scenario-based;
- safe to inspect before writing;
- useful even without hosted infrastructure.

Weak wedge:

- broad platform;
- generic instructions generator;
- universal agent abstraction;
- cross-repo command runner only.
