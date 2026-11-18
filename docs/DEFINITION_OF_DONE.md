# Definition of Done

A work item (user story, bug, tech-debt task) is **Done** only when every item
below holds. "Done" is indivisible: there is no "mostly done".

## Code

- [ ] Implemented against a linked issue (user story / bug / tech-debt label)
- [ ] Acceptance criteria from the issue are demonstrably met
- [ ] `go test -race ./...` passes with coverage for new logic
- [ ] `golangci-lint run` reports 0 issues (fixed **or** individually justified with `//nolint <reason>`)
- [ ] `go vet ./...` clean
- [ ] No new `govulncheck` findings introduced
- [ ] No secrets, private keys or session data committed
- [ ] Errors are handled or explicitly justified; no swallowed errors

## Process

- [ ] Opened as a PR against `main` (never pushed directly)
- [ ] CI is green on the PR (vet, lint, test, build matrix)
- [ ] Reviewed and approved by at least one other team member
- [ ] Conventional commit message (`feat:`, `fix:`, `ci:`, `docs:`, `test:`, `refactor:`, `chore:`)
- [ ] `CHANGELOG.md` updated for user-visible changes
- [ ] Documentation updated (README, docs/, inline godoc for new packages)
- [ ] Issue closed with a reference to the merging PR

## Forensics-specific gates

- [ ] Chain-of-custody behavior unchanged or improved (no silent audit gaps)
- [ ] No shell interpolation added anywhere; arguments stay argv-style
- [ ] Tool catalog changes keep `TestRegistryIntegrity` green
