// Package windows integrates Windows artifact analysis tools into ColdCase.
// These tools run on Linux hosts and operate on offline Windows disk images/hives.
package windows

import (
	"coldcase/pkg/runner"
)

// WindowsTool wraps a Windows artifact analysis binary.
type WindowsTool struct {
	name string
	bin  string
	desc string
}

func (w *WindowsTool) Name() string        { return w.name }
func (w *WindowsTool) Description() string { return w.desc }
func (w *WindowsTool) Run(args []string) error {
	bin := w.bin
	if bin == "" {
		bin = w.name
	}
	return runner.Run(runner.RunOpts{Binary: bin, Args: args})
}

// Tools returns all Windows artifact analysis tools.
func Tools() []*WindowsTool {
	return []*WindowsTool{
		{"regripper", "rip.pl", "RegRipper — comprehensive Windows registry hive analysis (Perl)"},
		{"regrippy", "regrippy", "RegRippy — Python framework for Windows registry forensics"},
		{"analyzeMFT", "analyzeMFT", "Parse and analyze the NTFS Master File Table"},
		{"ntfsls", "", "List files and directories in an NTFS image (ntfs-3g)"},
		{"ntfscat", "", "Extract files from an NTFS image (ntfs-3g)"},
		{"indxparse", "INDXParse", "NTFS $I30 index parser"},
		{"registry-dump", "python-registry", "Read/walk Windows Registry hives"},
	}
}
