package fan

import (
	"time"

	"github.com/cespare/xxhash"
)

type Target struct {
	Url string `yaml:"url"`

	// InvalidateAfter is the amount of time the target should remain in the cache before being removed.
	InvalidateAfter time.Duration `yaml:"invalidate_after"`

	CachedAt time.Time `yaml:"cached_at"`

	// Path is the on-disk path to the target executable.
	Path string `yaml:"path,omitempty"`
}

func (t Target) Hash() uint64 {
	h := xxhash.New()

	h.Write([]byte(t.Url))

	return h.Sum64()
}
