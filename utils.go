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

	fmt.Println("\nVolatility3 Memory Forensics:")
	volTools := []struct {
		name, desc string
	}{
		{"vol", "Main volatility3 framework"},
		{"volshell", "Interactive volatility shell"},
		{"windows.pslist", "List Windows processes"},
		{"windows.pstree", "Show Windows process tree"},
		{"windows.dlllist", "List Windows process DLLs"},
		{"windows.handles", "List Windows handles"},
		{"windows.cmdline", "Show Windows process command lines"},
		{"windows.envars", "Display Windows process environment variables"},
		{"windows.filescan", "Scan for Windows file objects"},
		{"windows.modules", "List Windows kernel modules"},
		{"windows.driverscan", "Scan for Windows driver objects"},
		{"windows.callbacks", "List Windows registered callbacks"},
		{"windows.services", "List Windows services"},
		{"windows.registry", "Windows registry analysis"},
		{"windows.hashdump", "Dump Windows password hashes"},
		{"linux.pslist", "List Linux processes"},
		{"linux.pstree", "Show Linux process tree"},
		{"linux.bash", "Recover Linux bash history"},
		{"linux.proc_maps", "Linux process memory maps"},
		{"mac.pslist", "List macOS processes"},
		{"mac.pstree", "Show macOS process tree"},
		{"info", "Display memory image information"},
	}

	for _, tool := range volTools {
		fmt.Printf("  %s - %s\n", tool.name, tool.desc)
	}

	fmt.Println("\nUtility Commands:")
	fmt.Printf("  %s - %s\n", "list", "Show this list of available tools")
	fmt.Printf("  %s - %s\n", "check", "Check which tools are installed")
	fmt.Printf("  %s - %s\n", "install", "Install project dependencies")
	fmt.Printf("  %s - %s\n", "deps", "Manage project dependencies")
	fmt.Printf("  %s - %s\n", "platform", "Show platform-specific setup information")
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

	volDeps := checkVolatility3Dependencies()
	for dep, installed := range volDeps {
		if installed {
			fmt.Printf("✓ %s - available\n", dep)
		} else {
			fmt.Printf("✗ %s - not found\n", dep)
		}
	}

	// Check for uv
	if checkToolInstalled("uv") {
		fmt.Println("✓ uv - available for fast dependency management")
	} else {
		fmt.Println("✗ uv - not found (optional for faster Python dependency management)")
	}

	// Show detected package manager
	if checkToolInstalled("apt") {
		fmt.Println("✓ apt - Ubuntu/Debian package manager detected")
	} else if checkToolInstalled("dnf") {
		fmt.Println("✓ dnf - Fedora/RHEL package manager detected")
	} else if checkToolInstalled("yum") {
		fmt.Println("✓ yum - RHEL/CentOS package manager detected")
	} else if checkToolInstalled("pacman") {
		fmt.Println("✓ pacman - Arch Linux package manager detected")
	} else if checkToolInstalled("brew") {
		fmt.Println("✓ brew - macOS package manager detected")
	} else if checkToolInstalled("choco") {
		fmt.Println("✓ choco - Chocolatey package manager detected")
	} else if checkToolInstalled("winget") {
		fmt.Println("✓ winget - Windows package manager detected")
	} else {
		fmt.Println("⚠️  No supported package manager detected")
	}
}
