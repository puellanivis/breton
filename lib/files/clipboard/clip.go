// Package clipboard implements a scheme for "clipboard:" and "clip:".
package clipboard

import (
	"context"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "clip", "clipboard")
}

type clipboard interface {
	Stat() (os.FileInfo, error)
	Read() ([]byte, error)
	Write([]byte) error
}

var clipboards = make(map[string]clipboard)

func getClip(uri *url.URL) (clipboard, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, os.ErrInvalid
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	clip := clipboards[path]
	if clip == nil {
		return nil, os.ErrNotExist
	}

	return clip, nil
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	clip, err := getClip(uri)
	if err != nil {
		return nil, &os.PathError{"open", uri.String(), err}
	}

	b, err := clip.Read()
	if err != nil {
		return nil, err
	}

	return wrapper.NewReaderFromBytes(b, uri, time.Now()), nil
}

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	clip, err := getClip(uri)
	if err != nil {
		return nil, &os.PathError{"create", uri.String(), err}
	}

	return wrapper.NewWriter(ctx, uri, func(b []byte) error {
		return clip.Write(b)
	}), nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	clip := clipboards[path]
	if clip == nil {
		return nil, &os.PathError{"readdir", uri.String(), os.ErrNotExist}
	}

	if path != "" {
		return nil, &os.PathError{"readdir", uri.String(), syscall.ENOTDIR}
	}

	if len(clipboards) < 1 {
		return nil, &os.PathError{"readdir", uri.String(), os.ErrNotExist}
	}

	var ret []os.FileInfo

	for _, clip := range clipboards {
		if fi, err := clip.Stat(); err == nil {
			ret = append(ret, fi)
		}
	}

	return ret, nil
}
