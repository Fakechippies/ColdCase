// Package tools provides the core Tool interface and shared helpers
// used by all ColdCase sub-packages.
package tools

import (
	"os"
	"os/exec"
)

// Tool is the common interface every ColdCase integration implements.
// Implement this interface to expose a new forensics tool.
type Tool interface {
	// Name returns the CLI command name (e.g. "pdf-parser").
	Name() string
	// Description returns a short one-line description of the tool.
	Description() string
	// Run invokes the underlying tool with the given arguments.
	Run(args []string) error
}

// ExecuteCommand runs an external binary, streaming its stdout/stderr
// directly to the calling process.
func ExecuteCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CheckToolInstalled reports whether the named binary is available on PATH.
func CheckToolInstalled(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
