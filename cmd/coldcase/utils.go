package main

import (
	"fmt"
	"os"

	"coldcase/pkg/catalog"
	"coldcase/pkg/runner"
	"coldcase/pkg/tools"
	vol3 "coldcase/pkg/volatility3"

	"github.com/spf13/cobra"
)

func listTools(cmd *cobra.Command, args []string) {
	for _, c := range catalog.Categories() {
		printCatalogCategory(c.Name, c.Tools)
	}

	fmt.Println("\nUtility Commands:")
	for _, u := range []struct{ n, d string }{
		{"list", "Show this list of available tools"},
		{"check", "Check which tools are installed"},
		{"install", "Install project dependencies"},
		{"deps", "Manage project dependencies"},
		{"platform", "Show platform-specific setup information"},
		{"serve", "Start the local ColdCase web UI"},
		{"container status", "Show container runtime and image status"},
		{"container build", "Build the ColdCase container image"},
		{"container pull", "Pull the ColdCase container image"},
		{"container shell", "Open an interactive shell in the container"},
	} {
		fmt.Printf("  %-26s - %s\n", u.n, u.d)
	}
}

// printCatalogCategory renders one catalog category using the same layout as
// the historical `coldcase list` output.
func printCatalogCategory(title string, ts []catalog.Tool) {
	fmt.Printf("\n%s:\n", title)
	for _, t := range ts {
		fmt.Printf("  %-26s - %s\n", t.Name, t.Description)
	}
}

func checkTools(cmd *cobra.Command, args []string) {
	fmt.Println("Checking installed tools...")

	// Every underlying binary in the catalog is checked once.
	seen := map[string]bool{}
	installed, missing := 0, 0
	for _, t := range catalog.All() {
		if seen[t.Binary] {
			continue
		}
		seen[t.Binary] = true
		if tools.CheckToolInstalled(t.Binary) {
			fmt.Printf("[*] %-20s installed\n", t.Binary)
			installed++
		} else {
			fmt.Printf("[x] %-20s not found\n", t.Binary)
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

	if _, err := os.Stat(catalog.DefaultSuitePath); err == nil {
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
