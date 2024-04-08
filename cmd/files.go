package cmd

import (
	"os"
	"path/filepath"
)

const (
	// CacheDir is the directory where the cache is stored.
	DefaultCacheDirName = "fan.cache"

	// ConfigFileName is the name of the configuration file.
	DefaultConfigFileName = "fan.config"
)

func DefaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return DefaultConfigFileName
	}

	return filepath.Join(dir, DefaultConfigFileName)
}

func DefaultCachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}

	return filepath.Join(dir, DefaultCacheDirName)
}
