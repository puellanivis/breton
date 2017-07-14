package files

import (
	"context"
	"net/url"
	"os"
)

type localFS struct{}

// Local implements a wrapper from the os functions Open, Create, and Readdir, to the files.FileStore implementation.
var Local FileStoreWithOptions = &fsWrapper{fs: &localFS{}}

func init() {
	RegisterScheme(Local, "file")
}

// Open opens up a local filesystem file specified in the uri.Path for reading.
func (h *localFS) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	return os.Open(uri.Path)
}

// Create opens up a local filesystem file specified in the uri.Path for writing. It will create a new one if it does not exist.
func (h *localFS) Create(ctx context.Context, uri *url.URL) (Writer, error) {
	return os.Create(uri.Path)
}

// List returns the whole slice of os.FileInfos for a specific local filesystem at uri.Path.
func (h *localFS) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	f, err := os.Open(uri.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}
