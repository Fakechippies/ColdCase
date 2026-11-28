# Changelog

All notable changes to ColdCase are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-11-18

First tagged release.

### Added
- CLI integrating 118+ forensics tools with hybrid native/container execution
  and automatic path remapping.
- Local web console (`coldcase serve`) with dashboard, tool catalog,
  availability graphs, evidence-upload runner and signed session review.
- Detached console lifecycle: `coldcase serve up/down/status/run`, with
  verified graceful shutdown and pid-file management.
- Grouped `--help` output separating ColdCase commands from forensics tools.
- Ed25519 release signing (`cmd/coldcase-sign`, GoReleaser + CI workflows).
- CI: vet/lint/test/build matrix, govulncheck, CodeQL, Dependabot, container
  publish to GHCR.
- Delivery process artifacts: Definition of Done, contributing guide, issue/PR
  templates, retrospective template.
- First unit tests (session crypto, tool catalog integrity, session-ID
  validation).

### Security
- Tool execution remains argv-only (no shell interpolation).
- Session audit entries are signed with Ed25519; tampering is detected
  by `coldcase session verify`.
- Session identifiers validated against an allow-list at the model layer.

## [Unreleased]

### Fixed
- Suite discovery: DidierStevens tools resolved `./DidierStevensSuite`
  relative to the working directory, so an installed binary only worked
  from the repo root. `coldcase` now searches `$COLDCASE_SUITE`, the
  working directory and its parents, the executable's directory, and
  `~/.coldcase/DidierStevensSuite`; `coldcase install` links the suite into
  that last location, and `coldcase suite [link]` reports or creates it.
- Web console: tool stdout and stderr now render — SSE payloads are JSON
  strings and the handlers read them as objects, which silently dropped
  every output chunk.
- Web console: embedded static assets revalidate on load, so browser
  caches can no longer keep stale UI code after an upgrade.
- Web console: run arguments tokenize like a shell (whitespace-separated,
  quotes supported), so pasted command lines work instead of collapsing
  into a single argument.
- Runner: terminal detection used `os.ModeCharDevice`, which also matches
  `/dev/null`; container runs without a real terminal now start correctly.
- Container image builds again on Ubuntu 24.04: `argus` does not exist
  (→ `argus-client` + `argus-server`), pcapfix compiles with
  `build-essential`, optional Python tools install independently, and the
  hayabusa / chainsaw / evtx_dump binaries are fetched from their current
  release formats.

### Changed
- The Run Tool page opens on a generic, file-agnostic tool (`hash`) so the
  first run always produces visible output; `1768` is described correctly
  ("Analyze Cobalt Strike beacons").
- Empty successful runs print a `[no output]` hint suggesting the tool may
  not handle the uploaded file.
