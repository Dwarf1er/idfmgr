package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate the ESP-IDF environment for the current project",
	Long:  `Automatically sets up the ESP-IDF environment variables for the version specified in .espidf-version`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := activateProject(); err != nil {
			fmt.Fprintf(os.Stderr, "Error activating ESP-IDF environment: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(activateCmd)
}

func activateProject() error {
	versionFile := ".espidf-version"
	versionData, err := os.ReadFile(versionFile)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", versionFile, err)
	}
	version := strings.TrimSpace(string(versionData))

	idfPath := filepath.Join(getESPBase(), version)
	if _, err := os.Stat(idfPath); os.IsNotExist(err) {
		return fmt.Errorf("ESP-IDF path does not exist: %s", idfPath)
	}

	switch runtime.GOOS {
	case "windows":
		return activateWindows(idfPath)
	default:
		return activateUnix(idfPath)
	}
}

func activateUnix(idfPath string) error {
	envVars, err := getESPIDFEnvironment(idfPath)
	if err != nil {
		return err
	}

	for _, e := range envVars {
		if e == "" || !strings.Contains(e, "=") {
			continue
		}

		parts := strings.SplitN(e, "=", 2)
		key := parts[0]
		val := parts[1]

		if !isValidEnvKey(key) {
			continue
		}

		val = strings.ReplaceAll(val, `'`, `'\''`)
		fmt.Printf("export %s='%s'\n", key, val)
	}

	fmt.Println("# To activate, run in your shell:")
	fmt.Println("# eval $(idfmgr activate)")
	return nil
}

func activateWindows(idfPath string) error {
	exportBat := filepath.Join(idfPath, "export.bat")
	if _, err := os.Stat(exportBat); os.IsNotExist(err) {
		return fmt.Errorf("could not find export.bat in %s", idfPath)
	}

	fmt.Println("Launching new PowerShell with ESP-IDF environment set...")

	psCommand := fmt.Sprintf("cmd /c \"%s && powershell\"", exportBat)

	cmd := exec.Command("powershell.exe", "-NoExit", "-Command", psCommand)
	cmd.Dir = idfPath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isValidEnvKey(key string) bool {
	if key == "" {
		return false
	}
	r := []rune(key)
	if !(r[0] >= 'A' && r[0] <= 'Z' || r[0] >= 'a' && r[0] <= 'z' || r[0] == '_') {
		return false
	}
	for _, c := range r[1:] {
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}