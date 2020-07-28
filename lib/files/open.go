package files

import (
	"context"
	"os"
)

// Open takes a context and a filename (which may be a URL) and
// returns a `files.Reader`, which will read the contents of that filename or URL.
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
		return os.Open(uri.Path)
	}

	fsys, ok := getFS(uri)
	if !ok {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  ErrNotSupported,
		}
	}

	return fsys.Open(ctx, uri)
}
