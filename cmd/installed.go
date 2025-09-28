package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var installedCmd = &cobra.Command{
	Use:   "installed",
	Short: "List installed ESP-IDF versions",
	Long:  `Show all currently installed ESP-IDF versions in the ESP_BASE directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listInstalledVersions(); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing installed versions: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installedCmd)
}

func listInstalledVersions() error {
	espBase := getESPBase()
	fmt.Printf("Installed ESP-IDF versions in %s:\n\n", espBase)

	if _, err := os.Stat(espBase); os.IsNotExist(err) {
		fmt.Printf("ESP_BASE directory doesn't exist: %s\n", espBase)
		fmt.Printf("No versions installed yet.\n")
		return nil
	}

	entries, err := os.ReadDir(espBase)
	if err != nil {
		return fmt.Errorf("failed to read ESP_BASE directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			idfPath := filepath.Join(espBase, entry.Name())
			if isValidESPIDFInstall(idfPath) {
				versions = append(versions, entry.Name())
			}
		}
	}

	if len(versions) == 0 {
		fmt.Println("No ESP-IDF versions found.")
		fmt.Printf("Install a version with: idfmgr install <version>\n")
		return nil
	}

	sort.Strings(versions)

	fmt.Printf("%-15s %s\n", "VERSION", "PATH")
	fmt.Printf("%s\n", strings.Repeat("-", 50))
	
	for _, version := range versions {
		fmt.Printf("%-15s %s\n", version, filepath.Join(espBase, version))
	}

	fmt.Printf("\nTotal: %d version(s) installed\n", len(versions))
	return nil
}

func isValidESPIDFInstall(path string) bool {
	requiredPaths := []string{
		"tools",
		"components",
		"export.sh",
	}

	for _, required := range requiredPaths {
		if _, err := os.Stat(filepath.Join(path, required)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}