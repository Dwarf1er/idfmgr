package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [idf.py args...]",
	Short: "Execute idf.py command with proper environment",
	Long:  `Run any idf.py command with the correct ESP-IDF environment automatically configured`,
	Example: `  idfmgr exec menuconfig
  idfmgr exec -p /dev/ttyUSB0 monitor
  idfmgr exec app-flash
  idfmgr exec erase-flash`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := execIdfPy(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func execIdfPy(args []string) error {
	versionFile := ".espidf-version"
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		return fmt.Errorf(".espidf-version file not found. Are you in an ESP-IDF project directory?")
	}

	versionData, err := os.ReadFile(versionFile)
	if err != nil {
		return fmt.Errorf("failed to read .espidf-version: %w", err)
	}

	version := strings.TrimSpace(string(versionData))
	idfPath := filepath.Join(getESPBase(), version)

	if _, err := os.Stat(idfPath); os.IsNotExist(err) {
		return fmt.Errorf("ESP-IDF version %s is not installed. Install it with: idfmgr install %s", version, version)
	}

	env, err := getESPIDFEnvironment(idfPath)
	if err != nil {
		return fmt.Errorf("failed to setup ESP-IDF environment: %w", err)
	}

	idfPyPath := filepath.Join(idfPath, "tools", "idf.py")

	cmdArgs := append([]string{idfPyPath}, args...)
	cmd := exec.Command("python3", cmdArgs...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}