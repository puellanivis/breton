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

func (h *descriptorHandler) open(uri *url.URL) (*os.File, error) {
	fd, err := strconv.ParseUint(filename(uri), 0, 64)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), uri.String()), nil
}

func (h *descriptorHandler) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	return h.open(uri)
}

func (h *descriptorHandler) Create(ctx context.Context, uri *url.URL) (Writer, error) {
	return h.open(uri)
}

func (h *descriptorHandler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	f, err := h.open(uri)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}
