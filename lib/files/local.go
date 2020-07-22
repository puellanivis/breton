package files

import (
	"context"
	"fmt"
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

func filename(uri *url.URL) string {
	if uri.Path != "" {
		return uri.Path
	}

	return uri.Opaque
}

// Open opens up a local filesystem file specified in the uri.Path for reading.
func (localFS) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	fmt.Println("os.Open:", filename(uri))
	return os.Open(filename(uri))
}

// Create opens up a local filesystem file specified in the uri.Path for writing. It will create a new one if it does not exist.
func (localFS) Create(ctx context.Context, uri *url.URL) (Writer, error) {
	return os.Create(filename(uri))
}

// List returns the whole slice of os.FileInfos for a specific local filesystem at uri.Path.
func (localFS) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return ioutil.ReadDir(filename(uri))
}
