package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type DidierStevensTool struct {
	name        string
	scriptPath  string
	description string
}

func (d *DidierStevensTool) Name() string {
	return d.name
}

func (d *DidierStevensTool) Description() string {
	return d.description
}

func (d *DidierStevensTool) Run(args []string) error {
	if !checkToolInstalled("python3") {
		return fmt.Errorf("python3 is required but not installed")
	}

	if _, err := os.Stat(d.scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script %s not found", d.scriptPath)
	}

	cmdArgs := []string{d.scriptPath}
	cmdArgs = append(cmdArgs, args...)

	return executeCommand("python3", cmdArgs...)
}

func addDidierStevensCommands() {
	suitePath := "./DidierStevensSuite"

	tools := []*DidierStevensTool{
		{
			name:        "1768",
			scriptPath:  filepath.Join(suitePath, "1768.py"),
			description: "Analyze 1768 PDF files",
		},
		{
			name:        "amsiscan",
			scriptPath:  filepath.Join(suitePath, "amsiscan.py"),
			description: "Scan AMSI cache",
		},
		{
			name:        "pdf-parser",
			scriptPath:  filepath.Join(suitePath, "pdf-parser.py"),
			description: "Parse PDF documents for analysis",
		},
		{
			name:        "pdfid",
			scriptPath:  filepath.Join(suitePath, "pdfid.py"),
			description: "Test PDF files for malicious content",
		},
		{
			name:        "oledump",
			scriptPath:  filepath.Join(suitePath, "oledump.py"),
			description: "Analyze OLE files (Office documents)",
		},
		{
			name:        "pecheck",
			scriptPath:  filepath.Join(suitePath, "pecheck.py"),
			description: "Display PE file information",
		},
		{
			name:        "base64dump",
			scriptPath:  filepath.Join(suitePath, "base64dump.py"),
			description: "Extract base64 strings from files",
		},
		{
			name:        "emldump",
			scriptPath:  filepath.Join(suitePath, "emldump.py"),
			description: "Extract and analyze EML email files",
		},
		{
			name:        "jpegdump",
			scriptPath:  filepath.Join(suitePath, "jpegdump.py"),
			description: "Analyze JPEG file structure and metadata",
		},
		{
			name:        "hash",
			scriptPath:  filepath.Join(suitePath, "hash.py"),
			description: "Calculate file hashes with multiple algorithms",
		},
		{
			name:        "cut-bytes",
			scriptPath:  filepath.Join(suitePath, "cut-bytes.py"),
			description: "Extract specific byte ranges from files",
		},
		{
			name:        "find-file-in-file",
			scriptPath:  filepath.Join(suitePath, "find-file-in-file.py"),
			description: "Find embedded files within other files",
		},
		{
			name:        "byte-stats",
			scriptPath:  filepath.Join(suitePath, "byte-stats.py"),
			description: "Calculate byte distribution statistics",
		},
		{
			name:        "extractscripts",
			scriptPath:  filepath.Join(suitePath, "extractscripts.py"),
			description: "Extract embedded scripts from files",
		},
		{
			name:        "cs-parse-traffic",
			scriptPath:  filepath.Join(suitePath, "cs-parse-traffic.py"),
			description: "Parse Cobalt Strike traffic",
		},
	}

	for _, tool := range tools {
		cmd := &cobra.Command{
			Use:   tool.name,
			Short: tool.description,
			Run: func(cmd *cobra.Command, args []string) {
				if err := tool.Run(args); err != nil {
					fmt.Printf("Error running %s: %v\n", tool.name, err)
					os.Exit(1)
				}
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
