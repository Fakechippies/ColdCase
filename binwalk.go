package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type BinwalkTool struct{}

func (b *BinwalkTool) Name() string {
	return "binwalk"
}

func (b *BinwalkTool) Description() string {
	return "Analyze and extract firmware images using Binwalk"
}

func (b *BinwalkTool) Run(args []string) error {
	if !checkToolInstalled("binwalk") {
		return fmt.Errorf("binwalk is not installed")
	}

	return executeCommand("binwalk", args...)
}

func addBinwalkCommand() {
	binwalkTool := &BinwalkTool{}

	cmd := &cobra.Command{
		Use:   "binwalk",
		Short: binwalkTool.Description(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := binwalkTool.Run(args); err != nil {
				fmt.Printf("Error running binwalk: %v\n", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(cmd)
}
