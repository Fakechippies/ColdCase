package main

import (
	"fmt"
	"os"
	"strings"

	"coldcase/pkg/binwalk"
	"coldcase/pkg/carving"
	"coldcase/pkg/didier"
	"coldcase/pkg/exiftool"
	"coldcase/pkg/hashing"
	"coldcase/pkg/malware"
	"coldcase/pkg/mobile"
	"coldcase/pkg/network"
	"coldcase/pkg/runner"
	"coldcase/pkg/sleuthkit"
	"coldcase/pkg/steg"
	"coldcase/pkg/sysutils"
	"coldcase/pkg/timeline"
	vol3 "coldcase/pkg/volatility3"
	wintools "coldcase/pkg/windows"

	"github.com/spf13/cobra"
)

const defaultSuitePath = "./DidierStevensSuite"

var rootCmd = &cobra.Command{
	Use:   "coldcase",
	Short: "Integrated Digital Forensics Tool",
	Long:  "A comprehensive CLI tool integrating 100+ digital forensics utilities",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Original tool categories
	addDidierStevensCommands()
	addExifToolCommand()
	addBinwalkCommand()
	addSleuthKitCommands()
	addVolatility3Commands()

	// New tool categories
	addGenericCommands("network", network.Tools())
	addGenericCommands("carving", carving.Tools())
	addGenericCommands("malware", malware.Tools())
	addGenericCommands("hashing", hashing.Tools())
	addPlasoCommandGroup()
	addGenericCommands("mobile", mobile.Tools())
	addGenericCommands("windows-artifacts", wintools.Tools())
	addGenericCommands("steg", steg.Tools())
	addGenericCommands("sysutils", sysutils.Tools())

	// Built-in utilities
	addListCommand()
	addCheckCommand()
}

// tool is a minimal interface matched by all pkg tool types.
type tool interface {
	Name() string
	Description() string
	Run(args []string) error
}

// addGenericCommands registers a slice of tools that satisfy the tool interface.
func addGenericCommands[T tool](_ string, tools []T) {
	for _, t := range tools {
		t := t // capture
		cmd := &cobra.Command{
			Use:   t.Name(),
			Short: t.Description(),
			Run: func(cmd *cobra.Command, args []string) {
				if err := t.Run(args); err != nil {
					fmt.Printf("Error running %s: %v\n", t.Name(), err)
					os.Exit(1)
				}
			},
		}
		rootCmd.AddCommand(cmd)
	}
}

// ─── DidierStevens ────────────────────────────────────────────────────────────

func addDidierStevensCommands() {
	for _, t := range didier.Tools(defaultSuitePath) {
		t := t
		cmd := &cobra.Command{
			Use:   t.Name(),
			Short: t.Description(),
			Run: func(cmd *cobra.Command, args []string) {
				if err := t.Run(args); err != nil {
					fmt.Printf("Error running %s: %v\n", t.Name(), err)
					os.Exit(1)
				}
			},
		}
		rootCmd.AddCommand(cmd)
	}
}

// ─── ExifTool ─────────────────────────────────────────────────────────────────

func addExifToolCommand() {
	t := exiftool.New()
	rootCmd.AddCommand(&cobra.Command{
		Use:   t.Name(),
		Short: t.Description(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := t.Run(args); err != nil {
				fmt.Printf("Error running exiftool: %v\n", err)
				os.Exit(1)
			}
		},
	})
}

// ─── Binwalk ──────────────────────────────────────────────────────────────────

func addBinwalkCommand() {
	t := binwalk.New()
	rootCmd.AddCommand(&cobra.Command{
		Use:   t.Name(),
		Short: t.Description(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := t.Run(args); err != nil {
				fmt.Printf("Error running binwalk: %v\n", err)
				os.Exit(1)
			}
		},
	})
}

// ─── Sleuth Kit ───────────────────────────────────────────────────────────────

func addSleuthKitCommands() {
	for _, t := range sleuthkit.Tools() {
		t := t
		rootCmd.AddCommand(&cobra.Command{
			Use:   t.Name(),
			Short: t.Description(),
			Run: func(cmd *cobra.Command, args []string) {
				if err := t.Run(args); err != nil {
					fmt.Printf("Error running %s: %v\n", t.Name(), err)
					os.Exit(1)
				}
			},
		})
	}
}

