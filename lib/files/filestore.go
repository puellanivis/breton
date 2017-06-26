package files

import (
	"context"
	"net/url"
	"os"
	"sort"
	"sync"
)

type FileStore interface {
	Open(ctx context.Context, filename *url.URL) (Reader, error)
	Create(ctx context.Context, filename *url.URL) (Writer, error)
	List(ctx context.Context, prefix *url.URL) ([]os.FileInfo, error)
}

var fsMap struct {
	sync.Mutex

	m map[string]FileStore
	keys []string
	sorted bool
}

func getFS(uri *url.URL) (FileStore, bool) {
	fsMap.Lock()
	defer fsMap.Unlock()

	if len(uri.Scheme) <= localDriveLength {
		return Local, true
	}

	if fsMap.m == nil {
		return nil, false
	}

	fs, ok := fsMap.m[uri.Scheme]
	return fs, ok
}

func RegisterScheme(fs FileStore, schemes ...string) {
	fsMap.Lock()
	defer fsMap.Unlock()

	if fsMap.m == nil {
		fsMap.m = make(map[string]FileStore)
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

func RegisteredSchemes() []string {
	fsMap.Lock()
	defer fsMap.Unlock()

	if !fsMap.sorted {
		sort.Strings(fsMap.keys)
		fsMap.sorted = true
	}

	return fsMap.keys
}
