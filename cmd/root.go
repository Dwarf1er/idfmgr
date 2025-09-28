package cmd

import (
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "esp-devkit",
	Short: "ESP32 development toolkit",
	Long:  `A unified CLI tool for ESP32 development with dual toolchain support`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}