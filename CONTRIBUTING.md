# Contributing to ColdCase

Thanks for contributing. This document is the short version of how we work;
the full quality bar lives in [docs/DEFINITION_OF_DONE.md](docs/DEFINITION_OF_DONE.md).

## Branches

| Branch | Purpose |
|--------|---------|
| `main` | Always releasable. Protected: PRs only, CI must pass. |
| `feat/<topic>` | New features (e.g. `feat/upload-runner`) |
| `fix/<topic>` | Bug fixes (e.g. `fix/sse-buffer`) |
| `chore/<topic>` | Tooling, deps, docs-only work |

Work happens on a feature branch, never directly on `main`.

## Commits

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(web): add drag-and-drop evidence upload
fix(runner): remap host paths with spaces correctly
test(session): cover tampered signature rejection
ci: add govulncheck to the security workflow
docs: document serve down/up lifecycle
chore(deps): bump cobra to v1.10.2
refactor(server): split view models out of handlers
```

Small, logical commits. One concern per commit — reviewable in minutes.

## Pull requests

1. Link the issue it implements (`Closes #123`).
2. Fill out the PR template (acceptance criteria + DoD checklist).
3. CI must be green before review.
4. Squash or merge commit per repo default; the PR title becomes the changelog entry.

## Sprint process

- Work is tracked as user-story issues (see `.github/ISSUE_TEMPLATE/user_story.yml`)
  on the project board: `Backlog → Sprint Backlog → In Progress → In Review → Done`.
- Sprints are two weeks. The sprint goal is fixed at planning; mid-sprint
  additions go to the backlog unless they are production incidents.
- Retrospectives are written after every sprint into
  `docs/retrospectives/` (see `TEMPLATE.md`) — honest notes, actionable items.

## Security

- Never commit keys, session data, evidence files, or tokens.
- Report vulnerabilities privately (GitHub Security Advisories), not as public issues.
- `gosec` findings are fixed or justified inline with `//nolint <reason>`.

## Local checks before pushing

```bash
make build test vet lint vuln
```
