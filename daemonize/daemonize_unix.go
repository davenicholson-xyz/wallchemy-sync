//go:build !windows

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

	// Unix-like systems
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session
	}

	if config.RedirectStreams {
		if devNull, err := os.OpenFile("/dev/null", os.O_RDWR, 0); err == nil {
			cmd.Stdin = devNull
			cmd.Stdout = devNull
			cmd.Stderr = devNull
		}
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
