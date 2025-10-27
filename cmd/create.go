package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	arduino    bool
	idfVersion string
	target     string
)

var createCmd = &cobra.Command{
	Use:   "create <project-name>",
	Short: "Create a new ESP-IDF project",
	Long:  `Create a new ESP-IDF project with either a base or arduino template`,
	Args:  cobra.ExactArgs(1),
	Example: `  idfmgr create my-project
  idfmgr create my-project --arduino
  idfmgr create my-project --version v5.1.2
  idfmgr create my-project --target esp32s3 --arduino`,
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		if err := createProject(projectName); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	createCmd.Flags().BoolVar(&arduino, "arduino", false, "Create Arduino-based project")
	createCmd.Flags().StringVar(&idfVersion, "version", "", "ESP-IDF version to use (default: latest installed)")
	createCmd.Flags().StringVarP(&target, "target", "t", "esp32", "Target chip (default: esp32)")
	rootCmd.AddCommand(createCmd)
}

func createProject(projectName string) error {
	version, _ := getLatestInstalledESPIDFVersion()
	if idfVersion != "" {
		if _, err := os.Stat(filepath.Join(getESPBase(), idfVersion)); err == nil {
			version = idfVersion
		}
	}

	idfPath := filepath.Join(getESPBase(), version)

	env, err := getESPIDFEnvironment(idfPath)
	if err != nil {
		return fmt.Errorf("failed to setup ESP-IDF environment: %w", err)
	}

	idfPyPath := filepath.Join(idfPath, "tools", "idf.py")

	cmd := exec.Command("python3", idfPyPath, "create-project", projectName)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	projectPath := filepath.Join(".", projectName)

	if err := applyCommonModifications(projectPath, version, idfPath, env); err != nil {
		return err
	}

	if arduino {
		if err := applyArduinoModifications(projectPath); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Project '%s' created successfully!\n", projectName)
	return nil
}

func applyCommonModifications(projectPath, version, idfPath string, env []string) error {
	if err := createESPIDFVersionFile(projectPath, version); err != nil {
		return fmt.Errorf("failed to create .espidf-version: %w", err)
	}

	if err := setTarget(projectPath, idfPath, env); err != nil {
		return fmt.Errorf("failed to set target: %w", err)
	}

	if err := createGitignore(projectPath); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	if err := createClangdConfig(projectPath); err != nil {
		return fmt.Errorf("failed to create .clangd: %w", err)
	}

	if err := createSdkconfigDefaults(projectPath); err != nil {
		return fmt.Errorf("failed to create sdkconfig.defaults: %w", err)
	}

	if err := modifyRootCMakeLists(projectPath); err != nil {
		return fmt.Errorf("failed to modify CMakeLists.txt: %w", err)
	}

	if err := renameMainFile(projectPath); err != nil {
		return fmt.Errorf("failed to rename main file: %w", err)
	}

	if err := initGitRepo(projectPath); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	return nil
}

func createESPIDFVersionFile(projectPath, version string) error {
	versionFile := filepath.Join(projectPath, ".espidf-version")
	content := fmt.Sprintf("%s\n", version)
	return os.WriteFile(versionFile, []byte(content), 0o644)
}

func setTarget(projectPath, idfPath string, env []string) error {
	idfPyPath := filepath.Join(idfPath, "tools", "idf.py")
	cmd := exec.Command("python3", idfPyPath, "set-target", "esp32")
	cmd.Dir = projectPath
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createGitignore(projectPath string) error {
	gitignoreContent := `.cache/
build/
build-clang/
sdkconfig
sdkconfig.old
*.bin
*.elf
*.map
managed_components/
dependencies.lock
sdkconfig
sdkconfig.old
warnings.txt
`
	gitignoreFile := filepath.Join(projectPath, ".gitignore")
	return os.WriteFile(gitignoreFile, []byte(gitignoreContent), 0o644)
}

func createClangdConfig(projectPath string) error {
	clangdContent := `CompileFlags:
  CompilationDatabase: build-clang
  Remove:
    -fno-shrink-wrap
    -fno-tree-switch-conversion
    -fstrict-volatile-bitfields
    -mtext-section-literals
    -mdisable-hardware-atomics
    -mlongcalls
`
	if arduino {
		clangdContent += `  Add:
    -I./components/arduino
`
	}

	clangdFile := filepath.Join(projectPath, ".clangd")
	return os.WriteFile(clangdFile, []byte(clangdContent), 0o644)
}

func createSdkconfigDefaults(projectPath string) error {
	sdkconfigContent := fmt.Sprintf(`# Target configuration
CONFIG_IDF_TARGET="%s"

# Other defaults
CONFIG_AUTOSTART_ARDUINO=n
CONFIG_FREERTOS_HZ=1000
`, target)

	sdkconfigFile := filepath.Join(projectPath, "sdkconfig.defaults")
	return os.WriteFile(sdkconfigFile, []byte(sdkconfigContent), 0o644)
}

func modifyRootCMakeLists(projectPath string) error {
	cmakeFile := filepath.Join(projectPath, "CMakeLists.txt")

	content, err := os.ReadFile(cmakeFile)
	if err != nil {
		return err
	}

	newContent := strings.ReplaceAll(string(content),
		fmt.Sprintf("project(%s)", filepath.Base(projectPath)),
		"project(main)")

	return os.WriteFile(cmakeFile, []byte(newContent), 0o644)
}

func renameMainFile(projectPath string) error {
	projectName := filepath.Base(projectPath)

	oldPath := filepath.Join(projectPath, "main", projectName+".c")
	newPath := filepath.Join(projectPath, "main", "main.c")

	if _, err := os.Stat(oldPath); err == nil {
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	} else if _, err := os.Stat(newPath); err != nil {
		return fmt.Errorf("could not find main source file")
	}

	mainCMakeFile := filepath.Join(projectPath, "main", "CMakeLists.txt")
	content, err := os.ReadFile(mainCMakeFile)
	if err != nil {
		return err
	}

	newContent := strings.ReplaceAll(string(content),
		projectName+".c",
		"main.c")

	return os.WriteFile(mainCMakeFile, []byte(newContent), 0o644)
}

func initGitRepo(projectPath string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func applyArduinoModifications(projectPath string) error {
	if err := addArduinoSubmodule(projectPath); err != nil {
		return fmt.Errorf("failed to add Arduino submodule: %w", err)
	}

	if err := addArduinoToRootCMakeLists(projectPath); err != nil {
		return fmt.Errorf("failed to modify root CMakeLists.txt: %w", err)
	}

	if err := addArduinoToMainCMakeLists(projectPath); err != nil {
		return fmt.Errorf("failed to modify main/CMakeLists.txt: %w", err)
	}

	if err := convertToArduinoMain(projectPath); err != nil {
		return fmt.Errorf("failed to convert to Arduino main: %w", err)
	}

	return nil
}

func addArduinoSubmodule(projectPath string) error {
	cmd := exec.Command("git", "submodule", "add",
		"https://github.com/espressif/arduino-esp32.git",
		"components/arduino")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func addArduinoToRootCMakeLists(projectPath string) error {
	cmakeFile := filepath.Join(projectPath, "CMakeLists.txt")
	content, err := os.ReadFile(cmakeFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	projectLineFound := false

	for _, line := range lines {
		if !projectLineFound && strings.Contains(line, "include($ENV{IDF_PATH}/tools/cmake/project.cmake)") {
			newLines = append(newLines, "set(EXTRA_COMPONENT_DIRS components/arduino)")
			newLines = append(newLines, line)
			projectLineFound = true
		} else {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(cmakeFile, []byte(newContent), 0o644)
}

func addArduinoToMainCMakeLists(projectPath string) error {
	mainCMakeFile := filepath.Join(projectPath, "main", "CMakeLists.txt")
	content, err := os.ReadFile(mainCMakeFile)
	if err != nil {
		return err
	}

	newContent := strings.ReplaceAll(string(content),
		"idf_component_register(SRCS",
		"idf_component_register(SRCS")

	if !strings.Contains(string(content), "REQUIRES") {
		newContent = strings.ReplaceAll(newContent,
			"INCLUDE_DIRS \".\")",
			"INCLUDE_DIRS \".\"\n                    REQUIRES arduino)")
	} else {
		newContent = strings.ReplaceAll(newContent,
			"REQUIRES",
			"REQUIRES arduino")
	}

	return os.WriteFile(mainCMakeFile, []byte(newContent), 0o644)
}

func convertToArduinoMain(projectPath string) error {
	oldPath := filepath.Join(projectPath, "main", "main.c")
	newPath := filepath.Join(projectPath, "main", "main.cpp")

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	arduinoContent := `#include "Arduino.h"

extern "C" void app_main()
{
    initArduino();
    
    // Arduino-like setup()
    Serial.begin(115200);
    while(!Serial) {
        ; // wait for serial port to connect
    }
    
    // Arduino-like loop()
    while(true) {
        Serial.println("loop");
        delay(1000);
    }
    
    // WARNING: if program reaches end of function app_main() the MCU will restart.
}
`

	if err := os.WriteFile(newPath, []byte(arduinoContent), 0o644); err != nil {
		return err
	}

	mainCMakeFile := filepath.Join(projectPath, "main", "CMakeLists.txt")
	content, err := os.ReadFile(mainCMakeFile)
	if err != nil {
		return err
	}

	newContent := strings.ReplaceAll(string(content), "main.c", "main.cpp")
	return os.WriteFile(mainCMakeFile, []byte(newContent), 0o644)
}

