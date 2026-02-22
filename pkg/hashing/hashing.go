// Package hashing integrates advanced hashing and integrity verification tools.
package hashing

import (
	"coldcase/pkg/runner"
)

// HashTool wraps a hashing binary.
type HashTool struct {
	name string
	desc string
}

func (h *HashTool) Name() string        { return h.name }
func (h *HashTool) Description() string { return h.desc }
func (h *HashTool) Run(args []string) error {
	return runner.Run(runner.RunOpts{Binary: h.name, Args: args})
}

// Tools returns all hashing tools.
func Tools() []*HashTool {
	return []*HashTool{
		{"md5deep", "Recursive MD5 hashing of files and directories"},
		{"hashdeep", "Audit-mode hashing with verification against a known-good set"},
		{"ssdeep", "Fuzzy (context-triggered piecewise) hashing for similarity analysis"},
		{"tlsh", "Trend Micro Locality Sensitive Hash — similarity scoring for malware"},
	}
}
