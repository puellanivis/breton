package files

import (
	"context"
	"net/url"
	"os"
)

// Open takes a Context and a filename (which may be a URL) and returns a files.Reader which will read the contents of that filename or URL.
func Open(ctx context.Context, filename string, options ...Option) (Reader, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		if err := applyOptions(os.Stdin, options); err != nil {
			return nil, err
		}
		return os.Stdin, nil
	}

	if uri, err := url.Parse(filename); err == nil {
		if root, ok := getRoot(ctx); ok {
			uri = root.ResolveReference(uri)
		}

		if fs, ok := getFS(uri); ok {
			return fs.Open(ctx, uri, options...)
		}
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	if err := applyOptions(f, options); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}

// List takes a Context and a filename (which may be a URL) and returns a list of os.FileInfo that describes the files contained in the directory or listing.
func List(ctx context.Context, filename string, options ...Option) ([]os.FileInfo, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		if err := applyOptions(os.Stdin, options); err != nil {
			return nil, err
		}
		return os.Stdin.Readdir(0)
	}

	if uri, err := url.Parse(filename); err == nil {
		if root, ok := getRoot(ctx); ok {
			uri = root.ResolveReference(uri)
		}

		if fs, ok := getFS(uri); ok {
			return fs.List(ctx, uri, options...)
		}
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := applyOptions(f, options); err != nil {
		return nil, err
	}

	return f.Readdir(0)
}
