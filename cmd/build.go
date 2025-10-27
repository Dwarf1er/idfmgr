package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	useClang bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the ESP-IDF project",
	Long:  `Build the current project using either GCC (default) or Clang toolchain`,
	Args:  cobra.NoArgs,
	Example: `  idfmgr build
  idfmgr build --clang`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := buildProject(); err != nil {
			fmt.Fprintf(os.Stderr, "Error building project: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	buildCmd.Flags().BoolVar(&useClang, "clang", false, "Build with Clang toolchain")
	rootCmd.AddCommand(buildCmd)
}

func buildProject() error {
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

	idfPath := filepath.Join(getESPBase(), version)
	if _, err := os.Stat(idfPath); os.IsNotExist(err) {
		return fmt.Errorf("ESP-IDF version %s is not installed. Install it with: idfmgr install %s", version, version)
	}

	env, err := getESPIDFEnvironment(idfPath)
	if err != nil {
		return fmt.Errorf("failed to setup ESP-IDF environment: %w", err)
	}

	idfPyPath := filepath.Join(idfPath, "tools", "idf.py")
	
	var cmdArgs []string
	var buildDir string

	if useClang {
		buildDir = "build-clang"
		fmt.Println("Building with Clang toolchain...")
		cmdArgs = []string{
			idfPyPath,
			"-DIDF_TOOLCHAIN=clang",
			"-B", buildDir,
			"build",
		}
	} else {
		buildDir = "build"
		fmt.Println("Building with GCC toolchain...")
		cmdArgs = []string{
			idfPyPath,
			"build",
		}
	}

	cmd := exec.Command("python3", cmdArgs...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✅ Build successful! Output in %s/\n", buildDir)
	return nil
}