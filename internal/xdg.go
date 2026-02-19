package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	cacheEnv  = "XDG_CACHE_HOME"
	configEnv = "XDG_CONFIG_HOME"
	dataEnv   = "XDG_DATA_HOME"
)

func CacheDir(appName string) (string, error) {
	return resolveDir(cacheEnv, ".cache", appName)
}

func ConfigDir(appName string) (string, error) {
	return resolveDir(configEnv, ".config", appName)
}

func DataDir(appName string) (string, error) {
	return resolveDir(dataEnv, filepath.Join(".local", "share"), appName)
}

func resolveDir(envKey string, fallbackSuffix string, appName string) (string, error) {
	if strings.TrimSpace(appName) == "" {
		return "", fmt.Errorf("app name required")
	}

	if baseDir := os.Getenv(envKey); baseDir != "" {
		return filepath.Join(baseDir, appName), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	return filepath.Join(homeDir, fallbackSuffix, appName), nil
}
