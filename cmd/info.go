package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show project and environment information",
	Long:  `Display information about the current ESP-IDF project including version, paths, and activation instructions`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := showInfo(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func showInfo() error {
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
	exportScript := filepath.Join(idfPath, "export.sh")

	fmt.Println("Project Information")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("ESP-IDF Version: %s\n", version)
	fmt.Printf("IDF Path:        %s\n", idfPath)

	if _, err := os.Stat(idfPath); os.IsNotExist(err) {
		fmt.Printf("\n   ESP-IDF version %s is not installed\n", version)
		fmt.Printf("Install it with: idfmgr install %s\n", version)
		return nil
	}

	hasBuildGcc := false
	hasBuildClang := false
	if _, err := os.Stat("build"); err == nil {
		hasBuildGcc = true
	}
	if _, err := os.Stat("build-clang"); err == nil {
		hasBuildClang = true
	}

	if hasBuildGcc || hasBuildClang {
		fmt.Println("\nBuild Status")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
		if hasBuildGcc {
			fmt.Println("✓ GCC build directory exists (build/)")
		}
		if hasBuildClang {
			fmt.Println("✓ Clang build directory exists (build-clang/)")
		}
	}

	fmt.Println("\nUsage")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Use idfmgr commands (recommended):")
	fmt.Println("  idfmgr build")
	fmt.Println("  idfmgr flash")
	fmt.Println("  idfmgr exec menuconfig")
	fmt.Println("  idfmgr exec monitor")

	fmt.Println("\nOr manually activate the environment:")
	fmt.Printf("  . %s\n", exportScript)
	fmt.Println("  idf.py build")

	return nil
}