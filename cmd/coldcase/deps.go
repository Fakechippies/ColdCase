package main

import (
	"fmt"
	"os"
	"os/exec"

	"coldcase/pkg/runner"
	"coldcase/pkg/tools"
	vol3 "coldcase/pkg/volatility3"

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
	installCmd.Flags().Bool("container", false, "Build the ColdCase container image instead of installing to host")
	rootCmd.AddCommand(installCmd)

	depsCmd := &cobra.Command{
		Use:   "deps",
		Short: "Manage project dependencies",
		Long:  "Commands for managing project dependencies",
	}
	depsCmd.AddCommand(&cobra.Command{Use: "install", Short: "Install Python dependencies", Run: installPythonDeps})
	depsCmd.AddCommand(&cobra.Command{Use: "check", Short: "Check for missing dependencies", Run: checkDeps})
	depsCmd.AddCommand(&cobra.Command{Use: "update", Short: "Update Python dependencies", Run: updateDeps})
	rootCmd.AddCommand(depsCmd)
}

func installDependencies(cmd *cobra.Command, args []string) {
	useContainer, _ := cmd.Flags().GetBool("container")
	if useContainer {
		fmt.Println("Building ColdCase container image...")
		rt := runner.DetectedRuntime()
		if rt == "" {
			fmt.Println("[!] No container runtime (docker/podman) found on PATH.")
			os.Exit(1)
		}
		image := runner.ImageName()
		buildCmd := exec.Command(rt, "build", "-t", image, ".")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			fmt.Printf("[x] Container build failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[*] Container image '%s' built successfully.\n", image)
		fmt.Println("Missing tools will now automatically run inside the container.")
		return
	}

	useUV, _ := cmd.Flags().GetBool("uv")
	skipSystem, _ := cmd.Flags().GetBool("skip-system")

	fmt.Println("Installing ColdCase dependencies...")

	if !skipSystem {
		fmt.Println("\nInstalling system dependencies...")
		switch {
		case tools.CheckToolInstalled("apt"):
			run("bash", "-c", "sudo apt update && sudo apt install -y python3 python3-pip exiftool binwalk sleuthkit")
		case tools.CheckToolInstalled("dnf"):
			run("bash", "-c", "sudo dnf install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
		case tools.CheckToolInstalled("yum"):
			run("bash", "-c", "sudo yum install -y python3 python3-pip perl-Image-ExifTool binwalk sleuthkit")
		case tools.CheckToolInstalled("pacman"):
			run("bash", "-c", "sudo pacman -Sy --needed python python-pip perl-image-exiftool binwalk sleuthkit")
		case tools.CheckToolInstalled("brew"):
			run("bash", "-c", "brew install python exiftool binwalk sleuthkit")
		case tools.CheckToolInstalled("choco"):
			run("cmd", "/C", "choco install -y python exiftool")
		case tools.CheckToolInstalled("winget"):
			run("cmd", "/C", "winget install -e --id Python.Python.3 ; winget install -e --id ExifTool.ExifTool")
		default:
			fmt.Println("[!] Could not detect package manager. Please install manually.")
			printManualInstallGuide()
		}
	}

	fmt.Println("\nInstalling Python dependencies...")
	installPythonDeps(cmd, args)

	if useUV && !tools.CheckToolInstalled("uv") {
		fmt.Println("\nInstalling uv...")
		run("curl", "-LsSf", "https://astral.sh/uv/install.sh", "|", "sh")
	}

	fmt.Println("\n[*] Dependency installation completed!")
	fmt.Println("Run 'coldcase check' to verify all dependencies are installed.")
}

func installPythonDeps(cmd *cobra.Command, args []string) {
	if _, err := os.Stat("./volatility3"); os.IsNotExist(err) {
		fmt.Println("[!] Volatility3 directory not found")
		return
	}
	fmt.Println("Installing Volatility3 Python dependencies...")
	if tools.CheckToolInstalled("uv") {
		fmt.Println("Using uv for faster installation...")
		if err := tools.ExecuteCommand("uv", "pip", "install", "-e", "./volatility3[full]"); err != nil {
			fmt.Printf("[!] uv install failed, falling back to pip: %v\n", err)
			pipInstall()
		}
	} else {
		pipInstall()
	}
}

func pipInstall() {
	if err := tools.ExecuteCommand("python3", "-m", "pip", "install", "-e", "./volatility3[full]"); err != nil {
		fmt.Printf("[x] Python dependency installation failed: %v\n", err)
		fmt.Println("Manual: cd volatility3 && python3 -m pip install -e .[full]")
	}
}

func checkDeps(cmd *cobra.Command, args []string) {
	fmt.Println("Checking for missing dependencies...")
	missing := false

	for _, t := range []string{"python3", "exiftool", "binwalk", "fls", "fsstat", "istat"} {
		if !tools.CheckToolInstalled(t) {
			fmt.Printf("[x] Missing: %s\n", t)
			missing = true
		}
	}
	for dep, ok := range vol3.CheckDependencies() {
		if !ok {
			fmt.Printf("[x] Missing: %s\n", dep)
			missing = true
		}
	}
	if _, err := os.Stat("./DidierStevensSuite"); os.IsNotExist(err) {
		fmt.Println("[x] Missing: DidierStevensSuite directory")
		missing = true
	}
	if !missing {
		fmt.Println("[*] All dependencies are installed!")
	} else {
		fmt.Println("\nRun 'coldcase install' to install missing dependencies.")
	}
}

func updateDeps(cmd *cobra.Command, args []string) {
	fmt.Println("Updating Python dependencies...")
	if tools.CheckToolInstalled("uv") {
		if err := tools.ExecuteCommand("uv", "pip", "install", "--upgrade", "./volatility3[full]"); err != nil {
			fmt.Printf("[!] uv update failed, falling back to pip: %v\n", err)
			tools.ExecuteCommand("python3", "-m", "pip", "install", "--upgrade", "./volatility3[full]")
		}
	} else {
		if err := tools.ExecuteCommand("python3", "-m", "pip", "install", "--upgrade", "./volatility3[full]"); err != nil {
			fmt.Printf("[x] Update failed: %v\n", err)
		}
	}
	fmt.Println("[*] Dependencies updated!")
}

// run is a thin wrapper that prints a warning on error without exiting.
func run(name string, args ...string) {
	if err := tools.ExecuteCommand(name, args...); err != nil {
		fmt.Printf("[!] Command failed: %v\n", err)
	}
}

func printManualInstallGuide() {
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
}
