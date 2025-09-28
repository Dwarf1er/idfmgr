package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	removeAll bool
	force     bool
)

var removeCmd = &cobra.Command{
	Use:   "remove [version...]",
	Short: "Remove installed ESP-IDF versions",
	Long: `Remove one or more installed ESP-IDF versions from the ESP_BASE directory.
Use --all to remove all installed versions.`,
	Example: `  esp-devkit remove v5.1.2
  esp-devkit remove v4.4.6 v5.0.0
  esp-devkit remove --all
  esp-devkit remove v5.1.2 --force`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := removeVersions(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing versions: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	removeCmd.Flags().BoolVar(&removeAll, "all", false, "Remove all installed versions")
	removeCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")
	rootCmd.AddCommand(removeCmd)
}

func removeVersions(versions []string) error {
	espBase := getESPBase()

	// Check if ESP_BASE directory exists
	if _, err := os.Stat(espBase); os.IsNotExist(err) {
		fmt.Printf("ESP_BASE directory doesn't exist: %s\n", espBase)
		return nil
	}

	var toRemove []string

	if removeAll {
		installed, err := getInstalledVersions(espBase)
		if err != nil {
			return fmt.Errorf("failed to get installed versions: %w", err)
		}
		
		if len(installed) == 0 {
			fmt.Println("No ESP-IDF versions are installed.")
			return nil
		}
		
		toRemove = installed
		fmt.Printf("Will remove all %d installed versions:\n", len(toRemove))
		for _, version := range toRemove {
			fmt.Printf("  - %s\n", version)
		}
	} else {
		if len(versions) == 0 {
			return fmt.Errorf("specify versions to remove or use --all flag")
		}
		
		for _, version := range versions {
			versionPath := filepath.Join(espBase, version)
			if _, err := os.Stat(versionPath); os.IsNotExist(err) {
				fmt.Printf("Warning: Version %s is not installed, skipping\n", version)
				continue
			}
			
			if !isValidESPIDFInstall(versionPath) {
				fmt.Printf("Warning: %s doesn't appear to be a valid ESP-IDF installation, skipping\n", version)
				continue
			}
			
			toRemove = append(toRemove, version)
		}
		
		if len(toRemove) == 0 {
			fmt.Println("No valid versions to remove.")
			return nil
		}
		
		fmt.Printf("Will remove %d version(s):\n", len(toRemove))
		for _, version := range toRemove {
			fmt.Printf("  - %s\n", version)
		}
	}

	totalSize, err := calculateTotalSize(espBase, toRemove)
	if err != nil {
		fmt.Printf("Warning: Could not calculate disk space: %v\n", err)
	} else {
		fmt.Printf("\nTotal disk space to be freed: %s\n", formatBytes(totalSize))
	}

	if !force {
		fmt.Print("\nAre you sure you want to proceed? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	var removed []string
	var failed []string

	for _, version := range toRemove {
		versionPath := filepath.Join(espBase, version)
		
		if err := os.RemoveAll(versionPath); err != nil {
			fmt.Printf("❌ Failed to remove %s: %v\n", version, err)
			failed = append(failed, version)
		} else {
			removed = append(removed, version)
		}
	}

	fmt.Printf("\n✅ Successfully removed %d version(s):\n", len(removed))
	for _, version := range removed {
		fmt.Printf("  - %s\n", version)
	}

	if len(failed) > 0 {
		fmt.Printf("\n❌ Failed to remove %d version(s):\n", len(failed))
		for _, version := range failed {
			fmt.Printf("  - %s\n", version)
		}
	}

	return nil
}

func getInstalledVersions(espBase string) ([]string, error) {
	entries, err := os.ReadDir(espBase)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versionPath := filepath.Join(espBase, entry.Name())
			if isValidESPIDFInstall(versionPath) {
				versions = append(versions, entry.Name())
			}
		}
	}

	return versions, nil
}

func calculateTotalSize(espBase string, versions []string) (int64, error) {
	var totalSize int64

	for _, version := range versions {
		versionPath := filepath.Join(espBase, version)
		size, err := getDirSize(versionPath)
		if err != nil {
			return 0, err
		}
		totalSize += size
	}

	return totalSize, nil
}

func getDirSize(path string) (int64, error) {
	var size int64
	
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, err
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}