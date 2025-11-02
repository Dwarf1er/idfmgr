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

var (
	skipPrereqs bool
	skipClang   bool
)

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a specific ESP-IDF version",
	Long:  `Download and install a specific ESP-IDF version to the ESP_BASE directory. Use 'latest' to install the lastest published version.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]
		if err := installVersion(version); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing version %s: %v\n", version, err)
			os.Exit(1)
		}
	},
}

func init() {
	installCmd.Flags().BoolVar(&skipPrereqs, "skip-prereqs", false, "Skip prerequisite checks")
	installCmd.Flags().BoolVar(&skipClang, "skip-clang", false, "Skip esp-clang installation")
	rootCmd.AddCommand(installCmd)
}

func installVersion(version string) error {
	fmt.Printf("Installing ESP-IDF version %s...\n", version)

	if version == "latest" {
		var err error
		version, err = getLatestESPIDFVersion()
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
	}

	espBase := getESPBase()
	installPath := filepath.Join(espBase, version)

	if _, err := os.Stat(installPath); err == nil {
		fmt.Printf("Version %s is already installed at %s\n", version, installPath)
		return nil
	}

	if !skipPrereqs {
		if err := checkPrerequisites(); err != nil {
			return fmt.Errorf("prerequisite check failed: %w", err)
		}
	}

	if err := os.MkdirAll(espBase, 0o755); err != nil {
		return fmt.Errorf("failed to create ESP_BASE directory: %w", err)
	}

	if err := cloneESPIDF(version, installPath); err != nil {
		return fmt.Errorf("failed to clone ESP-IDF: %w", err)
	}

	if err := runInstallScript(installPath); err != nil {
		return fmt.Errorf("failed to run install script: %w", err)
	}

	if !skipClang {
		if err := installESPClang(installPath); err != nil {
			fmt.Printf("Warning: Failed to install esp-clang: %v\n", err)
			fmt.Println("You can install it manually later with: idf_tools.py install esp-clang")
		}
	}

	fmt.Printf("ESP-IDF version %s installed successfully at %s\n", version, installPath)
	return nil
}

func checkPrerequisites() error {
	fmt.Println("Checking prerequisites...")

	prerequisites := []string{
		"git", "wget", "python3", "cmake", "ninja",
	}

	var missing []string
	for _, cmd := range prerequisites {
		if !commandExists(cmd) {
			missing = append(missing, cmd)
		}
	}

	if len(missing) > 0 {
		fmt.Printf("❌ Missing prerequisites: %s\n", strings.Join(missing, ", "))
		fmt.Println("\nPlease install them using your package manager:")

		switch runtime.GOOS {
		case "linux":
			if commandExists("apt-get") {
				fmt.Printf("  sudo apt-get install %s\n", strings.Join(missing, " "))
			} else if commandExists("yum") {
				fmt.Printf("  sudo yum install %s\n", strings.Join(missing, " "))
			} else if commandExists("pacman") {
				fmt.Printf("  sudo pacman -S %s\n", strings.Join(missing, " "))
			} else {
				fmt.Println("  Install using your distribution's package manager")
			}
		case "darwin":
			if commandExists("brew") {
				fmt.Printf("  brew install %s\n", strings.Join(missing, " "))
			} else {
				fmt.Println("  Install Homebrew first: https://brew.sh")
				fmt.Printf("  Then: brew install %s\n", strings.Join(missing, " "))
			}
		case "windows":
			fmt.Println("  Install using chocolatey, winget, or download manually")
		}

		return fmt.Errorf("missing prerequisites")
	}

	fmt.Println("All prerequisites are installed")
	return nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func cloneESPIDF(version, installPath string) error {
	fmt.Printf("Cloning ESP-IDF %s...\n", version)

	cmd := exec.Command("git", "clone",
		"-b", version,
		"--recursive",
		"--depth", "1",
		"https://github.com/espressif/esp-idf.git",
		installPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Println("ESP-IDF cloned successfully")
	return nil
}

func runInstallScript(installPath string) error {
	fmt.Println("Running ESP-IDF install script...")

	var installScript string
	var args []string

	if runtime.GOOS == "windows" {
		installScript = filepath.Join(installPath, "install.bat")
		args = []string{"esp32"}
	} else {
		installScript = filepath.Join(installPath, "install.sh")
		args = []string{"esp32"}
	}

	if _, err := os.Stat(installScript); os.IsNotExist(err) {
		return fmt.Errorf("install script not found: %s", installScript)
	}

	cmd := exec.Command(installScript, args...)
	cmd.Dir = installPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install script failed: %w", err)
	}

	fmt.Println("ESP-IDF install script completed")
	return nil
}

func installESPClang(installPath string) error {
	fmt.Println("Installing esp-clang...")

	cmd := exec.Command("pip3", "install", "-U", "pyclang")
	if err := cmd.Run(); err != nil {
		fmt.Println("Warning: Failed to install pyclang via pip3")
	}

	idfToolsScript := filepath.Join(installPath, "tools", "idf_tools.py")
	if _, err := os.Stat(idfToolsScript); os.IsNotExist(err) {
		return fmt.Errorf("idf_tools.py not found at %s", idfToolsScript)
	}

	cmd = exec.Command("python3", idfToolsScript, "install", "esp-clang")
	cmd.Dir = installPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("esp-clang installation failed: %w", err)
	}

	fmt.Println("esp-clang installed successfully")
	return nil
}

