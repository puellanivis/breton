package files

import (
	"context"
	"net/url"
	"os"
	"strconv"
)

type descriptorHandler struct{}

func init() {
	RegisterScheme(descriptorHandler{}, "fd")
}

func openFD(uri *url.URL, op string) (*os.File, error) {
	fd, err := strconv.ParseUint(filename(uri), 0, strconv.IntSize)
	if err != nil {
		return nil, &os.PathError{
			Op:   op,
			Path: uri.String(),
			Err:  err,
		}
		return nil, err
	}

	return os.NewFile(uintptr(fd), uri.String()), nil
}

func (descriptorHandler) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	return openFD(uri, "open")
}

func (descriptorHandler) Create(ctx context.Context, uri *url.URL) (Writer, error) {
	return openFD(uri, "create")
}

func (descriptorHandler) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	f, err := openFD(uri, "open")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}
