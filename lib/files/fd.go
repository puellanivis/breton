package files

import (
	"context"
	"net/url"
	"os"
	"strings"
)

type descriptorHandler struct{}

func init() {
	RegisterScheme(descriptorHandler{}, "fd")
}

func openFD(uri *url.URL) (*os.File, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, ErrURLCannotHaveAuthority
	}

	num := strings.TrimPrefix(uri.Path, "/")
	if num == "" {
		var err error
		num, err = url.PathUnescape(uri.Opaque)
		if err != nil {
			return nil, ErrURLInvalid
		}
	}

	fd, err := resolveFileHandle(num)
	if err != nil {
		return nil, err
	}

	// Canonicalize the name.
	uri = &url.URL{
		Scheme: "fd",
		Opaque: url.PathEscape(num),
	}

	return os.NewFile(uintptr(fd), uri.String()), nil
}

func (descriptorHandler) Open(ctx context.Context, uri *url.URL) (Reader, error) {
	f, err := openFD(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	return f, nil
}

func (descriptorHandler) Create(ctx context.Context, uri *url.URL) (Writer, error) {
	f, err := openFD(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	return f, nil
}

func (descriptorHandler) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	f, err := openFD(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  err,
		}
	}
	defer f.Close()

	return f.Readdir(0)
}
