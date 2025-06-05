package daemonize

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

const (
	// DaemonChildEnv is the environment variable used to identify child processes
	DaemonChildEnv = "DAEMON_CHILD"
	// DaemonChildValue is the value set for the daemon child environment variable
	DaemonChildValue = "1"
)

// Config holds configuration options for daemonization
type Config struct {
	// WorkingDir sets the working directory for the daemon process
	// Default: "/" on Unix, current directory on Windows
	WorkingDir string

	// RedirectStreams controls whether stdin/stdout/stderr are redirected to /dev/null
	// Only applies to Unix-like systems. Default: true
	RedirectStreams bool

	// Env allows you to pass additional environment variables to the daemon process
	Env []string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		WorkingDir:      getDefaultWorkingDir(),
		RedirectStreams: true,
		Env:             []string{},
	}
}

// Start daemonizes the current process if it's not already running as a daemon
// It returns true if the process should continue (i.e., it's the daemon child),
// false if it should exit (i.e., it's the parent process)
func Start() bool {
	return StartWithConfig(DefaultConfig())
}

// StartWithConfig daemonizes the current process with custom configuration
// It returns true if the process should continue (i.e., it's the daemon child),
// false if it should exit (i.e., it's the parent process)
func StartWithConfig(config *Config) bool {
	// Check if we're already running as the daemon child
	if os.Getenv(DaemonChildEnv) == DaemonChildValue {
		setupDaemonProcess(config)
		return true
	}

	// Fork the process to create the daemon
	if err := forkDaemon(config); err != nil {
		// If forking fails, continue as foreground process
		return true
	}

	// Parent process should exit
	return false
}

// IsDaemon returns true if the current process is running as a daemon child
func IsDaemon() bool {
	return os.Getenv(DaemonChildEnv) == DaemonChildValue
}

// forkDaemon creates a child process configured as a daemon
func forkDaemon(config *Config) error {
	// Prepare the command to start the child process
	cmd := exec.Command(os.Args[0], os.Args[1:]...)

	// Set up environment variables
	env := append(os.Environ(), DaemonChildEnv+"="+DaemonChildValue)
	env = append(env, config.Env...)
	cmd.Env = env

	// Configure platform-specific process attributes
	setupDaemonCommand(cmd, config)

	// Start the child process
	if err := cmd.Start(); err != nil {
		return err
	}

	// Exit the parent process
	os.Exit(0)
	return nil // This line will never be reached
}

// setupDaemonCommand configures the command for daemon execution
func setupDaemonCommand(cmd *exec.Cmd, config *Config) {
	if runtime.GOOS == "windows" {
		// Windows: Detach from parent process
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		if config.RedirectStreams {
			if devNull, err := os.OpenFile("/dev/null", os.O_RDWR, 0); err == nil {
				cmd.Stdin = devNull
				cmd.Stdout = devNull
				cmd.Stderr = devNull
			}
		}
	}
}

// setupDaemonProcess performs additional setup for the daemon child process
func setupDaemonProcess(config *Config) {
	if runtime.GOOS != "windows" && config.WorkingDir != "" {
		// Change working directory
		os.Chdir(config.WorkingDir)
	}
}

// getDefaultWorkingDir returns the default working directory for daemons
func getDefaultWorkingDir() string {
	if runtime.GOOS == "windows" {
		// On Windows, keep the current directory
		if wd, err := os.Getwd(); err == nil {
			return wd
		}
		return ""
	}
	// On Unix-like systems, use root directory
	return "/"
}
