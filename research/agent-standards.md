# Agent Standards Research

Date: 2026-04-19
Scope: `AGENTS.md`, `SKILL.md`, `CLAUDE.md`, Copilot instructions, OpenCode, MCP, and adapter implications.

## Question

Are agent guidance formats fragmented enough to justify `wkit`, or converging enough to make adapters a commodity?

## Short Answer

Both are true.

The ecosystem is fragmented in file names, scopes, discovery rules, skill formats, custom agent formats, command models, permissions, and MCP configuration. That creates real short-term value for adapters.

At the same time, the ecosystem is converging around a small number of concepts:

- repository instruction files;
- `AGENTS.md`;
- `SKILL.md`;
- tool-specific memory files;
- MCP for tools/context;
- generated or discovered custom agents.

Therefore `wkit` should support adapters, but should not make adapters the main moat.

## AGENTS.md

[AGENTS.md](https://agents.md/) presents itself as a simple, open Markdown format for guiding coding agents and says it is used by over 60k open-source projects.

[OpenAI Codex AGENTS.md docs](https://developers.openai.com/codex/guides/agents-md) describe `AGENTS.md` as custom instructions for Codex, with scoped behavior by directory. This supports the idea that project-specific guidance should be explicit and checked into repositories.

[GitHub Copilot repository instructions](https://docs.github.com/copilot/how-tos/configure-custom-instructions/add-repository-instructions) now describe multiple instruction types, including repository-wide `.github/copilot-instructions.md`, path-specific `.github/instructions/*.instructions.md`, and agent instructions through `AGENTS.md`.

Implication for `wkit`:

- Generating `AGENTS.md` is necessary.
- It is not sufficient.
- If AGENTS.md becomes ubiquitous, `wkit` must add value before the generated file: repo graph, context, change, scenario, and validation.

## Agent Skills / SKILL.md

[OpenAI Codex Skills](https://developers.openai.com/codex/skills) use a `SKILL.md` file with metadata, instructions, and optional scripts/references. The docs emphasize progressive disclosure: metadata is visible first, and full instructions load only when the skill is used.

[Agent Skills](https://agentskills.io/) frames `SKILL.md` as an open standard for packaging agent capabilities.

[Claude Code Skills](https://code.claude.com/docs/en/skills) also use `SKILL.md`, support optional supporting files and scripts, and recommend keeping the main skill file focused.

[OpenCode Skills](https://opencode.ai/docs/skills) discovers skills from several locations including `.opencode/skills`, `.claude/skills`, and `.agents/skills`.

Implication for `wkit`:

- `guidance/skills/<name>/SKILL.md` is a good canonical shape.
- Generated outputs should target multiple skill locations.
- Skills should stay task-specific, not become a dump of all workspace context.
- Large reference material should be linked and loaded only when needed.

## Claude Code Memory

[Claude Code memory docs](https://code.claude.com/docs/en/memory) define `CLAUDE.md` files for project, user, and organization scope. Claude treats them as context rather than enforced configuration. The docs also explain that Claude reads `CLAUDE.md`, not `AGENTS.md`, but a `CLAUDE.md` can import `AGENTS.md`.

Implication for `wkit`:

- A Claude adapter can be thin if `AGENTS.md` is canonical enough.
- `CLAUDE.md` should probably import generated portable guidance rather than duplicate it.
- Claude-specific guidance belongs below the import.
- User-scope writes need careful plan/diff/apply behavior.

## GitHub Copilot

GitHub Copilot supports several instruction surfaces:

- `.github/copilot-instructions.md` for repository-wide custom instructions;
- `.github/instructions/*.instructions.md` for path-specific instructions with `applyTo` frontmatter;
- `AGENTS.md` agent instructions;
- custom agents as Markdown files in `.github/agents/` or user-level locations.

Sources:

- [Add repository instructions](https://docs.github.com/copilot/how-tos/configure-custom-instructions/add-repository-instructions)
- [About custom agents](https://docs.github.com/en/copilot/concepts/agents/coding-agent/about-custom-agents)
- [Create custom agents for Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/create-custom-agents-for-cli)

Implication for `wkit`:

- Copilot adapter should start with repository instructions and `AGENTS.md`.
- Path-specific instructions are useful but can create conflict and scope complexity.
- Custom agents should remain out of core v0; they are tool-specific and still evolving.

## OpenCode

[OpenCode Agents](https://opencode.ai/docs/agents) supports primary agents and subagents, with model, prompts, tools, and permissions. [OpenCode Skills](https://opencode.ai/docs/skills) supports `SKILL.md` discovery across OpenCode, Claude-compatible, and `.agents` locations.

Implication for `wkit`:

- Skills are portable enough for v0.
- Agents/subagents are not portable enough for the canonical model.
- OpenCode adapter can be useful, but custom agent generation should remain optional and non-canonical.

## MCP

[Model Context Protocol](https://modelcontextprotocol.io/docs/learn/architecture) standardizes how applications provide context and tools to LLMs, using JSON-RPC 2.0 and a client-server model. MCP is becoming an important interoperability layer.

Implication for `wkit`:

- MCP is strategically relevant.
- MCP should not be required for the first MVP.
- A future `wkit` MCP server could expose workspace, repos, relations, changes, scenarios, and reports to agents.
- MCP configuration generation should be treated as a pack/adapter concern until real usage validates it.

## Adapter Strategy

Recommended adapter principles:

- Generated files must contain ownership markers.
- Generated files must be idempotent.
- Generated files must explain their source, for example "generated from `guidance/rules` and `coordination/workspace.yaml`."
- Repo-scope writes should be the default.
- User-scope writes should require explicit `--scope user`, exact target listing, and backups.
- Tool-native custom agents, subagents, commands, hooks, and MCP config should be optional outputs, not canonical model entities.

## Strategic Risk

The more the ecosystem converges, the less valuable it is to merely write files.

The durable value must be:

- canonical workspace relationships;
- cross-repo context selection;
- change and scenario semantics;
- validation evidence;
- safe adapter installation.

In other words: `wkit` should be the source of truth behind agent files, not the agent-file product itself.
