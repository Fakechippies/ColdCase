# Changelog

All notable changes to ColdCase are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Local web console (`coldcase serve`) with dashboard, tool catalog,
  availability graphs, evidence-upload runner and signed session review.
- Detached console lifecycle: `coldcase serve up/down/status/run`, with
  verified graceful shutdown and pid-file management.
- Grouped `--help` output separating ColdCase commands from forensics tools.
- Ed25519 release signing (`cmd/coldcase-sign`, GoReleaser + CI workflows).
- CI: vet/lint/test/build matrix, govulncheck, CodeQL, Dependabot, container
  publish to GHCR.
- Agile process artifacts: Definition of Done, contributing guide, issue/PR
  templates, retrospective template.
- First unit tests (session crypto, tool catalog integrity).

### Security
- Tool execution remains argv-only (no shell interpolation).
- Session audit entries are signed with Ed25519; tampering is detected
  by `coldcase session verify`.
