package files

import (
	"context"
	"net/url"
	"os"
)

type LocalFS struct{}

var Local FileStore = &LocalFS{}

func init() {
	RegisterScheme(Local, "file")
}

func (_ *LocalFS) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	return os.Open(uri.Path)
}

func (_ *LocalFS) Create(ctx context.Context, uri *url.URL) (Writer, error) { 
	return os.Create(uri.Path)
}

func (_ *LocalFS) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	f, err := os.Open(uri.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}
