// Package clipboard implements a scheme for "clipboard:" and "clip:".
package clipboard

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "clip", "clipboard")
}

type clipboard interface {
	os.FileInfo
	Read() ([]byte, error)
	Write([]byte) error
}

var clipboards = make(map[string]clipboard)

func getClip(name string) clipboard {
	if len(name) < 1 {
		// due to design of Go, nil must be explicitly passed here
		// otherwise it will be a type nil, which != nil.
		if defaultClipboard == nil {
			return nil
		}

		return defaultClipboard
	}

	clip, ok := clipboards[name]
	if !ok || clip == nil {
		return nil
	}
	return clip
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	clip := getClip(uri.Opaque)
	if clip == nil {
		return nil, os.ErrNotExist
	}

	b, err := clip.Read()
	if err != nil {
		return nil, err
	}

	return wrapper.NewReaderFromBytes(b, uri, time.Now()), nil
}

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	clip := getClip(uri.Opaque)
	if clip == nil {
		return nil, os.ErrNotExist
	}

	return wrapper.NewWriter(ctx, uri, func(b []byte) error {
		return clip.Write(b)
	}), nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	clip := getClip(uri.Opaque)
	if clip == nil {
		return nil, os.ErrNotExist
	}

	if !clip.IsDir() {
		return []os.FileInfo{clip}, nil
	}

	if len(clipboards) < 1 {
		return nil, os.ErrNotExist
	}

	var ret []os.FileInfo

	for _, clip := range clipboards {
		ret = append(ret, clip)
	}

	return ret, nil
}
