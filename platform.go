package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	platformCmd := &cobra.Command{
		Use:   "platform",
		Short: "Show platform-specific information and setup guides",
		Long:  "Display platform-specific setup instructions and dependency information",
		Run:   showPlatformInfo,
	}
	rootCmd.AddCommand(platformCmd)
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

	if checkToolInstalled("apt") {
		fmt.Println("[*] Detected: Ubuntu/Debian (apt)")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  sudo apt update")
		fmt.Println("  sudo apt install -y python3 python3-pip exiftool binwalk sleuthkit")
		fmt.Println("\nColdCase commands:")
		fmt.Println("  ./bin/coldcase install")
	} else if checkToolInstalled("dnf") {
		fmt.Println("[*] Detected: Fedora/RHEL (dnf)")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  sudo dnf install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
		fmt.Println("\nColdCase commands:")
		fmt.Println("  ./bin/coldcase install")
	} else if checkToolInstalled("yum") {
		fmt.Println("[*] Detected: RHEL/CentOS (yum)")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  sudo yum install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
		fmt.Println("\nColdCase commands:")
		fmt.Println("  ./bin/coldcase install")
	} else if checkToolInstalled("pacman") {
		fmt.Println("[*] Detected: Arch Linux (pacman)")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  sudo pacman -Sy --needed python python-pip perl-image-exiftool binwalk sleuthkit")
		fmt.Println("\nColdCase commands:")
		fmt.Println("  ./bin/coldcase install")
	} else {
		fmt.Println("[!] No supported package manager detected")
		fmt.Println("\nManual installation required. Choose your distribution:")
		fmt.Println("\nUbuntu/Debian:")
		fmt.Println("  sudo apt update && sudo apt install -y python3 python3-pip exiftool binwalk sleuthkit")
		fmt.Println("\nFedora/RHEL:")
		fmt.Println("  sudo dnf install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
		fmt.Println("\nArch Linux:")
		fmt.Println("  sudo pacman -Sy --needed python python-pip perl-image-exiftool binwalk sleuthkit")
	}

	fmt.Println("\nMemory Analysis Notes:")
	fmt.Println("- Most Volatility3 plugins work with Linux memory dumps")
	fmt.Println("- Windows memory analysis requires specific Windows memory dumps")
	fmt.Println("- Use 'coldcase info -f memory.dmp' to identify memory image type")
}

func showMacOSInfo() {
	fmt.Println("macOS Support:")
	fmt.Println("=============")

	if checkToolInstalled("brew") {
		fmt.Println("[*] Detected: Homebrew")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  brew install python exiftool binwalk sleuthkit")
		fmt.Println("\nColdCase commands:")
		fmt.Println("  ./bin/coldcase install")
		fmt.Println("  ./bin/coldcase install --uv  # for faster Python deps")
	} else {
		fmt.Println("[!] Homebrew not found")
		fmt.Println("\nInstall Homebrew first:")
		fmt.Println("  /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"")
		fmt.Println("\nThen install dependencies:")
		fmt.Println("  brew install python exiftool binwalk sleuthkit")
	}

	fmt.Println("\nmacOS Memory Analysis:")
	fmt.Println("- Volatility3 supports macOS memory analysis")
	fmt.Println("- Use mac.pslist, mac.pstree commands for macOS memory dumps")
	fmt.Println("- Some features may require elevated permissions")
}

func showWindowsInfo() {
	fmt.Println("Windows Support:")
	fmt.Println("===============")

	if checkToolInstalled("choco") {
		fmt.Println("[*] Detected: Chocolatey")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  choco install -y python exiftool")
		fmt.Println("\nColdCase commands:")
		fmt.Println("  .\\bin\\coldcase.exe install")
	} else if checkToolInstalled("winget") {
		fmt.Println("[*] Detected: Winget")
		fmt.Println("\nInstallation commands:")
		fmt.Println("  winget install -e --id Python.Python.3")
		fmt.Println("  winget install -e --id ExifTool.ExifTool")
		fmt.Println("\nColdCase commands:")
		fmt.Println("  .\\bin\\coldcase.exe install")
	} else {
		fmt.Println("[!] No package manager detected")
		fmt.Println("\nInstall Chocolatey or Winget first:")
		fmt.Println("\nChocolatey:")
		fmt.Println("  Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))")
		fmt.Println("\nWinget (Windows 10/11):")
		fmt.Println("  Available by default on recent Windows versions")

		fmt.Println("\nThen install dependencies:")
		fmt.Println("  choco install -y python exiftool")
		fmt.Println("  # OR")
		fmt.Println("  winget install -e --id Python.Python.3")
		fmt.Println("  winget install -e --id ExifTool.ExifTool")
	}

	fmt.Println("\nWindows Memory Analysis:")
	fmt.Println("- Volatility3 excels at Windows memory forensics")
	fmt.Println("- Most plugins are Windows-specific")
	fmt.Println("- Requires elevated privileges for memory acquisition")
	fmt.Println("- Common workflow:")
	fmt.Println("  .\\bin\\coldcase.exe info -f memory.dmp")
	fmt.Println("  .\\bin\\coldcase.exe windows.pslist -f memory.dmp")
}

func detectDistribution() string {
	if checkToolInstalled("apt") {
		return "Ubuntu/Debian"
	} else if checkToolInstalled("dnf") {
		return "Fedora/RHEL"
	} else if checkToolInstalled("yum") {
		return "RHEL/CentOS"
	} else if checkToolInstalled("pacman") {
		return "Arch Linux"
	} else {
		return "Unknown"
	}
}
