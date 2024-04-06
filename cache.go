package main

import "fmt"

type Cache struct {
	CacheDir string
}

func (c *Cache) AddTarget(target Target, path string) error {
	return fmt.Errorf("not implemented")
}

func (c *Cache) GetTargetPath(targe Target) (string, error) {
	return "", nil
}

func (c *Cache) CleanTarget(target Target) error {
	return fmt.Errorf("not implemented")
}

func (c *Cache) Clean() error {
	return fmt.Errorf("not implemented")
}
