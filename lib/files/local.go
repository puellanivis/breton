package files

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
)

type localFS struct{}

// Local implements a wrapper from os.Open, os.Create, and os.Readdir, to the files.FS implementation.
var Local FS = localFS{}

func init() {
	RegisterScheme(Local, "file")
}

// Open opens up a local filesystem file specified in the uri.Path for reading.
func (localFS) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	name, err := resolveFileURL(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	return os.Open(name)
}

// Create opens up a local filesystem file specified in the uri.Path for writing. It will create a new one if it does not exist.
func (localFS) Create(ctx context.Context, uri *url.URL) (Writer, error) {
	name, err := resolveFileURL(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	return os.Create(name)
}

// List returns the whole slice of os.FileInfos for a specific local filesystem at uri.Path.
func (localFS) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	name, err := resolveFileURL(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  err,
		}
	}

	return ioutil.ReadDir(name)
}
