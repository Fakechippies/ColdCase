# ColdCase Structure and Architecture Reference

This document is the canonical architecture map for ColdCase. It is intentionally detailed so it can be used as source material for system architecture diagrams, component diagrams, sequence diagrams, and data-flow diagrams later.

ColdCase is a Go-based CLI that exposes many third-party forensic tools behind one unified command surface. The core architectural idea is simple:

1. The `coldcase` binary exposes a normalized CLI built with Cobra.
2. Each forensic capability is represented by a thin wrapper package in `pkg/`.
3. Every wrapper delegates actual execution to the shared hybrid runner in `pkg/runner`.
4. The runner prefers native host binaries and falls back to a Docker/Podman container when the binary is absent.
5. If a forensic session is active, command execution is also routed through the audit logging path in `pkg/session`.

The system is therefore not a monolith that implements forensic parsing itself. It is an orchestration and control layer around many specialized external binaries and Python suites.

## 1. High-Level System Model

At the highest level, the repository is split into five architectural zones:

- `cmd/coldcase/`
  The CLI composition layer. This is where top-level commands and subcommands are registered.
- `pkg/runner/`
  The execution engine. This is the most important runtime component because almost every tool wrapper eventually calls into it.
- `pkg/session/`
  The chain-of-custody and audit subsystem. It persists sessions, hashes inputs, stores output, and signs log entries.
- `pkg/<category>/`
  Domain-specific adapter packages. These define tool metadata and convert CLI requests into `runner.RunOpts`.
- Bundled external suites and environment assets
  `DidierStevensSuite/`, `volatility3/`, and `Dockerfile` provide the external tooling that ColdCase wraps or ships alongside.

The control flow is:

`user shell -> cobra command -> package-specific tool wrapper -> runner -> native process or container process -> optional session logging -> terminal output`

## 2. Repository Layout

```text
ColdCase/
├── cmd/
│   └── coldcase/
│       ├── main.go
│       ├── utils.go
│       ├── deps.go
│       ├── platform.go
│       ├── container.go
│       ├── session.go
│       └── keys.go
├── pkg/
│   ├── tools/
│   ├── runner/
│   ├── session/
│   ├── didier/
│   ├── volatility3/
│   ├── network/
│   ├── carving/
│   ├── malware/
│   ├── hashing/
│   ├── timeline/
│   ├── mobile/
│   ├── windows/
│   ├── steg/
│   ├── sysutils/
│   ├── sleuthkit/
│   ├── binwalk/
│   └── exiftool/
├── DidierStevensSuite/
├── volatility3/
├── Dockerfile
├── README.md
├── go.mod
└── bin/
```

## 3. Architectural Layers

### Layer A: CLI Composition Layer

Location: `cmd/coldcase/`

Responsibility:

- Define the root Cobra command.
- Register every subcommand.
- Present a stable CLI contract to the operator.
- Translate command-line arguments into calls on internal packages.

Important design detail:

- ColdCase does not centralize all command registration in one place.
- Instead, several files in `cmd/coldcase/` use `init()` to attach commands to `rootCmd`.
- This means startup is distributed across multiple compilation units.

Files and responsibilities:

- [`cmd/coldcase/main.go`](/home/chips/Projects/ColdCase/cmd/coldcase/main.go)
  Defines `rootCmd`, registers most tool wrappers, and adds utility commands like `list` and `check`.
- [`cmd/coldcase/utils.go`](/home/chips/Projects/ColdCase/cmd/coldcase/utils.go)
  Implements the read-only discovery commands `list` and `check`.
- [`cmd/coldcase/deps.go`](/home/chips/Projects/ColdCase/cmd/coldcase/deps.go)
  Registers dependency management commands such as `install` and `deps`.
- [`cmd/coldcase/platform.go`](/home/chips/Projects/ColdCase/cmd/coldcase/platform.go)
  Prints platform-specific setup instructions.
- [`cmd/coldcase/container.go`](/home/chips/Projects/ColdCase/cmd/coldcase/container.go)
  Registers container lifecycle commands.
