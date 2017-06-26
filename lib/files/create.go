package files

import (
	"context"
	"net/url"
	"os"
)

func Create(ctx context.Context, filename string) (Writer, error) {
	switch filename {
	case "", "-", "/dev/stdout":
		return os.Stdout, nil
	case "/dev/stderr":
		return os.Stderr, nil
	}
 
	if uri, err := url.Parse(filename); err == nil {
		if fs, ok := getFS(uri); ok {
			return fs.Create(ctx, uri)
		}
	}

	return os.Create(filename)
}
