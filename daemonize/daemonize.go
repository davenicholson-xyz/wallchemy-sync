package daemonize

import (
	"os"
)

const (
	DaemonChildEnv   = "DAEMON_CHILD"
	DaemonChildValue = "1"
)

type Config struct {
	WorkingDir      string
	RedirectStreams bool
	Env             []string
}

func DefaultConfig() *Config {
	return &Config{
		WorkingDir:      getDefaultWorkingDir(),
		RedirectStreams: true,
		Env:             []string{},
	}
}

func Start() bool {
	return StartWithConfig(DefaultConfig())
}

func StartWithConfig(config *Config) bool {
	if os.Getenv(DaemonChildEnv) == DaemonChildValue {
		setupDaemonProcess(config)
		return true
	}
	if err := forkDaemon(config); err != nil {
		return true
	}
	return false
}

func IsDaemon() bool {
	return os.Getenv(DaemonChildEnv) == DaemonChildValue
}

func setupDaemonProcess(config *Config) {
	if config.WorkingDir != "" {
		os.Chdir(config.WorkingDir)
	}
}

func getDefaultWorkingDir() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return ""
}
