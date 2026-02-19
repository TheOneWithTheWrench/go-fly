package internal_test

import (
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

func TestConfigDir(t *testing.T) {
	var (
		appName = "fly"
		newSut  = func() func(string) (string, error) { return internal.ConfigDir }
	)

	t.Run("uses XDG_CONFIG_HOME when set", func(t *testing.T) {
		var (
			sut  = newSut()
			base = t.TempDir()
		)

		t.Setenv("XDG_CONFIG_HOME", base)
		t.Setenv("HOME", "")

		got, err := sut(appName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := filepath.Join(base, appName)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("falls back to home when XDG_CONFIG_HOME not set", func(t *testing.T) {
		var (
			sut  = newSut()
			home = t.TempDir()
		)

		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", home)

		got, err := sut(appName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := filepath.Join(home, ".config", appName)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

func TestCacheDir(t *testing.T) {
	var (
		appName = "fly"
		newSut  = func() func(string) (string, error) { return internal.CacheDir }
	)

	t.Run("uses XDG_CACHE_HOME when set", func(t *testing.T) {
		var (
			sut  = newSut()
			base = t.TempDir()
		)

		t.Setenv("XDG_CACHE_HOME", base)
		t.Setenv("HOME", "")

		got, err := sut(appName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := filepath.Join(base, appName)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("falls back to home when XDG_CACHE_HOME not set", func(t *testing.T) {
		var (
			sut  = newSut()
			home = t.TempDir()
		)

		t.Setenv("XDG_CACHE_HOME", "")
		t.Setenv("HOME", home)

		got, err := sut(appName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := filepath.Join(home, ".cache", appName)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

func TestDataDir(t *testing.T) {
	var (
		appName = "fly"
		newSut  = func() func(string) (string, error) { return internal.DataDir }
	)

	t.Run("uses XDG_DATA_HOME when set", func(t *testing.T) {
		var (
			sut  = newSut()
			base = t.TempDir()
		)

		t.Setenv("XDG_DATA_HOME", base)
		t.Setenv("HOME", "")

		got, err := sut(appName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := filepath.Join(base, appName)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("falls back to home when XDG_DATA_HOME not set", func(t *testing.T) {
		var (
			sut  = newSut()
			home = t.TempDir()
		)

		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("HOME", home)

		got, err := sut(appName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := filepath.Join(home, ".local", "share", appName)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}