- [`cmd/coldcase/session.go`](/home/chips/Projects/ColdCase/cmd/coldcase/session.go)
  Registers the forensic session lifecycle commands.
- [`cmd/coldcase/keys.go`](/home/chips/Projects/ColdCase/cmd/coldcase/keys.go)
  Registers Ed25519 key management commands.

### Layer B: Tool Adapter Layer

Location: `pkg/<category>/`

Responsibility:

- Model each external tool as a Go object with `Name()`, `Description()`, and `Run([]string) error`.
- Keep category-specific metadata together.
- Normalize differences between displayed command name and actual binary name.
- Route all real execution into the runner.

Important design detail:

- These adapters are intentionally thin.
- They do not implement forensic parsing logic themselves.
- Their primary job is metadata definition plus conversion into `runner.RunOpts`.

### Layer C: Execution Layer

Location: `pkg/runner/`

Responsibility:

- Decide whether execution happens on the host or in a container.
- Detect which container runtime is available.
- Detect filesystem paths in arguments and convert them into container mounts.
- Invoke external processes and return their output/error state.
- Integrate execution with session logging if a session is active.

This is the architectural center of ColdCase.

### Layer D: Forensic Session Layer

Location: `pkg/session/`

Responsibility:

- Persist session metadata.
- Track state transitions such as unlocked, locked, and sealed.
- Hash input files referenced by commands.
- Persist command output to disk.
- Sign log entries using Ed25519 when signing is enabled.
- Export session reports in JSON, Markdown, and HTML.

### Layer E: External Tooling Layer

Location:

- `DidierStevensSuite/`
- `volatility3/`
- container image built from `Dockerfile`
- host-installed binaries on the operator machine

Responsibility:

- Supply the actual forensic functionality.
- Execute the specialized parsing, carving, inspection, and analysis work.

ColdCase itself is the control plane around these tools.

## 4. Bootstrap and Command Registration Model

The CLI is assembled during program initialization.

Entry point:

