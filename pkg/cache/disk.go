package cache

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"

	fan "github.com/joshmeranda/fan/pkg"
	"gopkg.in/yaml.v3"
)

const (
	DefaultTargetMetadataFile = "metadata"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type diskCache struct {
	CacheDir string
}

func NewDiskCache(cacheDir string) Cache {
	return &diskCache{
		CacheDir: cacheDir,
	}
}

func (c *diskCache) pathForTarget(target fan.Target) string {
	return path.Join(c.CacheDir, fmt.Sprintf("%d", target.Hash()))
}

// AddTarget adds the given target to the cache, using path as the on-disk executable location.
func (c *diskCache) AddTarget(target fan.Target, executable string) error {
	path := c.pathForTarget(target)
	executablePath := filepath.Join(path, target.ExecutableName())
	metadataPath := filepath.Join(path, DefaultTargetMetadataFile)

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed creating cache location for dir: %w", err)
	}

	if err := os.Rename(executable, executablePath); err != nil {
		return fmt.Errorf("failed moving target executable to cache: %w", err)
	}

	target.CachedAt = time.Now().UTC()

	out, err := yaml.Marshal(target)
	if err != nil {
		return fmt.Errorf("failed marshalling target metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, out, 0o644); err != nil {
		return fmt.Errorf("failed writing target metadata to cache: %w", err)
	}

	return nil
}

func (c *diskCache) GetTargetForUrl(u string) (fan.Target, string, error) {
	target := fan.Target{Url: u}
	path := c.pathForTarget(target)

	if exists, err := PathExists(path); err != nil {
		return fan.Target{}, "", fmt.Errorf("failed checking for cached target: %w", err)
	} else if !exists {
		return fan.Target{}, "", ErrNotFound
	}

	metadataPath := filepath.Join(path, DefaultTargetMetadataFile)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fan.Target{}, "", fmt.Errorf("failed reading target metadata: %w", err)
	}

	if err := yaml.Unmarshal(data, &target); err != nil {
		return fan.Target{}, "", fmt.Errorf("failed unmarshalling target metadata: %w", err)
	}

	invalidAfter := target.CachedAt.Add(target.InvalidateAfter)

	if time.Now().UTC().After(invalidAfter) {
		if err := os.RemoveAll(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fan.Target{}, "", fmt.Errorf("unable to clean target from cache")
		}

		return fan.Target{}, "", ErrNotFound
	}

	return target, filepath.Join(path, target.ExecutableName()), nil
}

func (c *diskCache) cleanTargetDir(dir string) error {
	metadataPath := filepath.Join(dir, DefaultTargetMetadataFile)

	var target fan.Target
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed reading target metadata: %w", err)
	}

	if err := yaml.Unmarshal(data, &target); err != nil {
		return fmt.Errorf("failed unmarshalling target metadata: %w", err)
	}

	invalidAfter := target.CachedAt.Add(target.InvalidateAfter)

	if time.Now().UTC().After(invalidAfter) {
		if err := os.RemoveAll(dir); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unable to clean target from cache")
		}
	}

	return nil
}

func (c *diskCache) Clean() error {
	files, err := os.ReadDir(c.CacheDir)
	if err != nil && err.(*os.PathError).Err.(syscall.Errno) != syscall.ENOENT {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			targetPath := filepath.Join(c.CacheDir, file.Name())

			if err := c.cleanTargetDir(targetPath); err != nil {
				return fmt.Errorf("failed to clean target dir: %w", err)
			}
		}
	}

	return nil
}
