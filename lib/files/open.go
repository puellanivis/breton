package files

import (
	"context"
	"io/ioutil"
	"os"
)

// Open opens the file at the given filename, and
// returns a files.Reader, which will read the contents of that filename.
//
// All errors and reversion functions returned by Option arguments are discarded.
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

	uri := parsePath(ctx, filename)
	if isPath(uri) {
		return os.Open(filename)
	}

	if fs, ok := getFS(uri); ok {
		return fs.Open(ctx, uri)
	}

	return nil, ErrNotSupported
}

// ReadDir reads the directory at the given filename, and
// returns a slice of os.FileInfo, which describe the files contained in the directory.
func ReadDir(ctx context.Context, filename string) ([]os.FileInfo, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		return os.Stdin.Readdir(0)
	}

	uri := parsePath(ctx, filename)
	if isPath(uri) {
		return ioutil.ReadDir(filename)
	}

	if fs, ok := getFS(uri); ok {
		return fs.List(ctx, uri)
	}

	return nil, ErrNotSupported
}

// List reads the directory at the given filename, and
// returns a slice of os.FileInfo, which describe the files contained in the directory.
//
// Depcrecated: Use `ReadDir`.
func List(ctx context.Context, filename string) ([]os.FileInfo, error) {
	return ReadDir(ctx, filename)
}
