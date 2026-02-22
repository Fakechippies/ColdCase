package main

import (
	"fmt"
	"os"
	"os/exec"

	"coldcase/pkg/runner"

	"github.com/spf13/cobra"
)

func init() {
	containerCmd := &cobra.Command{
		Use:   "container",
		Short: "Manage the ColdCase container image",
		Long:  "Build, pull, inspect, and shell into the ColdCase container image used for tool fallback",
	}

	containerCmd.AddCommand(
		containerStatusCmd(),
		containerBuildCmd(),
		containerPullCmd(),
		containerShellCmd(),
	)

	rootCmd.AddCommand(containerCmd)
}

// ─── status ───────────────────────────────────────────────────────────────────

func containerStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Report container runtime and image status",
		Run: func(cmd *cobra.Command, args []string) {
			rt := runner.DetectedRuntime()
			if rt == "" {
				fmt.Println("[x] No container runtime found (docker or podman required)")
			} else {
				fmt.Printf("[*] Container runtime : %s\n", rt)
			}

			image := runner.ImageName()
			fmt.Printf("[*] Container image   : %s\n", image)

			if rt != "" {
				out, err := exec.Command(rt, "image", "inspect", image, "--format", "{{.Id}}").Output()
				if err == nil && len(out) > 0 {
					fmt.Printf("[*] Image status      : present (%s...)\n", string(out)[:12])
				} else {
					fmt.Println("[!] Image status      : not pulled/built yet")
					fmt.Printf("    Run: coldcase container pull\n")
					fmt.Printf("    Or:  coldcase container build\n")
				}
			}
		},
	}
}

// ─── build ────────────────────────────────────────────────────────────────────

func containerBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the ColdCase container image from the local Dockerfile",
		Run: func(cmd *cobra.Command, args []string) {
			rt := runner.DetectedRuntime()
			if rt == "" {
				fmt.Fprintln(os.Stderr, "Error: no container runtime found")
				os.Exit(1)
			}
			image := runner.ImageName()
			fmt.Printf("Building %s with %s...\n", image, rt)
			buildCmd := exec.Command(rt, "build", "-t", image, ".")
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr
			if err := buildCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("[*] Image %s built successfully.\n", image)
		},
	}
	return cmd
}

// ─── pull ─────────────────────────────────────────────────────────────────────

func containerPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull the ColdCase container image from a registry",
		Long: `Pull the ColdCase image. Override the registry image with:
  COLDCASE_IMAGE=myregistry/coldcase:v1 coldcase container pull`,
		Run: func(cmd *cobra.Command, args []string) {
			rt := runner.DetectedRuntime()
			if rt == "" {
				fmt.Fprintln(os.Stderr, "Error: no container runtime found")
				os.Exit(1)
			}
			image := runner.ImageName()
			fmt.Printf("Pulling %s...\n", image)
			pullCmd := exec.Command(rt, "pull", image)
			pullCmd.Stdout = os.Stdout
			pullCmd.Stderr = os.Stderr
			if err := pullCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Pull failed: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

// ─── shell ────────────────────────────────────────────────────────────────────

func containerShellCmd() *cobra.Command {
	var workdir string
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Drop into an interactive shell inside the container",
		Long: `Launch an interactive bash session in the ColdCase container.
The current directory is bind-mounted at /data inside the container.`,
		Run: func(cmd *cobra.Command, args []string) {
			rt := runner.DetectedRuntime()
			if rt == "" {
				fmt.Fprintln(os.Stderr, "Error: no container runtime found")
				os.Exit(1)
			}
			image := runner.ImageName()
			wd := workdir
			if wd == "" {
				var err error
				wd, err = os.Getwd()
				if err != nil {
					wd = "."
				}
			}
			shellCmd := exec.Command(rt, "run", "--rm", "-it",
				"-v", fmt.Sprintf("%s:/data", wd),
				"-w", "/data",
				image, "/bin/bash")
			shellCmd.Stdout = os.Stdout
			shellCmd.Stderr = os.Stderr
			shellCmd.Stdin = os.Stdin
			if err := shellCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Shell exited: %v\n", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVarP(&workdir, "workdir", "w", "", "Host directory to mount at /data (default: current dir)")
	return cmd
}
