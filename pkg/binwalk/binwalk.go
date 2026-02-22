// Package binwalk integrates the Binwalk firmware analysis tool.
package binwalk

import (
	"fmt"

	"coldcase/pkg/tools"
)

// BinwalkTool wraps the binwalk binary.
type BinwalkTool struct{}

// New returns a BinwalkTool instance.
func New() *BinwalkTool { return &BinwalkTool{} }

// Name returns the CLI command name ("binwalk").
func (b *BinwalkTool) Name() string { return "binwalk" }

// Description returns a short description.
func (b *BinwalkTool) Description() string {
	return "Analyze and extract firmware images using Binwalk"
}

// Run invokes binwalk with the provided arguments.
func (b *BinwalkTool) Run(args []string) error {
	if !tools.CheckToolInstalled("binwalk") {
		return fmt.Errorf("binwalk is not installed")
	}
	return tools.ExecuteCommand("binwalk", args...)
}
