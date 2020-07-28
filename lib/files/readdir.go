package files

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
)

// ReadDirFS defines an extension interface on FS, which also provides an ability to enumerate files given a prefix.
type ReadDirFS interface {
	FS
	ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error)
}

// ReadDir takes a context and a filename (which may be a URL) and
// returns a slice of `os.FileInfo` that describes the files contained in the directory or listing.
func ReadDir(ctx context.Context, filename string) ([]os.FileInfo, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		return os.Stdin.Readdir(0)
	}

	uri := parsePath(ctx, filename)
	if isPath(uri) {
		return ioutil.ReadDir(filename)
	}

	fsys, ok := getFS(uri)
	if !ok {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  ErrNotSupported,
		}
	}

	switch fsys := fsys.(type) {
	case ReadDirFS:
		return fsys.ReadDir(ctx, uri)

	case FileStore:
		return fsys.List(ctx, uri)
	}

	return nil, &os.PathError{
		Op:   "readdir",
		Path: filename,
		Err:  ErrNotSupported,
	}
}

// List takes a context and a filename (which may be a URL) and
// returns a slice of `os.FileInfo` that describes the files contained in the directory or listing.
//
// DEPRECATED: use `ReadDir`.
func List(ctx context.Context, filename string) ([]os.FileInfo, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		return os.Stdin.Readdir(0)
	}

	uri := parsePath(ctx, filename)
	if isPath(uri) {
		return ioutil.ReadDir(filename)
	}

	fsys, ok := getFS(uri)
	if !ok {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  ErrNotSupported,
		}
	}

	switch fsys := fsys.(type) {
	case ReadDirFS:
		return fsys.ReadDir(ctx, uri)

	case FileStore:
		return fsys.List(ctx, uri)
	}

	return nil, &os.PathError{
		Op:   "readdir",
		Path: filename,
		Err:  ErrNotSupported,
	}
}
