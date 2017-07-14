package files

import (
	"context"
	"net/url"
	"os"
)

// Create takes a context and a filename (which may be a URL) and returns a files.Writer that allows writing data to that local filename or URL.
func Create(ctx context.Context, filename string, options ...Option) (Writer, error) {
	switch filename {
	case "", "-", "/dev/stdout":
		if err := applyOptions(os.Stdout, options); err != nil {
			return nil, err
		}
		return os.Stdout, nil
	case "/dev/stderr":
		if err := applyOptions(os.Stderr, options); err != nil {
			return nil, err
		}
		return os.Stderr, nil
	}

	if uri, err := url.Parse(filename); err == nil {
		if root, ok := getRoot(ctx); ok {
			uri = root.ResolveReference(uri)
		}

		if fs, ok := getFS(uri); ok {
			return fs.Create(ctx, uri, options...)
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	if err := applyOptions(f, options); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}
