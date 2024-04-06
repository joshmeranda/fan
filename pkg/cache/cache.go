package cache

import (
	fan "github.com/joshmeranda/fan/pkg"
)

type Cache interface {
	AddTarget(target fan.Target) error

	GetTargetForUrl(url string) (fan.Target, error)

	Clean() error
}

// noopCache is a Cache implementation that does nothing, useful when caching is disabled.
type noopCache struct{}

func NewNoopCache() Cache {
	return &noopCache{}
}

func (c *noopCache) AddTarget(target fan.Target) error {
	return nil
}

func (c *noopCache) GetTargetForUrl(url string) (fan.Target, error) {
	return fan.Target{}, ErrNotFound
}

func (c *noopCache) Clean() error {
	return nil
}