// ─── Volatility3 ──────────────────────────────────────────────────────────────

func addVolatility3Commands() {
	for _, t := range vol3.Tools() {
		t := t
		cmd := &cobra.Command{
			Use:   t.Name(),
			Short: t.Description(),
			Long:  t.Description() + " — Volatility3 memory forensics",
			Run: func(cmd *cobra.Command, args []string) {
				if err := t.Run(args); err != nil {
					fmt.Printf("Error running %s: %v\n", t.Name(), err)
					os.Exit(1)
				}
			},
		}
		if strings.HasPrefix(t.Name(), "windows.") ||
			strings.HasPrefix(t.Name(), "linux.") ||
			strings.HasPrefix(t.Name(), "mac.") {
			cmd.Flags().StringP("file", "f", "", "Memory image file to analyze")
		}
		rootCmd.AddCommand(cmd)
	}
}

// ─── Built-in utilities ───────────────────────────────────────────────────────

func addListCommand() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available forensics tools",
		Run:   listTools,
	})
}

func addCheckCommand() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Check which tools are installed or available via container",
		Run:   checkTools,
	})
}

// ─── Plaso ──────────────────────────────────────────────────────────────────

func addPlasoCommandGroup() {
	plasoCmd := &cobra.Command{
		Use:   "plaso",
		Short: "Plaso (log2timeline) supertimeline generation and analysis",
	}

	for _, t := range timeline.Tools() {
		t := t
		var cmd *cobra.Command

		switch t.Name() {
		case "log2timeline":
			cmd = &cobra.Command{
				Use:   "parse",
				Short: t.Description(),
				Run: func(cmd *cobra.Command, args []string) {
					if err := t.Run(args); err != nil {
						fmt.Printf("Error running plaso parse: %v\n", err)
						os.Exit(1)
					}
				},
			}
		case "psort":
			cmd = &cobra.Command{
				Use:   "sort",
				Short: t.Description(),
				Run: func(cmd *cobra.Command, args []string) {
					if err := t.Run(args); err != nil {
						fmt.Printf("Error running plaso sort: %v\n", err)
						os.Exit(1)
					}
				},
			}
		case "psteal":
			cmd = &cobra.Command{
				Use:   "psteal",
				Short: t.Description(),
				Run: func(cmd *cobra.Command, args []string) {
					if err := t.Run(args); err != nil {
						fmt.Printf("Error running plaso psteal: %v\n", err)
						os.Exit(1)
					}
				},
			}
		case "hayabusa", "evtx_dump", "timeliner", "chainsaw":
			// Keep these as top-level if needed, or skip for now.
			// Actually, the user specifically mentioned plaso parse/parsers.
			// Let's add the other timeline tools back as top-level or in another group.
			rootCmd.AddCommand(&cobra.Command{
				Use:   t.Name(),
				Short: t.Description(),
				Run: func(cmd *cobra.Command, args []string) {
					if err := t.Run(args); err != nil {
						fmt.Printf("Error running %s: %v\n", t.Name(), err)
						os.Exit(1)
					}
				},
			})
			continue
		}

		if cmd != nil {
			plasoCmd.AddCommand(cmd)
		}
	}

	// Add the specific "parsers --list" subcommand
	plasoCmd.AddCommand(&cobra.Command{
		Use:   "parsers",
		Short: "List all available plaso parsers",
		Run: func(cmd *cobra.Command, args []string) {
			// Find log2timeline tool to run it with --parsers list
			for _, t := range timeline.Tools() {
				if t.Name() == "log2timeline" {
					if err := t.Run([]string{"--parsers", "list"}); err != nil {
						fmt.Printf("Error listing plaso parsers: %v\n", err)
						os.Exit(1)
					}
					return
				}
			}
		},
	})

	rootCmd.AddCommand(plasoCmd)
}

// ─── runner re-export for container.go ───────────────────────────────────────
// This allows container.go (same package) to use runner without its own import.
var _ = runner.ContainerAvailable
