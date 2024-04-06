package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultTargetExecutableFile = "executable"
	DefaultTargetMetadataFile   = "metadata"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type Cache struct {
	CacheDir string
}

func (c *Cache) pathForTarget(target Target) string {
	return path.Join(c.CacheDir, fmt.Sprintf("%d", target.Hash()))
}

// AddTarget adds the given target to the cache, using path as the on-disk executable location.
func (c *Cache) AddTarget(target Target) error {
	path := c.pathForTarget(target)
	executablePath := filepath.Join(path, DefaultTargetExecutableFile)
	metadataPath := filepath.Join(path, DefaultTargetMetadataFile)

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed creating cache location for dir: %w", err)
	}

	if err := os.Rename(target.Path, executablePath); err != nil {
		return fmt.Errorf("failed moving target executable to cache: %w", err)
	}

	target.CachedAt = time.Now().UTC()

	// No need to store path since it is deterministic
	target.Path = ""

	out, err := yaml.Marshal(target)
	if err != nil {
		return fmt.Errorf("failed marshalling target metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, out, 0o644); err != nil {
		return fmt.Errorf("failed writing target metadata to cache: %w", err)
	}

	return fmt.Errorf("not implemented")
}

func (c *Cache) GetTargetForUrl(u string) (Target, error) {
	target := Target{Url: u}
	path := c.pathForTarget(target)

	if exists, err := Exists(path); err != nil {
		return Target{}, fmt.Errorf("failed checking for cached target: %w", err)
	} else if !exists {
		return Target{}, ErrNotFound
	}

	metadataPath := filepath.Join(path, DefaultTargetMetadataFile)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return Target{}, fmt.Errorf("failed reading target metadata: %w", err)
	}

	if err := yaml.Unmarshal(data, &target); err != nil {
		return Target{}, fmt.Errorf("failed unmarshalling target metadata: %w", err)
	}

	target.Path = filepath.Join(path, DefaultTargetExecutableFile)

	return target, nil
}

func (c *Cache) cleanTargetDir(dir string) error {
	metadataPath := filepath.Join(dir, DefaultTargetMetadataFile)

	var target Target
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed reading target metadata: %w", err)
	}

	if err := yaml.Unmarshal(data, &target); err != nil {
		return fmt.Errorf("failed unmarshalling target metadata: %w", err)
	}

	invalidAfter := target.CachedAt.Add(target.InvalidateAfter)

	if time.Now().After(invalidAfter) {
		if err := os.RemoveAll(dir); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unable to clean target from cache")
		}
	}

	return nil
}

func (c *Cache) Clean() error {
	files, err := os.ReadDir(c.CacheDir)
	if err != nil {
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
