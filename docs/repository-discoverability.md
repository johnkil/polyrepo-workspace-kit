# Repository Discoverability

Status: public repository metadata and GitHub discoverability checklist.

This document tracks repository-level discoverability work that is not fully
captured by source code or release notes.

## GitHub Metadata

Repository description:

```text
Thin coordination layer for local polyrepo workspaces and coding-agent guidance
```

Topics:

- `agentic-coding`
- `ai-agents`
- `claude-code`
- `cli`
- `codex`
- `coding-agents`
- `developer-tools`
- `devtools`
- `github-copilot`
- `golang`
- `multi-repo`
- `polyrepo`
- `repository-management`
- `workspace-management`

Homepage is intentionally blank until the project has a real GitHub Pages or
external documentation landing page. Do not point homepage back to the same
repository URL.

## Social Preview

Use this repository asset for GitHub's social preview image:

```text
docs/assets/social-preview.jpg
```

It is prepared for GitHub's recommended high-quality social preview dimensions:

- `1280x640`
- under `1 MB`
- solid background for predictable rendering on social platforms

Upload it manually from:

```text
GitHub repository -> Settings -> Social preview -> Edit -> Upload an image
```

GitHub does not expose this setting through the normal repository source files,
so the committed asset is the source-controlled input and the uploaded social
preview is repository metadata.

## Community Profile

The repository includes:

- `README.md`
- `LICENSE`
- `CONTRIBUTING.md`
- `CODE_OF_CONDUCT.md`
- `SECURITY.md`
- `SUPPORT.md`
- issue forms under `.github/ISSUE_TEMPLATE/`
- `.github/pull_request_template.md`

GitHub community profile health was verified at `100%` after adding issue and
pull request templates.

## Future GitHub Pages

Only add GitHub Pages when the project needs an indexed documentation landing
page outside the repository README. A useful first version would be:

- one-page overview;
- install command;
- minimal demo;
- links to `docs/spec.md`, `docs/install.md`, and `docs/release.md`;
- no new product claims beyond the shipped CLI and documented roadmap.
