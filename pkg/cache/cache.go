package cache

import (
	fan "github.com/joshmeranda/fan/pkg"
)

// Cache is the mechanism used to store and server
type Cache interface {
	AddTarget(target fan.Target, executable string) error

	// GetTargetForUrl returns the target and path to executable for the given url, or an error if one occured.
	GetTargetForUrl(url string) (fan.Target, string, error)

	InvalidateUrl(url string) error

	List() ([]fan.Target, error)

	Clean() error
}

// noopCache is a Cache implementation that does nothing, useful when caching is disabled.
type noopCache struct{}

func NewNoopCache() Cache {
	return &noopCache{}
}

func (c *noopCache) AddTarget(target fan.Target, executable string) error {
	return nil
}

func (c *noopCache) GetTargetForUrl(url string) (fan.Target, string, error) {
	return fan.Target{}, "", ErrNotFound
}

func (c *noopCache) InvalidateUrl(string) error {
	return nil
}

func (c *noopCache) List() ([]fan.Target, error) {
	return make([]fan.Target, 0), nil
}

func (c *noopCache) Clean() error {
	return nil
}
