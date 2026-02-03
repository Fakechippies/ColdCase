package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type SleuthKitTool struct {
	tool string
}

func (s *SleuthKitTool) Name() string {
	return s.tool
}

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

func (s *SleuthKitTool) Run(args []string) error {
	if !checkToolInstalled(s.tool) {
		return fmt.Errorf("%s is not installed", s.tool)
	}

	return executeCommand(s.tool, args...)
}

func addSleuthKitCommands() {
	tools := []string{"fls", "fsstat", "istat", "jls", "tsk_loaddb"}

	for _, tool := range tools {
		sleuthTool := &SleuthKitTool{tool: tool}

		cmd := &cobra.Command{
			Use:   tool,
			Short: sleuthTool.Description(),
			Run: func(cmd *cobra.Command, args []string) {
				if err := sleuthTool.Run(args); err != nil {
					fmt.Printf("Error running %s: %v\n", sleuthTool.Name(), err)
					os.Exit(1)
				}
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
