package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func createTemp() (*os.File, error) {
	path := filepath.Join(os.TempDir(), "abcdefg")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	return f, nil
}

// todo: check content-type header
// todo: add authentication stuff (certs)
func FetchToPath(u string) (string, error) {
	out, err := createTemp()
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	return out.Name(), nil
}
