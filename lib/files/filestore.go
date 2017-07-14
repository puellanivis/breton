package files

import (
	"context"
	"net/url"
	"os"
	"sort"
	"sync"
)

// FileStore defines an interface which implements a system of accessing files for reading (Open) writing (Write) and directly listing (List)
type FileStore interface {
	Open(ctx context.Context, uri *url.URL) (Reader, error)
	Create(ctx context.Context, uri *url.URL) (Writer, error)
	List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error)
}

// FileStoreWithOptions defines an interface which implements a system of accessing files for reading (Open) writing (Write) and directly listing (List)
type FileStoreWithOptions interface {
	Open(ctx context.Context, uri *url.URL, options ...Option) (Reader, error)
	Create(ctx context.Context, uri *url.URL, options ...Option) (Writer, error)
	List(ctx context.Context, uri *url.URL, options ...Option) ([]os.FileInfo, error)
}

var fsMap struct {
	sync.Mutex

	m      map[string]FileStoreWithOptions
	keys   []string
	sorted bool
}

type fsWrapper struct {
	fs FileStore
}

func (fs *fsWrapper) Open(ctx context.Context, uri *url.URL, options ...Option) (Reader, error) {
	f, err := fs.fs.Open(ctx, uri)

	if err == nil && len(options) > 0 {
		return f, ErrNotSupported
	}

	return f, err
}

func (fs *fsWrapper) Create(ctx context.Context, uri *url.URL, options ...Option) (Writer, error) {
	f, err := fs.fs.Create(ctx, uri)

	if err == nil && len(options) > 0 {
		return f, ErrNotSupported
	}

	return f, err
}

func (fs *fsWrapper) List(ctx context.Context, uri *url.URL, options ...Option) ([]os.FileInfo, error) {
	list, err := fs.fs.List(ctx, uri)

	if err == nil && len(options) > 0 {
		return list, ErrNotSupported
	}

	return list, err
}

func getFS(uri *url.URL) (FileStoreWithOptions, bool) {
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

// RegisterScheme takes a FileStore and attaches to it the given schemes so
// that files.Open will use that FileStore when a files.Open() is performed
// with a URL of any of those schemes.
func RegisterScheme(fs interface{}, schemes ...string) {
	if towrap, ok := fs.(FileStore); ok {
		fs = &fsWrapper{fs: towrap}
	}

	if len(schemes) < 1 {
		return
	}

	fsMap.Lock()
	defer fsMap.Unlock()

	if fsMap.m == nil {
		fsMap.m = make(map[string]FileStoreWithOptions)
	}
	fsMap.sorted = false

	for _, scheme := range schemes {
		if _, ok := fsMap.m[scheme]; ok {
			// TODO: report duplicate scheme registration
			continue
		}

		fsMap.m[scheme] = fs.(FileStoreWithOptions)
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

	return fsMap.keys
}
