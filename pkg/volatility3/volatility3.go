// Package volatility3 integrates the Volatility3 memory forensics framework.
// It supports Windows, Linux, and macOS memory image analysis.
package volatility3

import (
	"fmt"
	"os"
	"path/filepath"

	"coldcase/pkg/tools"
)

// Volatility3Tool wraps a single Volatility3 plugin or top-level command.
type Volatility3Tool struct {
	name        string
	description string
	// command is the volatility3 plugin name (e.g. "windows.pslist").
	// An empty command means the raw vol.py entry point is used.
	command string
}

// Name returns the CLI command name.
func (v Volatility3Tool) Name() string { return v.name }

// Description returns a short description of this plugin.
func (v Volatility3Tool) Description() string { return v.description }

// Command returns the volatility3 plugin/sub-command name.
func (v Volatility3Tool) Command() string { return v.command }

// Run executes vol.py with the configured plugin and any extra arguments.
// volDir is the path to the directory containing vol.py (default: "./volatility3").
func (v Volatility3Tool) Run(args []string) error {
	return RunWithVolDir("volatility3", v.command, args)
}

// RunWithVolDir runs vol.py located at volDir/vol.py with the given plugin and args.
func RunWithVolDir(volDir, command string, args []string) error {
	volPath := filepath.Join(volDir, "vol.py")
	if !tools.CheckToolInstalled("python3") {
		return fmt.Errorf("python3 is required but not installed")
	}
	if _, err := os.Stat(volPath); os.IsNotExist(err) {
		return fmt.Errorf("volatility3 not found at %s", volPath)
	}
	cmdArgs := []string{volPath}
	if command != "" {
		cmdArgs = append(cmdArgs, command)
	}
	cmdArgs = append(cmdArgs, args...)
	return tools.ExecuteCommand("python3", cmdArgs...)
}

// Tools returns the full list of pre-defined Volatility3 tools.
func Tools() []Volatility3Tool {
	return []Volatility3Tool{
		{"vol", "Run volatility3 memory forensics framework", ""},
		{"volshell", "Interactive volatility shell", "volshell"},
		{"windows.pslist", "List running processes (Windows memory)", "windows.pslist"},
		{"windows.pstree", "Show process tree (Windows memory)", "windows.pstree"},
		{"windows.dlllist", "List DLLs for processes (Windows memory)", "windows.dlllist"},
		{"windows.handles", "List handles (Windows memory)", "windows.handles"},
		{"windows.cmdline", "Display process command lines (Windows memory)", "windows.cmdline"},
		{"windows.envars", "Display process environment variables (Windows memory)", "windows.envars"},
		{"windows.filescan", "Scan for file objects (Windows memory)", "windows.filescan"},
		{"windows.modules", "List loaded kernel modules (Windows memory)", "windows.modules"},
		{"windows.driverscan", "Scan for driver objects (Windows memory)", "windows.driverscan"},
		{"windows.callbacks", "List registered callbacks (Windows memory)", "windows.callbacks"},
		{"windows.services", "List services (Windows memory)", "windows.services"},
		{"windows.registry", "Registry analysis (Windows memory)", "windows.registry"},
		{"windows.hashdump", "Dump password hashes (Windows memory)", "windows.hashdump"},
		{"linux.pslist", "List running processes (Linux memory)", "linux.pslist"},
		{"linux.pstree", "Show process tree (Linux memory)", "linux.pstree"},
		{"linux.bash", "Recover bash history (Linux memory)", "linux.bash"},
		{"linux.proc_maps", "Process memory maps (Linux memory)", "linux.proc_maps"},
		{"mac.pslist", "List running processes (macOS memory)", "mac.pslist"},
		{"mac.pstree", "Show process tree (macOS memory)", "mac.pstree"},
		{"info", "Display information about a memory image", "info"},
		// Expanded plugins
		{"windows.malfind", "Detect injected code and memory anomalies (Windows memory)", "windows.malfind"},
		{"windows.mutantscan", "Scan for mutex objects — common malware indicators (Windows memory)", "windows.mutantscan"},
		{"windows.ssdt", "System Service Descriptor Table analysis (Windows memory)", "windows.ssdt"},
		{"windows.getsids", "Extract Security Identifiers for processes (Windows memory)", "windows.getsids"},
		{"windows.privs", "List process privileges (Windows memory)", "windows.privs"},
		{"windows.vadinfo", "Virtual Address Descriptor information (Windows memory)", "windows.vadinfo"},
		{"windows.dumpfiles", "Extract files cached in memory (Windows memory)", "windows.dumpfiles"},
		{"windows.mftscan", "Scan for MFT entries in memory (Windows memory)", "windows.mftscan"},
		{"linux.mount_info", "Linux mount point information (Linux memory)", "linux.mount_info"},
		{"mac.mount_info", "macOS mount point information (macOS memory)", "mac.mount_info"},
	}
}

// CheckDependencies returns a map of dependency name → installed status
// relevant to the Volatility3 package.
func CheckDependencies() map[string]bool {
	deps := map[string]bool{
		"python3": tools.CheckToolInstalled("python3"),
	}
	volPath := filepath.Join("volatility3", "vol.py")
	_, err := os.Stat(volPath)
	deps["volatility3"] = err == nil
	return deps
}
