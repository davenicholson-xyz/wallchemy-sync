package network

import (
	"os"
	"runtime"
)

// determinePath determines the IPC path based on configuration
func determinePath(config IPCConfig) string {
	if config.Path != "" {
		return config.Path
	}

	var defaultName string
	if config.AppName != "" {
		defaultName = config.AppName
	} else {
		defaultName = "wallchemy_sync"
	}

	if runtime.GOOS == "windows" {
		return `\\.\pipe\` + defaultName
	}
	return "/tmp/" + defaultName + ".sock"
}

// cleanupPath performs OS-specific cleanup
func cleanupPath(path string) {
	if runtime.GOOS != "windows" {
		os.Remove(path)
	}
}
