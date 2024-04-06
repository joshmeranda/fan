package main

import "time"

type Config struct {
	DefaultInvalidateAfter time.Duration
	CacheDir               string
	Aliases                []Target
}

func (c *Config) GetTargetForAlias(alias string) (Target, bool) {
	for _, target := range c.Aliases {
		if target.Alias == alias {
			return target, true
		}
	}

	return Target{}, false
}
