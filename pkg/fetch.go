package fan

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
)

const suffixCharset = "abcdefhijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomSuffix(l int) string {
	suffix := make([]byte, l)
	for i := range suffix {
		suffix[i] = suffixCharset[rand.Intn(len(suffixCharset))]
	}
	return string(suffix)
}

// todo: check content-type header
// todo: add authentication stuff (certs)
func FetchToPath(u string, path string) (string, error) {
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer out.Close()

	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || 400 <= resp.StatusCode {
		return "", fmt.Errorf("received failed status code %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	return out.Name(), nil
}

func Fetch(u string) (string, error) {
	path := filepath.Join(os.TempDir(), "fan-"+randomSuffix(8))
	return FetchToPath(u, path)
}
