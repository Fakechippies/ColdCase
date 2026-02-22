// Package runner provides the hybrid execution engine for ColdCase.
// It runs tools natively if available on PATH; otherwise it transparently
// proxies the invocation through a Docker or Podman container.
package runner

import (
	"fmt"
	"os"
	"os/exec"

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
	if tools.CheckToolInstalled(opts.Binary) {
		return runNative(opts)
	}
	rt, err := detectRuntime()
	if err != nil {
		return fmt.Errorf("'%s' not found on PATH and no container runtime available: %w", opts.Binary, err)
	}
	return runInContainer(rt, opts)
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

func runNative(opts RunOpts) error {
	cmd := exec.Command(opts.Binary, opts.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}
	return cmd.Run()
}

func runInContainer(runtime string, opts RunOpts) error {
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
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
