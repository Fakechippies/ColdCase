//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// detachProcess starts the child without inheriting the console window, so it
// survives the parent terminal closing.
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.DETACHED_PROCESS | syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
}

// pidAlive reports whether a process with this pid exists. Uses the pid file
// name match via tasklist as a best effort.
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
	name := strings.ToLower(strings.SplitN(string(out), ",", 2)[0])
	return strings.Contains(name, "coldcase")
}

func terminateProcess(pid int) {
	_ = sendWindowsSignal(pid, syscall.SIGTERM)
}

func killProcess(pid int) {
	_ = sendWindowsSignal(pid, syscall.SIGKILL)
}

// sendWindowsSignal uses taskkill; plain Signal() is not implemented for
// os.Process on Windows.
func sendWindowsSignal(pid int, _ os.Signal) error {
	force := ""
	if syscall.Signal(pid) == syscall.SIGKILL {
		force = "/F"
	}
	return exec.Command("taskkill", "/PID", fmt.Sprint(pid), force).Run()
}
