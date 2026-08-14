package internal_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIsolatedCommand(t *testing.T) {
	newSut := func() func(context.Context, string, ...string) *exec.Cmd {
		return internal.NewIsolatedCommand
	}

	t.Run("request a new session so the child cannot claim the terminal", func(t *testing.T) {
		sut := newSut()

		got := sut(context.Background(), "sleep", "1")

		require.NotNil(t, got.SysProcAttr)
		assert.True(t, got.SysProcAttr.Setsid)
	})

	t.Run("mark the child so shell integration skips startup work", func(t *testing.T) {
		sut := newSut()

		got := sut(context.Background(), "sleep", "1")

		assert.Contains(t, got.Env, internal.ChildEnvVar+"=1")
	})

	t.Run("run the child in its own session", func(t *testing.T) {
		var (
			sut = newSut()
			cmd = sut(context.Background(), "sleep", "5")
		)
		require.NoError(t, cmd.Start())
		t.Cleanup(func() { _ = cmd.Process.Kill() })

		got, err := syscall.Getsid(cmd.Process.Pid)

		require.NoError(t, err)
		assert.Equal(t, cmd.Process.Pid, got, "child should lead its own session")
		assert.NotEqual(t, sessionID(t), got)
	})

	t.Run("stop the child when the context is cancelled", func(t *testing.T) {
		var (
			ctx, cancel = context.WithCancel(context.Background())
			sut         = newSut()
			cmd         = sut(ctx, "sleep", "30")
		)
		require.NoError(t, cmd.Start())
		cancel()

		err := cmd.Wait()

		require.Error(t, err)
	})
}

func TestStartDetached(t *testing.T) {
	newSut := func() func(string, ...string) error {
		return internal.StartDetached
	}

	t.Run("run the process in its own session", func(t *testing.T) {
		var (
			pidFile = filepath.Join(t.TempDir(), "pid")
			sut     = newSut()
		)

		err := sut("/bin/sh", "-c", "echo $$ > "+pidFile+"; sleep 5")

		require.NoError(t, err)
		pid := waitForPid(t, pidFile)
		t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })

		got, sidErr := syscall.Getsid(pid)

		require.NoError(t, sidErr)
		assert.Equal(t, pid, got, "detached process should lead its own session")
		assert.NotEqual(t, sessionID(t), got)
	})

	t.Run("mark the process so shell integration skips startup work", func(t *testing.T) {
		var (
			envFile = filepath.Join(t.TempDir(), "env")
			sut     = newSut()
		)

		err := sut("/bin/sh", "-c", "printf %s \"$"+internal.ChildEnvVar+"\" > "+envFile)

		require.NoError(t, err)
		assert.Equal(t, "1", waitForContent(t, envFile))
	})

	t.Run("keep running after writing more than a pipe buffer to stdout", func(t *testing.T) {
		var (
			doneFile = filepath.Join(t.TempDir(), "done")
			sut      = newSut()
		)

		err := sut("/bin/sh", "-c", "head -c 200000 /dev/zero; echo done > "+doneFile)

		require.NoError(t, err)
		assert.Equal(t, "done", waitForContent(t, doneFile))
	})

	t.Run("return an error for an unknown executable", func(t *testing.T) {
		sut := newSut()

		err := sut(filepath.Join(t.TempDir(), "does-not-exist"))

		require.Error(t, err)
	})
}

func sessionID(t *testing.T) int {
	t.Helper()

	sid, err := syscall.Getsid(os.Getpid())
	require.NoError(t, err)

	return sid
}

func waitForPid(t *testing.T, path string) int {
	t.Helper()

	pid, err := strconv.Atoi(waitForContent(t, path))
	require.NoError(t, err)

	return pid
}

func waitForContent(t *testing.T, path string) string {
	t.Helper()

	var content string
	require.Eventually(t, func() bool {
		data, err := os.ReadFile(path)
		if err != nil {
			return false
		}

		content = strings.TrimSpace(string(data))
		return content != ""
	}, 5*time.Second, 10*time.Millisecond, "expected %s to be written", path)

	return content
}
