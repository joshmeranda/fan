package cmd

import (
	"time"
)

type Config struct {
	DefaultInvalidateAfter time.Duration
	CacheDir               string
	Aliases                map[string]string
}

func DefaultConfig() Config {
	return Config{
		DefaultInvalidateAfter: time.Hour * 24 * 7, // ~1 week
		CacheDir:               DefaultCachePath(),
		Aliases:                make(map[string]string, 0),
	}
}
