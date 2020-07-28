package files

import (
	"context"
	"net/url"
	"os"
)

// CreateFS defines an extention interface on FS, which also provides an ability to create a new file for read/write.
type CreateFS interface {
	FS
	Create(ctx context.Context, uri *url.URL) (Writer, error)
}

// Create takes a context and a filename (which may be a URL) and
// returns a files.Writer that allows writing data to that local filename or URL.
//
// All errors and reversion functions returned by Option arguments are discarded.
func Create(ctx context.Context, filename string, options ...Option) (Writer, error) {
	f, err := create(ctx, filename)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		_, _ = opt(f)
	}

	return f, nil
}

func create(ctx context.Context, filename string) (Writer, error) {
	switch filename {
	case "", "-", "/dev/stdout":
		return os.Stdout, nil
	case "/dev/stderr":
		return os.Stderr, nil
	}

	uri := parsePath(ctx, filename)
	if isPath(uri) {
		return os.Create(filename)
	}

	fsys, ok := getFS(uri)
	if !ok {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  ErrNotSupported,
		}
	}

	switch fsys := fsys.(type) {
	case CreateFS:
		return fsys.Create(ctx, uri)

		// case OpenFileFS: // implement
	}

	return nil, &os.PathError{
		Op:   "create",
		Path: filename,
		Err:  ErrNotSupported,
	}
}
