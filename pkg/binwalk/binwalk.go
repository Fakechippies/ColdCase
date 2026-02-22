// Package binwalk integrates the Binwalk firmware analysis tool.
package binwalk

import "coldcase/pkg/runner"

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
	return runner.Run(runner.RunOpts{
		Binary: "binwalk",
		Args:   args,
	})
}
