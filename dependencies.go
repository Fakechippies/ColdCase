package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install all project dependencies",
		Long:  "Install system dependencies, Python packages, and external tools",
		Run:   installDependencies,
	}
	installCmd.Flags().Bool("uv", false, "Use uv for Python dependency management (faster)")
	installCmd.Flags().Bool("skip-system", false, "Skip system package installation")
	rootCmd.AddCommand(installCmd)

	depsCmd := &cobra.Command{
		Use:   "deps",
		Short: "Manage project dependencies",
		Long:  "Commands for managing project dependencies",
	}

	depsCmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install Python dependencies",
		Run:   installPythonDeps,
	})

	depsCmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Check for missing dependencies",
		Run:   checkDeps,
	})

	depsCmd.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Update Python dependencies",
		Run:   updateDeps,
	})

	rootCmd.AddCommand(depsCmd)
}

func installDependencies(cmd *cobra.Command, args []string) {
	useUV, _ := cmd.Flags().GetBool("uv")
	skipSystem, _ := cmd.Flags().GetBool("skip-system")

	fmt.Println("Installing ColdCase dependencies...")

	if !skipSystem {
		fmt.Println("\nInstalling system dependencies...")
		if checkToolInstalled("apt") {
			systemCmd := "sudo apt update && sudo apt install -y python3 python3-pip exiftool binwalk sleuthkit"
			if err := executeCommand("bash", "-c", systemCmd); err != nil {
				fmt.Printf("⚠️  System package installation failed: %v\n", err)
			}
		} else if checkToolInstalled("dnf") {
			systemCmd := "sudo dnf install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit"
			if err := executeCommand("bash", "-c", systemCmd); err != nil {
				fmt.Printf("⚠️  System package installation failed: %v\n", err)
			}
		} else if checkToolInstalled("yum") {
			systemCmd := "sudo yum install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit"
			if err := executeCommand("bash", "-c", systemCmd); err != nil {
				fmt.Printf("⚠️  System package installation failed: %v\n", err)
			}
		} else if checkToolInstalled("pacman") {
			systemCmd := "sudo pacman -Sy --needed python python-pip perl-image-exiftool binwalk sleuthkit"
			if err := executeCommand("bash", "-c", systemCmd); err != nil {
				fmt.Printf("⚠️  System package installation failed: %v\n", err)
			}
		} else if checkToolInstalled("brew") {
			systemCmd := "brew install python exiftool binwalk sleuthkit"
			if err := executeCommand("bash", "-c", systemCmd); err != nil {
				fmt.Printf("⚠️  System package installation failed: %v\n", err)
			}
		} else if checkToolInstalled("choco") {
			systemCmd := "choco install -y python exiftool"
			if err := executeCommand("cmd", "/C", systemCmd); err != nil {
				fmt.Printf("⚠️  Chocolatey installation failed: %v\n", err)
			}
		} else if checkToolInstalled("winget") {
			systemCmd := "winget install -e --id Python.Python.3 ; winget install -e --id ExifTool.ExifTool"
			if err := executeCommand("cmd", "/C", systemCmd); err != nil {
				fmt.Printf("⚠️  Winget installation failed: %v\n", err)
			}
		} else {
			fmt.Println("⚠️  Could not detect package manager. Please install manually:")
			fmt.Println("\nLinux (Ubuntu/Debian):")
			fmt.Println("  sudo apt update && sudo apt install -y python3 python3-pip exiftool binwalk sleuthkit")
			fmt.Println("\nLinux (Fedora/RHEL):")
			fmt.Println("  sudo dnf install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
			fmt.Println("\nLinux (Arch):")
			fmt.Println("  sudo pacman -Sy --needed python python-pip perl-image-exiftool binwalk sleuthkit")
			fmt.Println("\nmacOS:")
			fmt.Println("  brew install python exiftool binwalk sleuthkit")
			fmt.Println("\nWindows:")
			fmt.Println("  choco install -y python exiftool")
			fmt.Println("  or")
			fmt.Println("  winget install -e --id Python.Python.3 ; winget install -e --id ExifTool.ExifTool")
		}
	}

	fmt.Println("\nInstalling Python dependencies...")
	installPythonDeps(cmd, args)

	if useUV {
		fmt.Println("\nInstalling uv for faster Python package management...")
		if !checkToolInstalled("uv") {
			if err := executeCommand("curl", "-LsSf", "https://astral.sh/uv/install.sh", "|", "sh"); err != nil {
				fmt.Printf("⚠️  uv installation failed: %v\n", err)
			}
		}
	}

	fmt.Println("\n✅ Dependency installation completed!")
	fmt.Println("Run 'coldcase check' to verify all dependencies are installed.")
}

func installPythonDeps(cmd *cobra.Command, args []string) {
	// Install volatility3 dependencies
	volPath := "./volatility3"
	if _, err := os.Stat(volPath); err == nil {
		fmt.Println("Installing Volatility3 Python dependencies...")

		useUV := checkToolInstalled("uv")

		if useUV {
			fmt.Println("Using uv for faster installation...")
			if err := executeCommand("uv", "pip", "install", "-e", "./volatility3[full]"); err != nil {
				fmt.Printf("⚠️  uv installation failed, falling back to pip: %v\n", err)
				fallBackToPip()
			}
		} else {
			fallBackToPip()
		}
	} else {
		fmt.Println("⚠️  Volatility3 directory not found")
	}
}

func fallBackToPip() {
	if err := executeCommand("python3", "-m", "pip", "install", "-e", "./volatility3[full]"); err != nil {
		fmt.Printf("❌ Python dependency installation failed: %v\n", err)
		fmt.Println("You may need to install manually:")
		fmt.Println("  cd volatility3 && python3 -m pip install -e .[full]")
	}
}

func checkDeps(cmd *cobra.Command, args []string) {
	fmt.Println("Checking for missing dependencies...")

	missing := false

	// Check system tools
	systemTools := []string{"python3", "exiftool", "binwalk"}
	for _, tool := range systemTools {
		if !checkToolInstalled(tool) {
			fmt.Printf("❌ Missing: %s\n", tool)
			missing = true
		}
	}

	// Check sleuth kit tools
	sleuthTools := []string{"fls", "fsstat", "istat"}
	for _, tool := range sleuthTools {
		if !checkToolInstalled(tool) {
			fmt.Printf("❌ Missing: %s (sleuthkit)\n", tool)
			missing = true
		}
	}

	// Check volatility3
	volDeps := checkVolatility3Dependencies()
	for dep, installed := range volDeps {
		if !installed {
			fmt.Printf("❌ Missing: %s\n", dep)
			missing = true
		}
	}

	// Check directories
	if _, err := os.Stat("./DidierStevensSuite"); os.IsNotExist(err) {
		fmt.Println("❌ Missing: DidierStevensSuite directory")
		missing = true
	}

	if !missing {
		fmt.Println("✅ All dependencies are installed!")
	} else {
		fmt.Println("\nRun 'coldcase install' to install missing dependencies.")
	}
}

func updateDeps(cmd *cobra.Command, args []string) {
	fmt.Println("Updating Python dependencies...")

	useUV := checkToolInstalled("uv")

	if useUV {
		fmt.Println("Using uv for faster updates...")
		if err := executeCommand("uv", "pip", "install", "--upgrade", "./volatility3[full]"); err != nil {
			fmt.Printf("⚠️  uv update failed, falling back to pip: %v\n", err)
			executeCommand("python3", "-m", "pip", "install", "--upgrade", "./volatility3[full]")
		}
	} else {
		if err := executeCommand("python3", "-m", "pip", "install", "--upgrade", "./volatility3[full]"); err != nil {
			fmt.Printf("❌ Update failed: %v\n", err)
		}
	}

	fmt.Println("✅ Dependencies updated!")
}
