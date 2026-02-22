// Package carving integrates file carving and data recovery tools.
package carving

import (
	"coldcase/pkg/runner"
)

// CarvingTool wraps a file carving or data recovery binary.
type CarvingTool struct {
	name string
	desc string
	// bin is the actual binary name when it differs from the display name.
	bin string
}

func (c *CarvingTool) Name() string        { return c.name }
func (c *CarvingTool) Description() string { return c.desc }
func (c *CarvingTool) Run(args []string) error {
	bin := c.bin
	if bin == "" {
		bin = c.name
	}
	return runner.Run(runner.RunOpts{Binary: bin, Args: args})
}

// Tools returns all carving and recovery tools.
func Tools() []*CarvingTool {
	return []*CarvingTool{
		{"foremost", "File carving based on file headers and footers", ""},
		{"scalpel", "Fast file carving from disk images", ""},
		{"photorec", "File recovery from disk images and cameras", "photorec"},
		{"bulk-extractor", "Feature extraction (emails, URLs, credit cards) from disk images", "bulk_extractor"},
		{"testdisk", "Partition and boot sector recovery", ""},
		{"ddrescue", "Data recovery from damaged storage devices", "ddrescue"},
		{"safecopy", "Data recovery tool for damaged media", "safecopy"},
	}
}
