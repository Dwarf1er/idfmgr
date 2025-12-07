package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	flashClang  bool
	openMonitor bool
	flashPort   string
)

var flashCmd = &cobra.Command{
	Use:   "flash",
	Short: "Flash the ESP-IDF project to device",
	Long:  `Flash the built project to the connected ESP32 device and optionally open serial monitor`,
	Args:  cobra.NoArgs,
	Example: `  idfmgr flash
  idfmgr flash --clang
  idfmgr flash --monitor
  idfmgr flash --clang --monitor --port /dev/ttyUSB0`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := flashProject(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flashing project: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	flashCmd.Flags().BoolVar(&flashClang, "clang", false, "Flash the Clang build")
	flashCmd.Flags().BoolVarP(&openMonitor, "monitor", "m", false, "Open serial monitor after flashing")
	flashCmd.Flags().StringVarP(&flashPort, "port", "p", "", "Serial port (auto-detected if not specified)")
	rootCmd.AddCommand(flashCmd)
}

func flashProject() error {
	versionFile := ".espidf-version"
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		return fmt.Errorf(".espidf-version file not found. Are you in an ESP-IDF project directory?")
	}

	versionData, err := os.ReadFile(versionFile)
	if err != nil {
		return fmt.Errorf("failed to read .espidf-version: %w", err)
	}
	version := string(versionData)
	version = version[:len(version)-1]

	buildDir := "build"
	if flashClang {
		buildDir = "build-clang"
		fmt.Println("Flashing Clang build...")
	} else {
		fmt.Println("Flashing GCC build...")
	}

	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		buildCommand := "idfmgr build"
		if flashClang {
			buildCommand = "idfmgr build --clang"
		}
		return fmt.Errorf("%s build directory not found. Build first with: %s", buildDir, buildCommand)
	}

	idfPath := filepath.Join(getESPBase(), version)
	if _, err := os.Stat(idfPath); os.IsNotExist(err) {
		return fmt.Errorf("ESP-IDF version %s is not installed. Install it with: idfmgr install %s", version, version)
	}

	env, err := getESPIDFEnvironment(idfPath)
	if err != nil {
		return fmt.Errorf("failed to setup ESP-IDF environment: %w", err)
	}

	idfPyPath := filepath.Join(idfPath, "tools", "idf.py")
	cmdArgs := []string{idfPyPath, "-B", buildDir}

	if flashPort != "" {
		cmdArgs = append(cmdArgs, "-p", flashPort)
	}

	if openMonitor {
		cmdArgs = append(cmdArgs, "flash", "monitor")
		fmt.Println("Flashing and opening serial monitor...")
	} else {
		cmdArgs = append(cmdArgs, "flash")
	}

	cmd := exec.Command("python3", cmdArgs...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("flash failed: %w", err)
	}

	if !openMonitor {
		fmt.Println("Flash successful!")
		fmt.Println("Tip: Use 'idfmgr flash --monitor' to open serial monitor after flashing")
	}

	return nil
}

