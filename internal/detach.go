package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// ChildEnvVar marks a process spawned by fly so shell integration can skip
// startup work that would otherwise recurse back into fly.
const ChildEnvVar = "FLY_CHILD"

// NewIsolatedCommand builds a command in its own session. Without this an
// interactive shell started as a backend takes the controlling terminal from
// the picker via tcsetpgrp and never gives it back when killed.
func NewIsolatedCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Env = append(cmd.Environ(), ChildEnvVar+"=1")

	return cmd
}

// StartDetached launches name in its own session with every standard stream on
// os.DevNull and does not wait for it to finish.
func StartDetached(name string, args ...string) error {
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w", os.DevNull, err)
	}
	defer devNull.Close()

	cmd := exec.Command(name, args...)
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Env = append(cmd.Environ(), ChildEnvVar+"=1")

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	return cmd.Process.Release()
}
