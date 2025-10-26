package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func getESPBase() string {
	espBase := os.Getenv("ESP_BASE")
	if espBase == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ".esp"
		}
		espBase = filepath.Join(homeDir, ".esp")
	}
	return espBase
}

func getLatestESPIDFVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/espressif/esp-idf/releases/latest")
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	return release.TagName, nil
}

func getLatestInstalledESPIDFVersion() (string, error) {
	versions, err := getInstalledVersions(getESPBase())
	if err != nil {
		return "", fmt.Errorf("failed to get installed versions: %w", err)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no ESP-IDF versions installed")
	}

	sort.Strings(versions)
	return versions[len(versions)-1], nil
}

func getESPIDFEnvironment(idfPath string) ([]string, error) {
	exportScript := filepath.Join(idfPath, "export.sh")

	cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s > /dev/null 2>&1 && env", exportScript))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to source export.sh: %w", err)
	}

	envVars := strings.Split(string(output), "\n")
	return envVars, nil
}

