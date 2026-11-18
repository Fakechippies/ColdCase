//go:build unix

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// detachProcess starts the child in its own session so it survives the parent
// terminal closing (no controlling tty, no SIGHUP propagation).
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// pidAlive reports whether a process with this pid exists.
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

// processIsColdcase verifies the pid actually belongs to a ColdCase binary
// before it is ever signalled. Best effort: on kernels without /proc we
// decline to verify (the caller decides how strict to be).
func processIsColdcase(pid int) bool {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return false
	}
	// cmdline is NUL-separated; the first element is the executable path.
	for _, part := range strings.Split(string(data), "\x00") {
		if part == "" {
			continue
		}
		base := filepath.Base(part)
		return base == "coldcase" || strings.HasPrefix(base, "coldcase")
	}
	return false
}

func terminateProcess(pid int) {
	_ = syscall.Kill(pid, syscall.SIGTERM)
}

func killProcess(pid int) {
	_ = syscall.Kill(pid, syscall.SIGKILL)
}
