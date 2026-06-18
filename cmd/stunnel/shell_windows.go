//go:build windows
// +build windows

package main

import (
	"os/exec"
)

func createShellCommand(shell string) *exec.Cmd {
	return exec.Command("cmd.exe")
}
