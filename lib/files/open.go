package files

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

// Open returns a files.Reader, which can be used to read content from the resource at the given URL.
//
// All errors and reversion functions returned by Option arguments are discarded.
func Open(ctx context.Context, url string, options ...Option) (Reader, error) {
	f, err := open(ctx, url)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		_, _ = opt(f)
	}

	return f, nil
}

func open(ctx context.Context, resource string) (Reader, error) {
	switch resource {
	case "", "-", "/dev/stdin":
		return os.Stdin, nil
	}

	if filepath.IsAbs(resource) {
		return os.Open(resource)
	}

	if uri, err := url.Parse(resource); err == nil {
		uri = resolveFilename(ctx, uri)

		if fs, ok := getFS(uri); ok {
			return fs.Open(ctx, uri)
		}
	}

	return os.Open(resource)
}

// ReadDir reads the directory or listing of the resource at the given URL, and
// returns a slice of os.FileInfo, which describe the files contained in the directory.
func ReadDir(ctx context.Context, url string) ([]os.FileInfo, error) {
	return readDir(ctx, url)
}

func readDir(ctx context.Context, resource string) ([]os.FileInfo, error) {
	switch resource {
	case "", "-", "/dev/stdin":
		return os.Stdin.Readdir(0)
	}

	if filepath.IsAbs(resource) {
		return ioutil.ReadDir(resource)
	}

	if uri, err := url.Parse(resource); err == nil {
		uri = resolveFilename(ctx, uri)

		if fs, ok := getFS(uri); ok {
			return fs.List(ctx, uri)
		}
	}

	return ioutil.ReadDir(resource)
}

// List reads the directory or listing of the resource at the given URL, and
// returns a slice of os.FileInfo, which describe the files contained in the directory.
//
// Depcrecated: Use `ReadDir`.
func List(ctx context.Context, url string) ([]os.FileInfo, error) {
	return readDir(ctx, url)
}
