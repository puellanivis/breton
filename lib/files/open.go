package files

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
)

// Open takes a Context and a filename (which may be a URL) and returns a
// files.Reader which will read the contents of that filename or URL. All
// errors and reversion functions returned by Option arguments are discarded.
func Open(ctx context.Context, filename string, options ...Option) (Reader, error) {
	f, err := open(ctx, filename)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		_, _ = opt(f)
	}

	return f, nil
}

func open(ctx context.Context, filename string) (Reader, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		return os.Stdin, nil
	}

	if uri, err := url.Parse(filename); err == nil {
		if root, ok := getRoot(ctx); ok {
			uri = root.ResolveReference(uri)
		}

		if fs, ok := getFS(uri); ok {
			return fs.Open(ctx, uri)
		}
	}

	return os.Open(filename)
}

// List takes a Context and a filename (which may be a URL) and returns a list
// of os.FileInfo that describes the files contained in the directory or listing.
func List(ctx context.Context, filename string) ([]os.FileInfo, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		return os.Stdin.Readdir(0)
	}

	if uri, err := url.Parse(filename); err == nil {
		if root, ok := getRoot(ctx); ok {
			uri = root.ResolveReference(uri)
		}

		if fs, ok := getFS(uri); ok {
			return fs.List(ctx, uri)
		}
	}

	return ioutil.ReadDir(filename)
}
