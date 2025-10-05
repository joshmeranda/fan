package cache

import (
	"fmt"

	fan "github.com/joshmeranda/fan/pkg"
)

// todo: will need to test with multiply nested runs of fan
// nestedCache works as a merger between 2 caches. One is a read-write cache, and the second is read-only. Read operations are first directed towards the read-only cache befor coming to the read-write cache.
type nestedCache struct {
	backing Cache
	store   Cache
}

func NewNestedCache(store Cache, backing Cache) Cache {
	return &nestedCache{
		backing: backing,
		store:   store,
	}
}

func (c *nestedCache) AddTarget(target fan.Target, executable string) error {
	if err := c.store.AddTarget(target, executable); err != nil {
		return fmt.Errorf("failed to store target in nested cache: %w", err)
	}

	return nil
}

func (c *nestedCache) GetTargetForUrl(u string) (fan.Target, string, error) {
	target, executable, err := c.backing.GetTargetForUrl(u)
	if err == nil {
		return target, executable, nil
	}

	target, executable, err = c.store.GetTargetForUrl(u)
	if err != nil {
		return fan.Target{}, "", err
	}

	return target, executable, nil
}

func (c *nestedCache) InvalidateUrl(url string) error {
	return c.store.InvalidateUrl(url)
}

func (c *nestedCache) Clean() error {
	return c.store.Clean()
}
