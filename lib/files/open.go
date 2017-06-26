package files

import (
	"context"	
	"net/url"
	"os"
)

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

func List(ctx context.Context, filename string) ([]os.FileInfo, error) {
	switch filename {
	case "", "-", "/dev/stdin":
		return os.Stdin.Readdir(0)
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}
