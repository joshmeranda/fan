package main

import "time"

type Target struct {
	Url   string
	Alias string

	// InvalidateAfter is the amount of time the target should remain in the cache before being removed.
	InvalidateAfter time.Duration
}
