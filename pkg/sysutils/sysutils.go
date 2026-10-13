// Package sysutils provides thin wrappers around standard system utilities
// commonly used during binary and filesystem analysis.
package sysutils

import (
	"coldcase/pkg/runner"
)

// SysUtil wraps a system utility binary.
type SysUtil struct {
	name string
	desc string
}

func (s *SysUtil) Name() string        { return s.name }
func (s *SysUtil) Description() string { return s.desc }
func (s *SysUtil) RunOpts(args []string) runner.RunOpts {
	return runner.RunOpts{Binary: s.name, Args: args}
}
func (s *SysUtil) Run(args []string) error {
	return runner.Run(s.RunOpts(args))
}

// Tools returns all system utility wrappers.
func Tools() []*SysUtil {
	return []*SysUtil{
		{"xxd", "Hex dump of a file or stdin"},
		{"objdump", "Display information from object and binary files"},
		{"readelf", "Display information about ELF binaries"},
		{"nm", "List symbols from object files"},
		{"file", "Determine file type via magic number detection"},
		{"ldd", "List dynamic dependencies of executable files"},
	}
}
