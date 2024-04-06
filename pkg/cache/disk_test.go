package cache_test

import (
	"errors"
	"os"
	"testing"
	"time"

	fan "github.com/joshmeranda/fan/pkg"
	"github.com/joshmeranda/fan/pkg/cache"
)

func setup(t *testing.T) (cache.Cache, error) {
	t.Helper()

	cacheDir := t.TempDir()

	_, err := os.Create(t.Name())
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		os.Remove(t.Name())
	})

	return cache.NewDiskCache(cacheDir), nil
}

func TestTarget(t *testing.T) {
	c, err := setup(t)
	if err != nil {
		t.Fatal("failed to setup test: %w", err)
	}

	target := fan.Target{
		Url:             "http://example.com",
		Path:            t.Name(),
		InvalidateAfter: time.Second * 1, // will be cleaned up after next call to c.Clean()
	}

	_, err = c.GetTargetForUrl(target.Url)
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("expected failure but found: %s", err)
	}

	if err := c.AddTarget(target); err != nil {
		t.Fatalf("expected success but found: %s", err)
	}

	target, err = c.GetTargetForUrl(target.Url)
	if err != nil {
		t.Fatalf("expected success but found: %s", err)
	}

	time.Sleep(time.Second)

	if err := c.Clean(); err != nil {
		t.Fatalf("expected success but found: %s", err)
	}

	_, err = c.GetTargetForUrl(target.Url)
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("expected failure but found: %s", err)
	}
}
