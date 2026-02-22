// Package didier integrates the DidierStevens Suite of Python-based
// forensics tools. Each tool is backed by a .py script found under
// the suite directory (default: ./DidierStevensSuite).
package didier

import (
	"path/filepath"

	"coldcase/pkg/runner"
)

// DidierStevensTool wraps a single DidierStevens Python script.
type DidierStevensTool struct {
	name        string
	scriptPath  string
	description string
}

// Name returns the CLI command name for this tool.
func (d *DidierStevensTool) Name() string { return d.name }

// Description returns a short description of this tool.
func (d *DidierStevensTool) Description() string { return d.description }

// Run executes the Python script with the given arguments.
// Returns an error if python3 is not installed or the script is missing.
func (d *DidierStevensTool) Run(args []string) error {
	cmdArgs := append([]string{d.scriptPath}, args...)
	return runner.Run(runner.RunOpts{
		Binary: "python3",
		Args:   cmdArgs,
	})
}

// ScriptPath returns the resolved path to the underlying .py script.
func (d *DidierStevensTool) ScriptPath() string { return d.scriptPath }

// Tools builds the full set of DidierStevens tools using suitePath as the
// directory that contains the .py scripts (e.g. "./DidierStevensSuite").
func Tools(suitePath string) []*DidierStevensTool {
	defs := []struct {
		name, script, desc string
	}{
		{"1768", "1768.py", "Analyze 1768 PDF files"},
		{"amsiscan", "amsiscan.py", "Scan AMSI cache"},
		{"pdf-parser", "pdf-parser.py", "Parse PDF documents for analysis"},
		{"pdfid", "pdfid.py", "Test PDF files for malicious content"},
		{"oledump", "oledump.py", "Analyze OLE files (Office documents)"},
		{"pecheck", "pecheck.py", "Display PE file information"},
		{"base64dump", "base64dump.py", "Extract base64 strings from files"},
		{"emldump", "emldump.py", "Extract and analyze EML email files"},
		{"jpegdump", "jpegdump.py", "Analyze JPEG file structure and metadata"},
		{"hash", "hash.py", "Calculate file hashes with multiple algorithms"},
		{"cut-bytes", "cut-bytes.py", "Extract specific byte ranges from files"},
		{"find-file-in-file", "find-file-in-file.py", "Find embedded files within other files"},
		{"byte-stats", "byte-stats.py", "Calculate byte distribution statistics"},
		{"extractscripts", "extractscripts.py", "Extract embedded scripts from files"},
		{"cs-parse-traffic", "cs-parse-traffic.py", "Parse Cobalt Strike traffic"},
	}

	result := make([]*DidierStevensTool, 0, len(defs))
	for _, d := range defs {
		result = append(result, &DidierStevensTool{
			name:        d.name,
			scriptPath:  filepath.Join(suitePath, d.script),
			description: d.desc,
		})
	}
	return result
}
