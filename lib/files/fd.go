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

func (h *descriptorHandler) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	fd, err := strconv.ParseUint(filename(uri), 0, 64)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), uri.String()), nil
}

func (h *descriptorHandler) Create(ctx context.Context, uri *url.URL) (Writer, error) {
	fd, err := strconv.ParseUint(filename(uri), 0, 64)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), uri.String()), nil
}

func (h *descriptorHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	fd, err := strconv.ParseUint(filename(uri), 0, 64)
	if err != nil {
		return nil, err
	}

	f, err := os.NewFile(uintptr(fd), uri.String()), nil
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}
