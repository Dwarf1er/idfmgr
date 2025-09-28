package cmd

import (
	"os"
	"path/filepath"
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