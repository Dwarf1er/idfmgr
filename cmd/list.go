package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// GitHub release structure
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available ESP-IDF versions from GitHub",
	Long:  `Fetch and display available ESP-IDF versions from the official GitHub releases`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listAvailableVersions(); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing versions: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listAvailableVersions() error {
	resp, err := http.Get("https://api.github.com/repos/espressif/esp-idf/releases")
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	fmt.Printf("\nAvailable ESP-IDF versions (showing latest 20):\n")
	fmt.Printf("%-15s %-20s %-12s %s\n", "VERSION", "PUBLISHED", "TYPE", "NAME")
	fmt.Printf("%s\n", strings.Repeat("-", 70))

	sort.Slice(releases, func(i, j int) bool {
		return releases[i].PublishedAt.After(releases[j].PublishedAt)
	})

	count := 0
	for _, release := range releases {
		if count >= 20 {
			break
		}

		releaseType := "stable"
		if release.Prerelease {
			releaseType = "prerelease"
		}

		fmt.Printf("%-15s %-20s %-12s %s\n",
			release.TagName,
			release.PublishedAt.Format("2006-01-02"),
			releaseType,
			release.Name,
		)
		count++
	}

	fmt.Printf("\nTo install a version: idfmgr install <version>\n")
	return nil
}