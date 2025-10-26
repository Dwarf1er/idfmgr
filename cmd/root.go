package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "idfmgr",
	Short: "Manage ESP-IDF installations, versions, and projects",
	Long:  `idfmgr simplifies ESP32 development by managing multiple ESP-IDF versions, creating projects with templates, and supporting both GCC and Clang toolchain.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}

