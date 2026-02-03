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
	fmt.Println("  1768             - Analyze 1768 PDF files")
	fmt.Println("  amsiscan         - Scan AMSI cache")
	fmt.Println("  pdf-parser       - Parse PDF documents for analysis")
	fmt.Println("  pdfid            - Test PDF files for malicious content")
	fmt.Println("  oledump          - Analyze OLE files (Office documents)")
	fmt.Println("  pecheck          - Display PE file information")
	fmt.Println("  base64dump       - Extract base64 strings from files")
	fmt.Println("  emldump          - Extract and analyze EML email files")
	fmt.Println("  jpegdump         - Analyze JPEG file structure and metadata")
	fmt.Println("  hash             - Calculate file hashes with multiple algorithms")
	fmt.Println("  cut-bytes        - Extract specific byte ranges from files")
	fmt.Println("  find-file-in-file- Find embedded files within other files")
	fmt.Println("  byte-stats       - Calculate byte distribution statistics")
	fmt.Println("  extractscripts   - Extract embedded scripts from files")
	fmt.Println("  cs-parse-traffic - Parse Cobalt Strike traffic")

	fmt.Println("\nGeneral Tools:")
	fmt.Println("  exif     - Extract metadata from files using ExifTool")
	fmt.Println("  binwalk  - Analyze and extract firmware images")

	fmt.Println("\nSleuth Kit:")
	fmt.Println("  fls        - List directory and file entries")
	fmt.Println("  fsstat     - Display file system details")
	fmt.Println("  istat      - Display image metadata")
	fmt.Println("  jls        - List journal entries")
	fmt.Println("  tsk_loaddb - Load image into database")

	fmt.Println("\nUtility Commands:")
	fmt.Println("  list     - Show this list of available tools")
	fmt.Println("  check    - Check which tools are installed")
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