- [`cmd/coldcase/main.go`](/home/chips/Projects/ColdCase/cmd/coldcase/main.go#L27)
  `main()` calls `rootCmd.Execute()`.

Registration pattern:

- `main.go` registers most top-level forensic tools.
- `container.go`, `deps.go`, `platform.go`, `session.go`, and `keys.go` each attach their own command trees in `init()`.
- Because of this, the full command graph is the sum of multiple `init()` functions across the package.

Operational implication:

- Anyone generating a command hierarchy diagram should treat `cmd/coldcase/` as a distributed command-registration package, not a single-file CLI definition.

Main registration groups:

- Tool wrappers from category packages
- Built-in info commands: `list`, `check`, `platform`
- Dependency commands: `install`, `deps install`, `deps check`, `deps update`
- Container commands: `container status`, `container build`, `container pull`, `container shell`
- Session commands: `session start`, `resume`, `stop`, `status`, `list`, `verify`, `lock`, `unlock`, `seal`, `export`
- Key commands: `keys generate`, `keys export-public`

## 5. Request Execution Flow

The standard execution sequence for a forensic tool command is:

1. The operator runs a CLI command such as `coldcase tshark -r capture.pcap`.
2. Cobra resolves the subcommand to a registered tool wrapper.
3. The wrapper constructs `runner.RunOpts`.
4. `pkg/runner.Run` checks whether an active forensic session exists.
5. `pkg/runner.Run` checks whether the requested binary exists on the host.
6. If present, the tool executes natively.
7. If missing, the runner detects `docker` or `podman` and executes inside the configured container image.
8. If a session logger is active, the runner hashes input files, stores combined output, and writes an audit entry.
9. Output is printed to the terminal regardless of execution mode.

This means the tool wrappers do not know or care whether the tool runs natively or in a container. That decision is intentionally centralized.

## 6. Core Execution Engine

Primary files:

- [`pkg/runner/runner.go`](/home/chips/Projects/ColdCase/pkg/runner/runner.go)
- [`pkg/runner/volumes.go`](/home/chips/Projects/ColdCase/pkg/runner/volumes.go)

### 6.1 `RunOpts`

`RunOpts` is the internal execution contract used by every wrapper.

Fields:

- `Binary`
  Name of the executable to invoke on host or inside container.
- `Args`
  Raw argument vector passed through to the tool.
- `NeedsRoot`
  Signals that container execution needs extra capabilities such as `NET_ADMIN` and `NET_RAW`.
- `WorkDir`
  Optional working directory override.

### 6.2 Native vs Container Routing

Routing logic in [`pkg/runner/runner.go`](/home/chips/Projects/ColdCase/pkg/runner/runner.go#L36):

- If `tools.CheckToolInstalled(opts.Binary)` succeeds, the runner uses `runNative`.
- Otherwise it calls `detectRuntime`.
- If neither `docker` nor `podman` is available, execution fails.
- If a runtime is available, it uses `runInContainer`.

Runtime selection:

- Default search order is `docker`, then `podman`.
- `COLDCASE_RUNTIME` overrides runtime selection.
- `COLDCASE_IMAGE` overrides the default image name `coldcase:latest`.

### 6.3 Native Execution

Native execution characteristics:

- Uses `exec.Command`.
- Uses `CombinedOutput`, so stdout and stderr are captured together.
- Prints captured output back to the terminal with `fmt.Print`.
- Returns the combined output buffer to the caller for session logging.

Architectural consequence:

- The runner captures output centrally, which makes session logging possible without changing each wrapper package.

### 6.4 Container Execution

Container execution characteristics:

- Uses `docker run` or `podman run`.
- Runs with `--rm -i` and `-t` when stdin is a TTY.
- Adds `--cap-add NET_ADMIN` and `--cap-add NET_RAW` for tools flagged with `NeedsRoot`.
- Appends auto-generated bind mounts for detected filesystem paths.
- Appends `-w` if `WorkDir` is set.
- Invokes the image followed by the requested binary and remapped arguments.

### 6.5 Path Detection and Bind Mount Generation

Implemented in [`pkg/runner/volumes.go`](/home/chips/Projects/ColdCase/pkg/runner/volumes.go).

Purpose:

- When a user passes local evidence files or directories as arguments, those paths must exist inside the container too.
- ColdCase scans arguments for path-like values and maps them into container-visible paths.

Algorithm:

1. Examine each argument.
2. Treat values beginning with `/`, `./`, `../`, or containing a path separator as possible paths.
3. Convert candidate paths to absolute paths.
4. Ignore non-existent paths.
5. For each valid path, mount its parent directory read-only into `/data/volN`.
6. Replace the original argument with the corresponding container-side path.

Example:

- Host argument: `/evidence/case1/dump.pcap`
- Host directory mounted as: `/evidence/case1:/data/vol0:ro`
- Remapped argument inside container: `/data/vol0/dump.pcap`

Important boundary:

- Path detection is heuristic.
- It works best for arguments that are direct file paths.
- It is less aware of complex flag/value syntaxes embedded into a single shell token.

## 7. Session and Audit Subsystem

Primary files:

- [`pkg/session/types.go`](/home/chips/Projects/ColdCase/pkg/session/types.go)
- [`pkg/session/manager.go`](/home/chips/Projects/ColdCase/pkg/session/manager.go)
- [`pkg/session/logger.go`](/home/chips/Projects/ColdCase/pkg/session/logger.go)
- [`pkg/session/crypto.go`](/home/chips/Projects/ColdCase/pkg/session/crypto.go)
- [`pkg/session/export.go`](/home/chips/Projects/ColdCase/pkg/session/export.go)

This subsystem makes ColdCase more than a generic wrapper runner. It adds stateful forensic workflow controls.

### 7.1 Session Persistence Model

Base directory:

- `~/.coldcase/`

Subdirectories used now:

- `~/.coldcase/sessions/<session-id>/session.json`
- `~/.coldcase/sessions/<session-id>/outputs/`
- `~/.coldcase/keys/private.key`
- `~/.coldcase/keys/public.key`

Session lifecycle manager:

- `session.NewManager()` ensures `~/.coldcase/sessions` exists.
- `Create`, `Load`, `Save`, and `List` operate on JSON-backed session state.

### 7.2 Session Data Model

Core struct:

- `Session`

Important fields:

- `ID`
- `Investigator`
- `Email`
- `Created`
- `State`
- `Encrypted`
- `Signed`
- `Commands`
- `Evidence`
- `Signature`
- `SealedAt`

Command log model:

- `CommandEntry`

Important fields:

- exact command metadata
- input file metadata
- output preview
- output file path
- duration
- working directory
- optional signature

Evidence tracking model:

- `EvidenceFile`
- `FileMetadata`

### 7.3 Session State Machine

Defined states:

- `unlocked`
  Commands may execute and be logged.
- `locked`
  Session is read-only from the runner’s perspective. Tool execution under that active session is rejected.
- `sealed`
  Session is treated as permanently closed.

Behavior in runner:

- If `COLDCASE_SESSION_ID` is set, the runner loads the session.
- If the session is not `unlocked`, `runner.Run` returns an error before tool execution.

### 7.4 Logging Flow During Command Execution

When a session is active and unlocked:

1. `runner.Run` notes the start time.
2. The external tool runs.
3. Each CLI argument is checked with `os.Stat`; any existing path is treated as an input file candidate.
4. `Logger.HashInputFile` collects metadata and SHA-256 hash.
5. Combined stdout/stderr is stored under `outputs/`.
6. A truncated preview is stored inline in `session.json`.
7. A `CommandEntry` is appended and persisted.

### 7.5 Signing and Verification

Signing implementation:

- Key generation uses Ed25519.
- Private/public keys are stored under `~/.coldcase/keys/`.
- If session signing is enabled and a private key can be loaded, each command entry is signed.

Verification implementation:

- `Manager.VerifySession` recreates the signed data string for each command entry.
- The public key is used to validate the stored signature.

Signed fields currently include:

- command index
- UTC timestamp
- full command string
- working directory

### 7.6 Encryption Support

The crypto package contains:

- AES-256-GCM helpers
- PBKDF2-based key derivation
- generic encrypt/decrypt functions

Current architectural status:

- Encryption primitives exist in code.
- Session output persistence is not fully encrypted yet.
- `Logger.SaveOutput` explicitly notes that encrypted writes are a future path.

For diagram accuracy, encryption should be shown as a partially implemented subsystem rather than a fully enforced persistence layer.

### 7.7 Export Paths

Export formats:

- JSON
- Markdown
- HTML

Export role:

- Transform session state into human-readable investigation reports.
- Provide a stable reporting endpoint separate from execution and persistence.

## 8. Tool Adapter Packages

Each package under `pkg/` is a registry plus lightweight execution adapter.

Common pattern:

- Define a small struct containing metadata.
- Implement `Name`, `Description`, and `Run`.
- Expose `Tools()` or `New()`.
- Call `runner.Run`.

### 8.1 Shared Tool Contract

Shared interface location:

- [`pkg/tools/tools.go`](/home/chips/Projects/ColdCase/pkg/tools/tools.go)

This package defines:

- `Tool` interface
- `ExecuteCommand` helper
- `CheckToolInstalled` helper

Architectural role:

- Provide a minimal shared abstraction for command registration and dependency checks.

### 8.2 DidierStevens Integration

Package:

- [`pkg/didier/didier.go`](/home/chips/Projects/ColdCase/pkg/didier/didier.go)

How it works:

- The displayed commands such as `pdfid` and `oledump` are mapped to Python scripts under `DidierStevensSuite/`.
- The wrapper actually runs `python3 <scriptPath> ...args`.
- Therefore the real runtime dependency is `python3` plus the script bundle.

Architectural relationship:

- CLI command name -> Didier wrapper -> `runner.Run(Binary="python3", Args=[scriptPath, ...])`

### 8.3 Volatility3 Integration

Package:

- [`pkg/volatility3/volatility3.go`](/home/chips/Projects/ColdCase/pkg/volatility3/volatility3.go)

How it works:

- ColdCase exposes Volatility3 plugins as first-class CLI commands such as `windows.pslist`.
- The wrapper translates them into `python3 volatility3/vol.py <plugin> ...args`.
- Raw entry points like `vol` and `volshell` are also represented.

Architectural relationship:

- CLI plugin command -> Volatility wrapper -> `runner.Run(Binary="python3", Args=[vol.py, plugin, ...])`

Important note:

- The plugin name is modeled separately from the displayed command metadata.
- Dependency checks confirm both `python3` availability and the presence of `volatility3/vol.py`.

### 8.4 Timeline and Plaso Integration

Package:

- [`pkg/timeline/timeline.go`](/home/chips/Projects/ColdCase/pkg/timeline/timeline.go)

How it works:

- `timeline.Tools()` returns both Plaso binaries and non-Plaso timeline tools.
- `main.go` creates a special grouped `plaso` command tree.
- `log2timeline` is exposed as `plaso parse`.
- `psort` is exposed as `plaso sort`.
- `psteal` is exposed as `plaso psteal`.
- `plaso parsers` dispatches `log2timeline --parsers list`.
- Non-Plaso tools like `hayabusa`, `evtx_dump`, `timeliner`, and `chainsaw` remain top-level commands.

Architectural meaning:

- The timeline package is both a tool registry and a source for CLI reshaping logic in `main.go`.

### 8.5 Network Tools

Package:

- [`pkg/network/network.go`](/home/chips/Projects/ColdCase/pkg/network/network.go)

Special behavior:

- Some tools are marked `needsRoot`.
- When routed through a container, the runner adds network-related Linux capabilities.

Examples of elevated tools:

- `tcpdump`
- `zeek`
- `ngrep`
- `tcpreplay`
- `argus`
- `p0f`

### 8.6 Other Adapter Packages

These all follow the same basic wrapper model:

- [`pkg/carving/carving.go`](/home/chips/Projects/ColdCase/pkg/carving/carving.go)
- [`pkg/malware/malware.go`](/home/chips/Projects/ColdCase/pkg/malware/malware.go)
- [`pkg/hashing/hashing.go`](/home/chips/Projects/ColdCase/pkg/hashing/hashing.go)
- [`pkg/mobile/mobile.go`](/home/chips/Projects/ColdCase/pkg/mobile/mobile.go)
- [`pkg/windows/windows.go`](/home/chips/Projects/ColdCase/pkg/windows/windows.go)
- [`pkg/steg/steg.go`](/home/chips/Projects/ColdCase/pkg/steg/steg.go)
- [`pkg/sysutils/sysutils.go`](/home/chips/Projects/ColdCase/pkg/sysutils/sysutils.go)
- [`pkg/sleuthkit/sleuthkit.go`](/home/chips/Projects/ColdCase/pkg/sleuthkit/sleuthkit.go)
- [`pkg/binwalk/binwalk.go`](/home/chips/Projects/ColdCase/pkg/binwalk/binwalk.go)
- [`pkg/exiftool/exiftool.go`](/home/chips/Projects/ColdCase/pkg/exiftool/exiftool.go)

Common variations across them:

- Some expose a single tool via `New()`.
- Some expose many tools via `Tools()`.
- Some separate display name from actual binary name via a `bin` field.
- All eventually delegate to `runner.Run`.

## 9. Container Architecture

Primary file:

- [`Dockerfile`](/home/chips/Projects/ColdCase/Dockerfile)

Purpose:

- Provide a fallback runtime containing the binaries ColdCase expects to find.
- Reduce host dependency burden.
- Keep CLI semantics consistent whether tools are installed locally or not.

### 9.1 Build Stages

Stage 1: `system`

- Installs operating-system packages and common forensic binaries on Ubuntu 24.04.
- Includes core runtimes such as Python, Perl, Ruby, Git, and network tools.

Stage 2: `python-tools`

- Creates a Python virtual environment under `/opt/coldcase-venv`.
- Installs Python-based forensic packages and supporting ecosystems.
- Installs additional tooling such as `zsteg` and `bulk_extractor`.

Stage 3: `final`

- Copies repository-bundled suites:
  - `volatility3/`
  - `DidierStevensSuite/`
- Installs Volatility3 in editable mode.
- Pulls extra standalone binaries like Hayabusa and Chainsaw when available.
- Sets `/data` as the working mount point for evidence access.

### 9.2 Relationship to the Runner

The runner assumes:

- The container image contains the same tool names passed as `opts.Binary`.
- The image can access user evidence through bind mounts.
- The tool command line can remain mostly unchanged after path remapping.

So the Docker image is not optional architecture fluff. It is a runtime extension of the execution engine.

## 10. Built-In Operational Commands

ColdCase includes non-forensic operational command groups that support the main architecture.

### 10.1 Discovery Commands

Implemented in [`cmd/coldcase/utils.go`](/home/chips/Projects/ColdCase/cmd/coldcase/utils.go):

- `list`
  Enumerates all registered wrappers and utility commands.
- `check`
  Performs local dependency checks and reports container fallback availability.

### 10.2 Dependency Commands

Implemented in [`cmd/coldcase/deps.go`](/home/chips/Projects/ColdCase/cmd/coldcase/deps.go):

- `install`
  Installs host dependencies or builds the container.
- `deps install`
  Installs Python dependencies, mainly for Volatility3.
- `deps check`
  Reports missing dependencies.
- `deps update`
  Updates Python dependencies.

Architectural role:

- These commands are the environment bootstrap layer.
- They are not part of the execution path, but they determine whether native execution succeeds.

### 10.3 Platform Guide Command

Implemented in [`cmd/coldcase/platform.go`](/home/chips/Projects/ColdCase/cmd/coldcase/platform.go):

- Provides OS-specific setup instructions.
- Detects package-manager availability.

Architectural role:

- This is the operator guidance layer for installation and environment preparation.

### 10.4 Container Management Commands

Implemented in [`cmd/coldcase/container.go`](/home/chips/Projects/ColdCase/cmd/coldcase/container.go):

- `container status`
- `container build`
- `container pull`
- `container shell`

Architectural role:

- This command group is the manual control plane for the same container subsystem the runner uses automatically.

### 10.5 Session and Key Commands

Implemented in:

- [`cmd/coldcase/session.go`](/home/chips/Projects/ColdCase/cmd/coldcase/session.go)
- [`cmd/coldcase/keys.go`](/home/chips/Projects/ColdCase/cmd/coldcase/keys.go)

Architectural role:

- These commands set up and manage the forensic accountability subsystem.
- They do not execute forensic tools directly, but they change how the runner behaves during later executions.

## 11. Data and Persistence Flows

There are three main data classes in ColdCase:

### 11.1 Ephemeral Runtime Data

Examples:

- CLI args
- runtime-selected binary name
- detected container runtime
- bind mounts
- working directory
- combined process output buffer

This data exists during command execution and is mainly owned by `pkg/runner`.

### 11.2 Session Metadata

Examples:

- session identity
- investigator info
- command history
- evidence hashes
- state transitions
- signatures

This data is persisted in `session.json` under `~/.coldcase/sessions/<id>/`.

### 11.3 External Evidence and Tool Assets

Examples:

- PCAPs
- memory images
- registry hives
- firmware images
- media artifacts
- bundled Python tools

This data is not owned by ColdCase, but ColdCase references it, mounts it, hashes it, and passes it to external tools.

## 12. Dependency and Integration Graph

Internal dependency direction is mostly one-way:

- `cmd/coldcase` depends on category packages, `runner`, and `session`
- category packages depend on `runner`
- `runner` depends on `session` and `tools`
- `session` is mostly self-contained
- `tools` is the lowest shared helper package

External integration points:

- host shell environment
- host filesystem
- `docker` or `podman`
- host-installed binaries
- Python runtime
- bundled `DidierStevensSuite`
- bundled `volatility3`
- generated files under `~/.coldcase`

## 13. Sequence Models

### 13.1 Standard Native Execution Sequence

1. User invokes `coldcase <tool> ...`
2. Cobra routes to wrapper
3. Wrapper calls `runner.Run`
4. Runner checks active session
5. Runner checks host binary presence
6. Runner executes native binary
7. Runner prints output
8. Runner logs command if session is active

### 13.2 Container Fallback Sequence

1. User invokes `coldcase <tool> ...`
2. Cobra routes to wrapper
3. Wrapper calls `runner.Run`
4. Runner checks host binary presence and fails lookup
5. Runner detects container runtime
6. Runner scans args for paths
7. Runner generates bind mounts and remapped args
8. Runner runs `docker/podman run ... image binary args`
9. Runner prints output
10. Runner logs command if session is active

### 13.3 Signed Session Execution Sequence

1. User exports `COLDCASE_SESSION_ID`
2. User invokes a forensic tool
3. Runner loads session
4. Runner confirms state is `unlocked`
5. Tool executes
6. Logger hashes referenced input files
7. Logger stores output
8. Logger signs command entry if signing is enabled and key is available
9. Session JSON is updated

## 14. Extension Model

ColdCase is designed to make adding a new wrapped tool cheap.

Typical extension steps:

1. Add a new adapter package or update an existing category package in `pkg/`.
2. Implement the minimal tool interface methods.
3. In `Run`, call `runner.Run` with the correct `Binary`, `Args`, and optional `NeedsRoot`.
4. Register the tool in `cmd/coldcase/main.go` or a relevant command file.
5. Add the corresponding binary to the `Dockerfile` if container fallback should support it.
6. Update discovery/check commands if needed.

Architectural rule:

- New functionality should prefer thin adapters over bespoke execution logic.
- The runner should remain the only place where native-vs-container routing is decided.

## 15. Current Implementation Notes and Accuracy Boundaries

These details matter for architecture diagrams because some repository claims are aspirational while others are fully implemented.

- Session signing is implemented and verified per command entry.
- Session encryption primitives exist, but output encryption is not fully wired into persistence.
- The CLI is modular but not plugin-based; commands are compiled in.
- The execution model is centralized and real; almost every tool path converges on `pkg/runner`.
- Most forensic capability comes from external binaries, not internal parsers.
- Container fallback is real and implemented.
- Tool grouping is partly semantic; most wrappers are metadata registries rather than complex logic modules.
- The old tracked architecture file was named `STRUCTURE` without an extension; this `STRUCTURE.md` is the updated, more explicit documentation artifact.

## 16. Diagram-Friendly Summary

If you need to generate diagrams from this document later, use these component blocks:

- User / Investigator
- Shell / Terminal
- Cobra CLI Layer
- Command Registration Layer (`cmd/coldcase`)
- Tool Adapter Layer (`pkg/<category>`)
- Shared Tool Helpers (`pkg/tools`)
- Hybrid Runner (`pkg/runner`)
- Session Manager (`pkg/session/manager`)
- Session Logger (`pkg/session/logger`)
- Crypto Services (`pkg/session/crypto`)
- Session Report Exporters (`pkg/session/export`)
- Host Binary Environment
- Container Runtime (`docker` or `podman`)
- ColdCase Container Image
- Bundled Python Suites (`DidierStevensSuite`, `volatility3`)
- Evidence Files / Host Filesystem
- Session Storage (`~/.coldcase`)

Primary edges between those blocks:

- User -> Cobra CLI
- Cobra CLI -> Tool Adapter
- Tool Adapter -> Hybrid Runner
- Hybrid Runner -> Host Binary Environment
- Hybrid Runner -> Container Runtime
- Container Runtime -> ColdCase Container Image
- Hybrid Runner -> Evidence Files
- Hybrid Runner -> Session Logger
- Session Logger -> Session Storage
- Session Logger -> Crypto Services
- Session commands -> Session Manager
- Key commands -> Crypto Services + Session Storage
- Volatility/Didier adapters -> bundled Python suites

This is the actual backbone of the current repository.
