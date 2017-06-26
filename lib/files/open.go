package files

import (
	"context"
	"net/url"
	"os"
)

// Open takes a Context and a filename (which may be a URL) and returns a files.Reader which will read the contents of that filename or URL.
func Open(ctx context.Context, filename string) (Reader, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		return os.Stdin, nil
	}

	if uri, err := url.Parse(filename); err == nil {
		if fs, ok := getFS(uri); ok {
			return fs.Open(ctx, uri)
		}
	}

	return os.Open(filename)
}

// List takes a Context and a filename (which may be a URL) and returns a list of os.FileInfo that describes the files contained in the directory or listing.
func List(ctx context.Context, filename string) ([]os.FileInfo, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		return os.Stdin.Readdir(0)
	}

	if uri, err := url.Parse(filename); err == nil {
		if fs, ok := getFS(uri); ok {
			return fs.List(ctx, uri)
		}
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}
