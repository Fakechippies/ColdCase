// Package exiftool integrates ExifTool for metadata extraction.
package exiftool

import (
	"fmt"

	"coldcase/pkg/tools"
)

// ExifTool wraps the exiftool binary.
type ExifTool struct{}

// New returns an ExifTool instance.
func New() *ExifTool { return &ExifTool{} }

// Name returns the CLI command name ("exif").
func (e *ExifTool) Name() string { return "exif" }

// Description returns a short description.
func (e *ExifTool) Description() string {
	return "Extract metadata from files using ExifTool"
}

// Run invokes exiftool with the provided arguments.
func (e *ExifTool) Run(args []string) error {
	if !tools.CheckToolInstalled("exiftool") {
		return fmt.Errorf("exiftool is not installed")
	}
	return tools.ExecuteCommand("exiftool", args...)
}
