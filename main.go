package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

type Tool interface {
	Name() string
	Description() string
	Run(args []string) error
}

var rootCmd = &cobra.Command{
	Use:   "df-tool",
	Short: "Integrated Digital Forensics Tool",
	Long:  "A comprehensive CLI tool integrating various digital forensics utilities",
}

func init() {
	addDidierStevensCommands()
	addExifToolCommand()
	addBinwalkCommand()
	addSleuthKitCommands()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkToolInstalled(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
