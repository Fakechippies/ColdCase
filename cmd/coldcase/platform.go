package main

import (
	"fmt"
	"runtime"
	"strings"

	"coldcase/pkg/tools"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "platform",
		Short: "Show platform-specific information and setup guides",
		Long:  "Display platform-specific setup instructions and dependency information",
		Run:   showPlatformInfo,
	})
}

func showPlatformInfo(cmd *cobra.Command, args []string) {
	osName := runtime.GOOS
	fmt.Printf("Detected Platform: %s\n\n", strings.Title(osName))
	switch osName {
	case "linux":
		showLinuxInfo()
	case "darwin":
		showMacOSInfo()
	case "windows":
		showWindowsInfo()
	default:
		fmt.Printf("Unsupported platform: %s\n", osName)
	}
}

func showLinuxInfo() {
	fmt.Println("Linux Distribution Support:")
	fmt.Println("=========================")
	switch {
	case tools.CheckToolInstalled("apt"):
		fmt.Println("[*] Detected: Ubuntu/Debian (apt)")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  sudo apt update")
		fmt.Println("  sudo apt install -y python3 python3-pip exiftool binwalk sleuthkit")
	case tools.CheckToolInstalled("dnf"):
		fmt.Println("[*] Detected: Fedora/RHEL (dnf)")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  sudo dnf install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
	case tools.CheckToolInstalled("yum"):
		fmt.Println("[*] Detected: RHEL/CentOS (yum)")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  sudo yum install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
	case tools.CheckToolInstalled("pacman"):
		fmt.Println("[*] Detected: Arch Linux (pacman)")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  sudo pacman -Sy --needed python python-pip perl-image-exiftool binwalk sleuthkit")
	default:
		fmt.Println("[!] No supported package manager detected")
		fmt.Println("\nManual installation required. Choose your distribution:")
		fmt.Println("\nUbuntu/Debian:  sudo apt update && sudo apt install -y python3 python3-pip exiftool binwalk sleuthkit")
		fmt.Println("Fedora/RHEL:    sudo dnf install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
		fmt.Println("Arch Linux:     sudo pacman -Sy --needed python python-pip perl-image-exiftool binwalk sleuthkit")
	}
	fmt.Println("\nRun: ./bin/coldcase install")
	fmt.Println("\nMemory Analysis Notes:")
	fmt.Println("- Use 'coldcase info -f memory.dmp' to identify memory image type")
}

func showMacOSInfo() {
	fmt.Println("macOS Support:")
	fmt.Println("=============")
	if tools.CheckToolInstalled("brew") {
		fmt.Println("[*] Detected: Homebrew")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  brew install python exiftool binwalk sleuthkit")
		fmt.Println("\nRun: ./bin/coldcase install  (or --uv for faster Python deps)")
	} else {
		fmt.Println("[!] Homebrew not found")
		fmt.Println("\nInstall Homebrew first:")
		fmt.Println(`  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
		fmt.Println("\nThen: brew install python exiftool binwalk sleuthkit")
	}
	fmt.Println("\nmacOS Memory Analysis: use mac.pslist / mac.pstree commands.")
}

func showWindowsInfo() {
	fmt.Println("Windows Support:")
	fmt.Println("===============")
	switch {
	case tools.CheckToolInstalled("choco"):
		fmt.Println("[*] Detected: Chocolatey")
		fmt.Println("\n  choco install -y python exiftool")
		fmt.Println("\n  .\\bin\\coldcase.exe install")
	case tools.CheckToolInstalled("winget"):
		fmt.Println("[*] Detected: Winget")
		fmt.Println("\n  winget install -e --id Python.Python.3")
		fmt.Println("  winget install -e --id ExifTool.ExifTool")
		fmt.Println("\n  .\\bin\\coldcase.exe install")
	default:
		fmt.Println("[!] No package manager detected. Install Chocolatey or Winget first.")
	}
	fmt.Println("\nWindows Memory Analysis: most Volatility3 plugins are Windows-specific.")
	fmt.Println("  .\\bin\\coldcase.exe info -f memory.dmp")
	fmt.Println("  .\\bin\\coldcase.exe windows.pslist -f memory.dmp")
}
