//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// Creation flag values from the Win32 API (not exported by the syscall
// package on all Go versions).
const (
	detachedProcess       = 0x00000008
	createNewProcessGroup = 0x00000200
)

// detachProcess starts the child without inheriting the console window, so it
// survives the parent terminal closing.
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: detachedProcess | createNewProcessGroup,
		HideWindow:    true,
	}
}

// pidAlive reports whether a process with this pid exists. Uses tasklist as a
// best effort; Windows has no /proc.
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	out, err := exec.Command("tasklist", "/FI",
		fmt.Sprintf("PID eq %d", pid), "/NH", "/FO", "CSV").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), fmt.Sprintf("\"%d\"", pid))
}

// processIsColdcase verifies the pid actually belongs to a ColdCase binary.
func processIsColdcase(pid int) bool {
	out, err := exec.Command("tasklist", "/FI",
		fmt.Sprintf("PID eq %d", pid), "/NH", "/FO", "CSV").Output()
	if err != nil {
		return false
	}
	fields := strings.Split(string(out), ",")
	if len(fields) == 0 {
		return false
	}
	name := strings.ToLower(strings.TrimSpace(fields[0]))
	return strings.Contains(name, "coldcase")
}

func terminateProcess(pid int) {
	_ = taskkill(pid, false)
}

func killProcess(pid int) {
	_ = taskkill(pid, true)
}

// taskkill uses the Windows utility; os.Process.Signal is not implemented on
// Windows.
func taskkill(pid int, force bool) error {
	args := []string{"/PID", fmt.Sprint(pid)}
	if force {
		args = append(args, "/F")
	}
	return exec.Command("taskkill", args...).Run()
}
