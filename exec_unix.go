//go:build !windows
// +build !windows

package main

import "os/exec"

func createCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
