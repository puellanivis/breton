package files

import (
	"context"
	"net/url"
	"sort"
	"sync"
)

// FS defines an interface, which implements a minimal set of functionality for a filesystem from this package.
type FS interface {
	Open(ctx context.Context, uri *url.URL) (Reader, error)
}

var fsMap struct {
	sync.Mutex

	m      map[string]FS
	keys   []string
	sorted bool
}

func getFS(uri *url.URL) (FS, bool) {
	fsMap.Lock()
	defer fsMap.Unlock()

	fs, ok := fsMap.m[uri.Scheme]
	return fs, ok
}

// RegisterScheme takes an FS and attaches to it the given schemes so
// that files.Open will use that FS when a files.Open() is performed
// with a URL with any of those schemes.
func RegisterScheme(fs FS, schemes ...string) {
	if len(schemes) < 1 {
		return
	}

	fsMap.Lock()
	defer fsMap.Unlock()

	if fsMap.m == nil {
		fsMap.m = make(map[string]FS)
	}
	fsMap.sorted = false

	for _, scheme := range schemes {
		if _, ok := fsMap.m[scheme]; ok {
			// TODO: report duplicate scheme registration
			continue
		}

		fsMap.m[scheme] = fs
		fsMap.keys = append(fsMap.keys, scheme)
	}
}

// RegisteredSchemes returns a slice of strings that describe all registered schemes.
func RegisteredSchemes() []string {
	fsMap.Lock()
	defer fsMap.Unlock()

	if !fsMap.sorted {
		sort.Strings(fsMap.keys)
		fsMap.sorted = true
	}

	return append([]string(nil), fsMap.keys...)
}
