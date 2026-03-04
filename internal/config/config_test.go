package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("return defaults when config file is missing", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		t.Setenv(EnvConfigPath, "")

		got, err := Load()

		require.NoError(t, err)
		assert.Equal(t, Default(), got)
	})

	t.Run("load config from xdg path", func(t *testing.T) {
		xdgHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdgHome)
		t.Setenv(EnvConfigPath, "")

		path := filepath.Join(xdgHome, "fly", "config.toml")
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(`version = 1

[clone]
default_directory = "/work/repos"
group_by_owner = true

[sources]
enabled = ["remote", "zoxide", "local"]

[picker]
title = "Pick a repo"
prompt_marker = "repo> "
window_position = "bottom"
matcher = "substring"
sorter = "minipick"
`), 0o644))

		got, err := Load()

		require.NoError(t, err)
		assert.Equal(t, 1, got.Version)
		assert.Equal(t, "/work/repos", got.Clone.DefaultDirectory)
		assert.True(t, got.Clone.GroupByOwner)
		assert.Equal(t, []string{"remote", "zoxide", "local"}, got.Sources.Enabled)
		assert.Equal(t, "Pick a repo", got.Picker.Title)
		assert.Equal(t, "repo> ", got.Picker.PromptMarker)
		assert.Equal(t, "bottom", got.Picker.WindowPosition)
		assert.Equal(t, "substring", got.Picker.Matcher)
		assert.Equal(t, "minipick", got.Picker.Sorter)
	})

	t.Run("load config from env override path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		t.Setenv(EnvConfigPath, path)
		require.NoError(t, os.WriteFile(path, []byte(`version = 1

[clone]
default_directory = "./repos"
group_by_owner = false

[sources]
enabled = ["zoxide"]

[picker]
window_position = "top"
matcher = "orderedchars"
sorter = "signal_minipick"
`), 0o644))

		got, err := Load()

		require.NoError(t, err)
		assert.Equal(t, "./repos", got.Clone.DefaultDirectory)
		assert.False(t, got.Clone.GroupByOwner)
		assert.Equal(t, []string{"zoxide"}, got.Sources.Enabled)
	})

	t.Run("return error for unknown key", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		t.Setenv(EnvConfigPath, path)
		require.NoError(t, os.WriteFile(path, []byte(`version = 1

[sources]
enabled = ["zoxide"]

[picker]
window_position = "top"
matcher = "orderedchars"
sorter = "signal_minipick"
unexpected = true
`), 0o644))

		_, err := Load()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown config keys")
	})

	t.Run("return error for invalid source", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		t.Setenv(EnvConfigPath, path)
		require.NoError(t, os.WriteFile(path, []byte(`version = 1

[sources]
enabled = ["zoxide", "invalid"]

[picker]
window_position = "top"
matcher = "orderedchars"
sorter = "signal_minipick"
`), 0o644))

		_, err := Load()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported source")
	})
}

func TestPickerOptions(t *testing.T) {
	t.Run("build picker options from defaults", func(t *testing.T) {
		_, err := Default().PickerOptions(io.Discard)

		require.NoError(t, err)
	})

	t.Run("return error for invalid picker config", func(t *testing.T) {
		cfg := Default()
		cfg.Picker.Matcher = "unknown"

		_, err := cfg.PickerOptions(io.Discard)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported picker.matcher")
	})
}
