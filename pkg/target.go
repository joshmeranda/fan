package fan

import (
	"net/url"
	"strings"
	"time"

	"github.com/cespare/xxhash"
)

const (
	defaultTargetExecutableFile = "executable"
)

type Target struct {
	Url string `yaml:"url"`

	// InvalidateAfter is the amount of time the target should remain in the cache before being removed.
	InvalidateAfter time.Duration `yaml:"invalidate_after"`

	CachedAt time.Time `yaml:"cached_at"`
}

func (t Target) ExecutableName() string {
	u, err := url.Parse(t.Url)
	if err != nil {
		return defaultTargetExecutableFile
	}

	switch {
	case u.Path != "":
		components := strings.Split(u.Path, "/")
		return components[len(components)-1]
	case u.Host != "":
		components := strings.Split(u.Host, ":")
		return components[0]
	default:
		return defaultTargetExecutableFile
	}
}

func (t Target) Hash() uint64 {
	h := xxhash.New()

	h.Write([]byte(t.Url))

	return h.Sum64()
}
