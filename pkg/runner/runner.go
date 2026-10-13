// Package runner provides the hybrid execution engine for ColdCase.
// It runs tools natively if available on PATH; otherwise it transparently
// proxies the invocation through a Docker or Podman container.
package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"coldcase/pkg/session"
	"coldcase/pkg/tools"
)

const (
	// DefaultImage is the container image name used for fallback execution.
	DefaultImage = "coldcase:latest"
	// EnvImage allows overriding the container image via environment variable.
	EnvImage = "COLDCASE_IMAGE"
	// EnvRuntime allows forcing "docker" or "podman" via environment variable.
	EnvRuntime = "COLDCASE_RUNTIME"
)

// RunOpts controls how a tool is executed.
type RunOpts struct {
	// Binary is the name of the host binary to look up (e.g. "tshark").
	Binary string
	// Args are the arguments to pass to the binary.
	Args []string
	// NeedsRoot signals the tool may need elevated privileges (e.g. tcpdump).
	// When running in a container, NET_ADMIN capability is added.
	NeedsRoot bool
	// WorkDir is an optional working directory override.
	WorkDir string
	// SessionID, when set, overrides the COLDCASE_SESSION_ID environment
	// variable for this invocation. This lets a long-lived process (such as the
	// web server) route concurrent executions to different sessions without
	// mutating process-global state.
	SessionID string
}

// Run executes opts.Binary with opts.Args and streams the tool's stdout and
// stderr directly to the current process (matching the historical CLI
// behaviour). Native execution is attempted first. If the binary is not found
// on PATH, Run falls back to running the command inside a container.
func Run(opts RunOpts) error {
	return run(context.Background(), opts, os.Stdout, os.Stderr, isTTY())
}

// RunStream executes opts.Binary with opts.Args exactly like Run, but streams
// stdout and stderr to the supplied writers instead of the process streams.
// This is the entry point used by the web UI to render live tool output.
//
// The same session audit logging / chain-of-custody behaviour as Run is
// preserved: if a session is active (via opts.SessionID or
// COLDCASE_SESSION_ID) and unlocked, the combined output is stored, input
// files are hashed, and a signed CommandEntry is appended.
//
// The provided context cancels the child process; any output produced before
// cancellation is still recorded to the session.
func RunStream(ctx context.Context, opts RunOpts, stdout, stderr io.Writer) error {
	return run(ctx, opts, stdout, stderr, false)
}

