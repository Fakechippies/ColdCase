// Package sleuthkit integrates The Sleuth Kit filesystem analysis tools.
package sleuthkit

import (
	"fmt"

	"coldcase/pkg/tools"
)

// SleuthKitTool wraps a single Sleuth Kit binary.
type SleuthKitTool struct {
	tool string
}

// New returns a SleuthKitTool for the given binary name (e.g. "fls").
func New(tool string) *SleuthKitTool { return &SleuthKitTool{tool: tool} }

// Name returns the binary name used as the CLI command.
func (s *SleuthKitTool) Name() string { return s.tool }

// Description returns a human-readable description of the tool.
func (s *SleuthKitTool) Description() string {
	descriptions := map[string]string{
		"fls":        "List directory and file entries",
		"fsstat":     "Display file system details",
		"istat":      "Display image metadata",
		"jls":        "List journal entries",
		"tsk_loaddb": "Load image into database",
	}
	if desc, ok := descriptions[s.tool]; ok {
		return desc
	}
	return fmt.Sprintf("Run %s from Sleuth Kit", s.tool)
}

// Run invokes the Sleuth Kit binary with the provided arguments.
func (s *SleuthKitTool) Run(args []string) error {
	if !tools.CheckToolInstalled(s.tool) {
		return fmt.Errorf("%s is not installed (install sleuthkit)", s.tool)
	}
	return tools.ExecuteCommand(s.tool, args...)
}

// Tools returns the standard set of Sleuth Kit tools.
func Tools() []*SleuthKitTool {
	names := []string{"fls", "fsstat", "istat", "jls", "tsk_loaddb"}
	result := make([]*SleuthKitTool, len(names))
	for i, n := range names {
		result[i] = New(n)
	}
	return result
}
