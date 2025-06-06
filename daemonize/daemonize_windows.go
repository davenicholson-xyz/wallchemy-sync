//go:build windows

package daemonize

import (
	"os"
	"os/exec"
	"syscall"
)

func forkDaemon(config *Config) error {
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	env := append(os.Environ(), DaemonChildEnv+"="+DaemonChildValue)
	env = append(env, config.Env...)
	cmd.Env = env

	// Windows-specific process creation
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | syscall.DETACHED_PROCESS,
	}

	// Handle stream redirection on Windows
	if config.RedirectStreams {
		// Redirect to NUL (Windows equivalent of /dev/null)
		if nul, err := os.OpenFile("NUL", os.O_RDWR, 0); err == nil {
			cmd.Stdin = nul
			cmd.Stdout = nul
			cmd.Stderr = nul
		}
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