// run is the single execution path shared by Run and RunStream.
func run(ctx context.Context, opts RunOpts, stdout, stderr io.Writer, tty bool) error {
	sID := opts.SessionID
	if sID == "" {
		sID = session.GetActiveSessionID()
	}

	var logger *session.Logger
	var sess *session.Session

	if sID != "" {
		m := session.NewManager()
		var err error
		sess, err = m.Load(sID)
		if err == nil {
			if sess.State != session.StateUnlocked {
				return fmt.Errorf("active session '%s' is %s and read-only", sID, sess.State)
			}
			logger = session.NewLogger(sess)
		}
	}

	start := time.Now()

	// Capture combined output for the audit trail while streaming it live.
	var combined bytes.Buffer
	out := io.MultiWriter(stdout, &combined)
	errOut := io.MultiWriter(stderr, &combined)

	var runErr error
	if tools.CheckToolInstalled(opts.Binary) {
		runErr = runNative(ctx, opts, out, errOut)
	} else {
		rt, err := detectRuntime()
		if err != nil {
			return fmt.Errorf("'%s' not found on PATH and no container runtime available: %w", opts.Binary, err)
		}
		runErr = runInContainer(ctx, rt, opts, out, errOut, tty)
	}

	if logger != nil && sess != nil {
		if err := logExecution(logger, sess, opts, start, combined.Bytes(), runErr); err != nil {
			fmt.Fprintf(os.Stderr, "\n[!] failed to write session log: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "\n[*] Signed entry logged to session: %s\n", sID)
		}
	}

	return runErr
}

// logExecution records a completed (or cancelled) execution to the active
// session, hashing any input files referenced by the arguments.
func logExecution(logger *session.Logger, sess *session.Session, opts RunOpts, start time.Time, output []byte, runErr error) error {
	duration := time.Since(start)

	// Map input files.
	var inputFiles []session.FileMetadata
	for _, arg := range opts.Args {
		if _, err := os.Stat(arg); err == nil {
			meta, err := logger.HashInputFile(arg)
			if err == nil {
				inputFiles = append(inputFiles, meta)
			}
		}
	}

	// Save output.
	idx := len(sess.Commands) + 1
	outPath, _ := logger.SaveOutput(idx, opts.Binary, output)

	preview := string(output)
	if len(output) > 500 {
		preview = string(output[:500]) + "..."
	}

	wd, _ := os.Getwd()
	entry := session.CommandEntry{
		Index:            idx,
		Timestamp:        start,
		Command:          opts.Binary,
		FullCommand:      fmt.Sprintf("%s %v", opts.Binary, opts.Args),
		Args:             opts.Args,
		InputFiles:       inputFiles,
		WorkingDirectory: wd,
		ExitCode:         ExitCode(runErr),
		DurationMS:       duration.Milliseconds(),
		OutputPreview:    preview,
		OutputFile:       outPath,
	}
	return logger.LogCommand(entry)
}

// ExitCode translates a process error into a shell-style exit code
// (0 on success, the process exit code for an *exec.ExitError, -1 otherwise).
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// ─── availability ─────────────────────────────────────────────────────────────

// ContainerAvailable reports whether Docker or Podman is available.
func ContainerAvailable() bool {
	_, err := detectRuntime()
	return err == nil
}

// DetectedRuntime returns the name of the available container runtime
// ("docker" or "podman"), or an empty string if none is found.
func DetectedRuntime() string {
	rt, _ := detectRuntime()
	return rt
}

// ImageName returns the container image to use (respects COLDCASE_IMAGE env).
func ImageName() string {
	if img := os.Getenv(EnvImage); img != "" {
		return img
	}
	return DefaultImage
}

// Availability describes how (or whether) a binary can be executed.
type Availability struct {
	// Binary is the underlying executable being looked up.
	Binary string
	// Native is true when the binary exists on the host PATH.
	Native bool
	// Container is true when a container runtime is available for fallback.
	Container bool
	// Runtime is the detected container runtime name, if any.
	Runtime string
	// Image is the configured container image name.
	Image string
}

// Available reports whether the binary can run natively or via container.
func (a Availability) Available() bool { return a.Native || a.Container }

// CheckAvailability inspects the host for a binary and the container runtime.
func CheckAvailability(binary string) Availability {
	rt := DetectedRuntime()
	return Availability{
		Binary:    binary,
		Native:    tools.CheckToolInstalled(binary),
		Container: rt != "",
		Runtime:   rt,
		Image:     ImageName(),
	}
}

// ─── internal ─────────────────────────────────────────────────────────────────

func runNative(ctx context.Context, opts RunOpts, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, opts.Binary, opts.Args...)
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runInContainer(ctx context.Context, runtime string, opts RunOpts, stdout, stderr io.Writer, tty bool) error {
	cmd := exec.CommandContext(ctx, runtime, containerArgs(runtime, opts, tty)...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// containerArgs builds the `docker`/`podman` run argument vector. It is shared
// between the buffered and streaming paths so container semantics cannot drift.
func containerArgs(_ string, opts RunOpts, tty bool) []string {
	image := ImageName()

	dockerArgs := []string{"run", "--rm", "-i"}

	// Attach a tty when an interactive terminal is present.
	if tty {
		dockerArgs = append(dockerArgs, "-t")
	}

	// Capability for network-level tools.
	if opts.NeedsRoot {
		dockerArgs = append(dockerArgs, "--cap-add", "NET_ADMIN", "--cap-add", "NET_RAW")
	}

	// Auto-detect file paths in args and bind-mount them.
	mounts, remapped := detectMounts(opts.Args)
	for _, m := range mounts {
		dockerArgs = append(dockerArgs, "-v", m)
	}

	// Working directory.
	if opts.WorkDir != "" {
		dockerArgs = append(dockerArgs, "-w", opts.WorkDir)
	}

	dockerArgs = append(dockerArgs, image, opts.Binary)
	dockerArgs = append(dockerArgs, remapped...)
	return dockerArgs
}

func detectRuntime() (string, error) {
	// Respect explicit override.
	if rt := os.Getenv(EnvRuntime); rt != "" {
		if _, err := exec.LookPath(rt); err == nil {
			return rt, nil
		}
		return "", fmt.Errorf("COLDCASE_RUNTIME=%q not found on PATH", rt)
	}
	for _, rt := range []string{"docker", "podman"} {
		if _, err := exec.LookPath(rt); err == nil {
			return rt, nil
		}
	}
	return "", fmt.Errorf("neither docker nor podman found on PATH")
}

func isTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
