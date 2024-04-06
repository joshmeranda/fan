package main

import (
	"time"
)

type Config struct {
	DefaultInvalidateAfter time.Duration
	CacheDir               string
	Aliases                map[string]string
}
