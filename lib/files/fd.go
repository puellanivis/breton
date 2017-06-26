package files

import (
	"context"
	"net/url"
	"os"
	"strconv"
)

type descriptorHandler struct{}

func init() {
	RegisterScheme(&descriptorHandler{}, "fd")
}

func filename(uri *url.URL) (uintptr, error) {
        filename := uri.Opaque
        if uri.Opaque == "" {
                filename = uri.String()

		if len(uri.Scheme)+3 < len(filename) {
                	filename = filename[len(uri.Scheme)+3:]
		}
        }

	u, err := strconv.ParseUint(filename, 0, 64)
	return uintptr(u), err
}

func (_ *descriptorHandler) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	fd, err := filename(uri)
	if err != nil {
		return nil, err
	}

	return os.NewFile(fd, uri.String()), nil
}

func (_ *descriptorHandler) Create(ctx context.Context, uri *url.URL) (Writer, error) { 
	fd, err := filename(uri)
	if err != nil {
		return nil, err
	}

	return os.NewFile(fd, uri.String()), nil
}

func (_ *descriptorHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	fd, err := filename(uri)
	if err != nil {
		return nil, err
	}

	f := os.NewFile(fd, uri.String())
	defer f.Close()

	return f.Readdir(0)
}
