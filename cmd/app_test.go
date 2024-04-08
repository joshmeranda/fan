package cmd_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cespare/xxhash"
	"github.com/joshmeranda/fan/cmd"
	"github.com/joshmeranda/fan/pkg/cache"
	"github.com/phayes/freeport"
	"gopkg.in/yaml.v3"
)

func setup(t *testing.T) (string, string, string) {
	http.HandleFunc("/script", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "#!/usr/bin/bash\nexit 0")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	})

	port, err := freeport.GetFreePort()
	if err != nil {
		t.Fatalf("could not determine free port: %s", err)
	}

	server := http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	go func() {
		err = server.ListenAndServe()
	}()

	for {
		if err != nil {
			t.Fatalf("could not start server: %s", err)
		}

		_, healthErr := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
		if healthErr == nil {
			t.Log("server is up and running")
			break
		}
	}

	configPath := fmt.Sprintf("%s.config", t.Name())
	config := cmd.Config{
		DefaultInvalidateAfter: time.Second * 1,
		CacheDir:               fmt.Sprintf("%s.cache", t.Name()),
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("could not marshal config: %s", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("could not write config file: %s", err)
	}

	t.Cleanup(func() {
		server.Close()
		os.RemoveAll(config.CacheDir)
		os.Remove(configPath)
	})

	return server.Addr, configPath, config.CacheDir
}

func Exists(t *testing.T, path string) bool {
	t.Helper()

	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}

func TestMain(t *testing.T) {
	addr, configPath, cacheDir := setup(t)

	hash := xxhash.New()
	hash.Write([]byte("http://" + addr + "/script"))
	sum := hash.Sum64()

	targetCacheDir := path.Join(cacheDir, fmt.Sprintf("%d", sum))

	app := cmd.App()

	t.Run("Cache is empty", func(t *testing.T) {
		if Exists(t, cacheDir) {
			t.Fatalf("cache dir should exist (yet): %s", targetCacheDir)
		}
	})

	t.Run("No cache", func(t *testing.T) {
		if err := app.Run([]string{"fan", "--config", configPath, "run", fmt.Sprintf("http://%s/script", addr)}); err != nil {
			t.Fatalf("app failed with error: %s", err)
		}

		if !Exists(t, targetCacheDir) {
			t.Fatalf("cache dir does not exist: %s", targetCacheDir)
		}

		if !Exists(t, path.Join(targetCacheDir, cache.DefaultTargetExecutableFile)) {
			t.Fatalf("executable does not exist in cache")
		}

		if !Exists(t, path.Join(targetCacheDir, cache.DefaultTargetMetadataFile)) {
			t.Fatalf("metadata does not exist in cache")
		}
	})

	t.Run("Cached", func(t *testing.T) {
		if err := app.Run([]string{"fan", "--config", configPath, "run", fmt.Sprintf("http://%s/script", addr)}); err != nil {
			t.Fatalf("app failed with error: %s", err)
		}
	})

	t.Run("Aliased", func(t *testing.T) {
		if err := app.Run([]string{"fan", "--config", configPath, "alias", "script", fmt.Sprintf("http://%s/script", addr)}); err != nil {
			t.Fatalf("failed to add alias: %s", err)
		}

		if err := app.Run([]string{"fan", "--config", configPath, "run", "script"}); err != nil {
			t.Fatalf("app failed with error: %s", err)
		}
	})

	t.Run("cache clean", func(t *testing.T) {
		time.Sleep(time.Second * 1)
		if err := app.Run([]string{"fan", "--config", configPath, "cache", "clean"}); err != nil {
			t.Fatalf("app failed with error: %s", err)
		}

		if Exists(t, targetCacheDir) {
			t.Fatalf("cache dir should not exist: %s", targetCacheDir)
		}
	})
}
