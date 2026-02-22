package main

import (
	"fmt"
	"os"

	"coldcase/pkg/carving"
	"coldcase/pkg/didier"
	"coldcase/pkg/hashing"
	"coldcase/pkg/malware"
	"coldcase/pkg/mobile"
	"coldcase/pkg/network"
	"coldcase/pkg/runner"
	"coldcase/pkg/sleuthkit"
	"coldcase/pkg/steg"
	"coldcase/pkg/sysutils"
	"coldcase/pkg/timeline"
	"coldcase/pkg/tools"
	vol3 "coldcase/pkg/volatility3"
	wintools "coldcase/pkg/windows"

	"github.com/spf13/cobra"
)

func listTools(cmd *cobra.Command, args []string) {
	printCategory("DidierStevens Suite", didier.Tools(defaultSuitePath))
	printCategory("Network Forensics", network.Tools())
	printCategory("File Carving & Recovery", carving.Tools())
	printCategory("Malware & Pattern Matching", malware.Tools())
	printCategory("Hashing & Verification", hashing.Tools())
	printCategory("Timeline & Log Analysis", getNonPlasoTimelineTools())
	fmt.Println("\nPlaso (log2timeline):")
	fmt.Printf("  %-26s - %s\n", "plaso parse", "Plaso — multi-source supertimeline generation")
	fmt.Printf("  %-26s - %s\n", "plaso sort", "Plaso — sort and filter Plaso storage files")
	fmt.Printf("  %-26s - %s\n", "plaso psteal", "Plaso — extract and output timeline in one step")
	fmt.Printf("  %-26s - %s\n", "plaso parsers --list", "List all available plaso parsers")

	printCategory("Mobile Forensics", mobile.Tools())
	printCategory("Windows Artifacts", wintools.Tools())
	printCategory("Steganography & Media", steg.Tools())
	printCategory("System Utilities", sysutils.Tools())
	printCategory("Sleuth Kit", sleuthkit.Tools())
	printCategory("Volatility3 Memory Forensics", vol3.Tools())

	fmt.Println("\nGeneral Tools:")
	fmt.Printf("  %-26s - %s\n", "exif", "Extract metadata from files using ExifTool")
	fmt.Printf("  %-26s - %s\n", "binwalk", "Analyze and extract firmware images")

	fmt.Println("\nUtility Commands:")
	for _, u := range []struct{ n, d string }{
		{"list", "Show this list of available tools"},
		{"check", "Check which tools are installed"},
		{"install", "Install project dependencies"},
		{"deps", "Manage project dependencies"},
		{"platform", "Show platform-specific setup information"},
		{"container status", "Show container runtime and image status"},
		{"container build", "Build the ColdCase container image"},
		{"container pull", "Pull the ColdCase container image"},
		{"container shell", "Open an interactive shell in the container"},
	} {
		fmt.Printf("  %-26s - %s\n", u.n, u.d)
	}
}

type namedTool interface {
	Name() string
	Description() string
}

func printCategory[T namedTool](title string, ts []T) {
	fmt.Printf("\n%s:\n", title)
	for _, t := range ts {
		fmt.Printf("  %-26s - %s\n", t.Name(), t.Description())
	}
}

func getNonPlasoTimelineTools() []namedTool {
	var results []namedTool
	for _, t := range timeline.Tools() {
		switch t.Name() {
		case "log2timeline", "psort", "psteal":
			continue
		default:
			results = append(results, t)
		}
	}
	return results
}

func checkTools(cmd *cobra.Command, args []string) {
	fmt.Println("Checking installed tools...")

	binaries := []string{
		"python3", "exiftool", "binwalk",
		"fls", "fsstat", "istat", "jls", "tsk_loaddb",
		"tshark", "tcpdump", "zeek", "ngrep", "tcpflow", "pcapfix",
		"foremost", "scalpel", "photorec", "bulk_extractor", "testdisk", "ddrescue",
		"yara", "floss", "strings", "capa",
		"md5deep", "hashdeep", "ssdeep", "tlsh",
		"log2timeline", "hayabusa", "evtx_dump", "timeliner", "chainsaw",
		"adb", "ideviceinfo",
		"regripper", "regrippy", "analyzeMFT", "ntfsls", "ntfscat", "indxparse", "python-registry",
		"steghide", "zsteg", "mediainfo", "stegdetect",
		"xxd", "objdump", "readelf", "nm", "file", "ldd",
	}

	installed, missing := 0, 0
	for _, t := range binaries {
		if tools.CheckToolInstalled(t) {
			fmt.Printf("[*] %-20s installed\n", t)
			installed++
		} else {
			fmt.Printf("[x] %-20s not found\n", t)
			missing++
		}
	}

	fmt.Println()
	for dep, ok := range vol3.CheckDependencies() {
		if ok {
			fmt.Printf("[*] %-20s available\n", dep)
		} else {
			fmt.Printf("[x] %-20s not found\n", dep)
		}
	}

	if _, err := os.Stat("./DidierStevensSuite"); err == nil {
		fmt.Println("[*] DidierStevensSuite   found")
	} else {
		fmt.Println("[x] DidierStevensSuite   not found")
	}

	fmt.Println()

	// Container runtime
	rt := runner.DetectedRuntime()
	if rt != "" {
		fmt.Printf("[*] Container runtime    %s\n", rt)
		fmt.Printf("[*] Container image      %s\n", runner.ImageName())
		fmt.Println("    (missing tools will run via container)")
	} else {
		fmt.Println("[x] No container runtime (docker/podman not found)")
		fmt.Println("    Missing tools cannot fall back to container")
	}

	fmt.Printf("\nSummary: %d installed, %d missing\n", installed, missing)
	if missing > 0 {
		fmt.Println("Run 'coldcase install --container' to set up the container image")
		fmt.Println("Or  'coldcase install' to install tools on the host")
	}
}
