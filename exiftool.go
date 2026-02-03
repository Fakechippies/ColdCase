package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type ExifTool struct{}

func (e *ExifTool) Name() string {
	return "exiftool"
}

func (e *ExifTool) Description() string {
	return "Extract metadata from files using ExifTool"
}

func (e *ExifTool) Run(args []string) error {
	if !checkToolInstalled("exiftool") {
		return fmt.Errorf("exiftool is not installed")
	}

	return executeCommand("exiftool", args...)
}

func addExifToolCommand() {
	exifTool := &ExifTool{}

	cmd := &cobra.Command{
		Use:   "exif",
		Short: exifTool.Description(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := exifTool.Run(args); err != nil {
				fmt.Printf("Error running exiftool: %v\n", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(cmd)
}
