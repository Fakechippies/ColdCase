package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Volatility3Tool struct {
	name        string
	description string
	command     string
}

func (v Volatility3Tool) Name() string {
	return v.name
}

func (v Volatility3Tool) Description() string {
	return v.description
}

func (v Volatility3Tool) Run(args []string) error {
	volPath := filepath.Join("volatility3", "vol.py")
	if !checkToolInstalled("python3") {
		return fmt.Errorf("python3 is required but not installed")
	}

	if _, err := os.Stat(volPath); os.IsNotExist(err) {
		return fmt.Errorf("volatility3 not found at %s", volPath)
	}

	cmdArgs := []string{volPath}
	if v.command != "" {
		cmdArgs = append(cmdArgs, v.command)
	}
	cmdArgs = append(cmdArgs, args...)

	return executeCommand("python3", cmdArgs...)
}

var volatility3Tools = []Volatility3Tool{
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
}

func addVolatility3Commands() {
	for _, tool := range volatility3Tools {
		cmd := &cobra.Command{
			Use:   tool.Name(),
			Short: tool.Description(),
			Long:  tool.Description() + " - Volatility3 memory forensics",
			Run: func(cmd *cobra.Command, args []string) {
				tool := cmd.Annotations["tool"]
				for _, t := range volatility3Tools {
					if t.Name() == tool {
						if err := t.Run(args); err != nil {
							fmt.Printf("Error running %s: %v\n", tool, err)
							os.Exit(1)
						}
						break
					}
				}
			},
		}
		cmd.Annotations = map[string]string{"tool": tool.Name()}

		// Add common flags for memory analysis
		if strings.HasPrefix(tool.Name(), "windows.") ||
			strings.HasPrefix(tool.Name(), "linux.") ||
			strings.HasPrefix(tool.Name(), "mac.") {
			cmd.Flags().StringP("file", "f", "", "Memory image file to analyze")
		}

		rootCmd.AddCommand(cmd)
	}
}

func checkVolatility3Dependencies() map[string]bool {
	deps := make(map[string]bool)
	deps["python3"] = checkToolInstalled("python3")

	// Check if volatility3 directory exists and contains vol.py
	volPath := filepath.Join("volatility3", "vol.py")
	if _, err := os.Stat(volPath); err == nil {
		deps["volatility3"] = true
	} else {
		deps["volatility3"] = false
	}

	return deps
}
