package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available forensics tools",
		Run:   listTools,
	}
	rootCmd.AddCommand(listCmd)

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check which tools are installed",
		Run:   checkTools,
	}
	rootCmd.AddCommand(checkCmd)
}

func listTools(cmd *cobra.Command, args []string) {
	fmt.Println("Available Forensics Tools:")

	fmt.Println("\nDidierStevens Suite:")
	tools := []struct {
		name, desc string
	}{
		{"1768", "Analyze 1768 PDF files"},
		{"amsiscan", "Scan AMSI cache"},
		{"pdf-parser", "Parse PDF documents for analysis"},
		{"pdfid", "Test PDF files for malicious content"},
		{"oledump", "Analyze OLE files (Office documents)"},
		{"pecheck", "Display PE file information"},
		{"base64dump", "Extract base64 strings from files"},
		{"emldump", "Extract and analyze EML email files"},
		{"jpegdump", "Analyze JPEG file structure and metadata"},
		{"hash", "Calculate file hashes with multiple algorithms"},
		{"cut-bytes", "Extract specific byte ranges from files"},
		{"find-file-in-file", "Find embedded files within other files"},
		{"byte-stats", "Calculate byte distribution statistics"},
		{"extractscripts", "Extract embedded scripts from files"},
		{"cs-parse-traffic", "Parse Cobalt Strike traffic"},
	}

	for _, tool := range tools {
		fmt.Printf("  %s - %s\n", tool.name, tool.desc)
	}

	fmt.Println("\nGeneral Tools:")
	fmt.Printf("  %s - %s\n", "exif", "Extract metadata from files using ExifTool")
	fmt.Printf("  %s - %s\n", "binwalk", "Analyze and extract firmware images")

	fmt.Println("\nSleuth Kit:")
	sleuthTools := []struct {
		name, desc string
	}{
		{"fls", "List directory and file entries"},
		{"fsstat", "Display file system details"},
		{"istat", "Display image metadata"},
		{"jls", "List journal entries"},
		{"tsk_loaddb", "Load image into database"},
	}

	for _, tool := range sleuthTools {
		fmt.Printf("  %s - %s\n", tool.name, tool.desc)
	}

	fmt.Println("\nUtility Commands:")
	fmt.Printf("  %s - %s\n", "list", "Show this list of available tools")
	fmt.Printf("  %s - %s\n", "check", "Check which tools are installed")
}

func checkTools(cmd *cobra.Command, args []string) {
	fmt.Println("Checking installed tools...")

	tools := []string{
		"python3",
		"exiftool",
		"binwalk",
		"fls",
		"fsstat",
		"istat",
		"jls",
		"tsk_loaddb",
	}

	for _, tool := range tools {
		if checkToolInstalled(tool) {
			fmt.Printf("✓ %s - installed\n", tool)
		} else {
			fmt.Printf("✗ %s - not found\n", tool)
		}
	}

	if _, err := os.Stat("./DidierStevensSuite"); err == nil {
		fmt.Println("✓ DidierStevensSuite - found")
	} else {
		fmt.Println("✗ DidierStevensSuite - not found")
	}
}
