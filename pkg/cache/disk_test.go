package cache_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	fan "github.com/joshmeranda/fan/pkg"
	"github.com/joshmeranda/fan/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestAddTarget(t *testing.T) {
	cacheDir := t.TempDir()
	cache := cache.NewDiskCache(cacheDir)

	t.Cleanup(func() {
		if err := os.Remove(cacheDir); err != nil {
			t.Logf("failed to remove cache directory: %s", err)
		}
	})

	t.Run("GetNonExistantTarget", func(t *testing.T) {
		target, executable, err := cache.GetTargetForUrl("http://exapmle.com")
		assert.Zero(t, target)
		assert.Zero(t, executable)
		assert.EqualError(t, err, "not found")
	})

	t.Run("CanAddAndGetTarget", func(t *testing.T) {
		f, err := os.CreateTemp("", strings.Replace(t.Name()+"-executable-*", "/", "-", -1))
		if err != nil {
			t.Fatalf("failed to create file: %s", err)
		}

		target := fan.Target{
			Url:             "https://example.com",
			InvalidateAfter: time.Hour * 1,
		}

		err = cache.AddTarget(target, f.Name())
		assert.NoError(t, err)

		target, executable, err := cache.GetTargetForUrl(target.Url)

		assert.WithinDuration(t, time.Now(), target.CachedAt, time.Second*1)
		target.CachedAt = time.Time{}

		assert.Equal(t, fan.Target{
			Url:             "https://example.com",
			InvalidateAfter: time.Hour * 1,
		}, target)
		assert.Equal(t, filepath.Join(cacheDir, fmt.Sprint(target.Hash()), "example.com"), executable)
		assert.NoError(t, err)
	})

	t.Run("FailsToAddDuplicateTarget", func(t *testing.T) {

		f, err := os.CreateTemp("", strings.Replace(t.Name()+"-executable-*", "/", "-", -1))
		if err != nil {
			t.Fatalf("failed to create file: %s", err)
		}

		target := fan.Target{
			Url:             "https://example.com",
			InvalidateAfter: time.Hour * 1,
		}

		err = cache.AddTarget(target, f.Name())
		assert.NoError(t, err)

		err = cache.AddTarget(target, f.Name())
		assert.Error(t, err)
	})
}
