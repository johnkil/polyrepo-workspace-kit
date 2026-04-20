# Security Policy

## Supported Versions

Polyrepo Workspace Kit is early v0.x software. Security fixes target the current `main` development line and the latest public v0.x release.

## Reporting a Vulnerability

Do not report suspected vulnerabilities in public issues.

Use GitHub private vulnerability reporting:

https://github.com/johnkil/polyrepo-workspace-kit/security/advisories/new

If private vulnerability reporting is temporarily unavailable, do not open a public issue with vulnerability details. Contact the maintainer out of band before disclosure.

Please include:

- affected command or workflow;
- operating system and Go version;
- reproduction steps;
- expected and actual behavior;
- whether local files, generated adapter outputs, or repo checkouts are affected.

## Security-Relevant Areas

Please pay special attention to:

- path traversal and symlink escapes;
- generated file overwrite behavior;
- backup behavior;
- installer source and target symlink containment;
- scenario command execution boundaries;
- validation of YAML manifests and report paths;
- Git subprocess handling.

## Project Posture

`wkit` is a local CLI. It does not run a daemon, host a service, load secrets, or execute scenario commands through a shell by default. Repo-local scripts are responsible for any shell features, environment setup, or secret loading they require.

Generated guidance installs are derived file writes. Installer planning, diffing, and applying should treat symlinked sources, symlinked targets, and target parent directories that resolve outside their intended boundary as blocked rather than following them.
