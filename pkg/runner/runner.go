// Package runner provides the hybrid execution engine for ColdCase.
// It runs tools natively if available on PATH; otherwise it transparently
// proxies the invocation through a Docker or Podman container.
package runner

import (
	"fmt"
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
}

// Run executes opts.Binary with opts.Args.
// Native execution is attempted first. If the binary is not found on PATH,
// Run falls back to running the command inside a container.
func Run(opts RunOpts) error {
	sID := session.GetActiveSessionID()
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
	var runErr error
	var output []byte

	if tools.CheckToolInstalled(opts.Binary) {
		output, runErr = runNative(opts)
	} else {
		rt, err := detectRuntime()
		if err != nil {
			return fmt.Errorf("'%s' not found on PATH and no container runtime available: %w", opts.Binary, err)
		}
		output, runErr = runInContainer(rt, opts)
	}

	if logger != nil && sess != nil {
		duration := time.Since(start)

		// Map input files
		var inputFiles []session.FileMetadata
		for _, arg := range opts.Args {
			if _, err := os.Stat(arg); err == nil {
				meta, err := logger.HashInputFile(arg)
				if err == nil {
					inputFiles = append(inputFiles, meta)
				}
			}
		}

		// Save output
		idx := len(sess.Commands) + 1
		outPath, _ := logger.SaveOutput(idx, opts.Binary, output)

		preview := ""
		if len(output) > 500 {
			preview = string(output[:500]) + "..."
		} else {
			preview = string(output)
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
			ExitCode:         0, // Simplified for now
			DurationMS:       duration.Milliseconds(),
			OutputPreview:    preview,
			OutputFile:       outPath,
		}
		_ = logger.LogCommand(entry)
		fmt.Fprintf(os.Stderr, "\n[*] Signed entry logged to session: %s\n", sID)
	}

	return runErr
}

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

// ─── internal ─────────────────────────────────────────────────────────────────

func runNative(opts RunOpts) ([]byte, error) {
	cmd := exec.Command(opts.Binary, opts.Args...)
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	// Capture output while still showing it to the user
	output, err := cmd.CombinedOutput()
	fmt.Print(string(output))
	return output, err
}

func runInContainer(runtime string, opts RunOpts) ([]byte, error) {
	image := ImageName()

	dockerArgs := []string{"run", "--rm", "-i"}

	// Attach a tty when stdin is a terminal.
	if isTTY() {
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

	cmd := exec.Command(runtime, dockerArgs...)
	output, err := cmd.CombinedOutput()
	fmt.Print(string(output))
	return output, err
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
