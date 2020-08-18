package files

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
)

// Create returns a files.Writer, which can be used to write content to the resource at the given URL.
//
// If the given URL is a local filename, the file will be created, and truncated before this function returns.
//
// All errors and reversion functions returned by Option arguments are discarded.
func Create(ctx context.Context, url string, options ...Option) (Writer, error) {
	f, err := create(ctx, url)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		_, _ = opt(f)
	}

	return f, nil
}

func create(ctx context.Context, resource string) (Writer, error) {
	switch resource {
	case "", "-", "/dev/stdout":
		return os.Stdout, nil
	case "/dev/stderr":
		return os.Stderr, nil
	}

	if filepath.IsAbs(resource) {
		return os.Create(resource)
	}

	if uri, err := url.Parse(resource); err == nil {
		uri = resolveFilename(ctx, uri)

		if fs, ok := getFS(uri); ok {
			return fs.Create(ctx, uri)
		}
	}

	return os.Create(resource)
}
