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

func (handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	clip, err := getClip(uri)
	if err != nil {
		return nil, files.PathError("open", uri.String(), err)
	}

	b, err := clip.Read()
	if err != nil {
		return nil, err
	}

	return wrapper.NewReaderFromBytes(b, uri, time.Now()), nil
}

func (handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	clip, err := getClip(uri)
	if err != nil {
		return nil, files.PathError("create", uri.String(), err)
	}

	return wrapper.NewWriter(ctx, uri, func(b []byte) error {
		return clip.Write(b)
	}), nil
}

func (handler) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  os.ErrInvalid,
		}
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	clip := clipboards[path]
	if clip == nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  os.ErrNotExist,
		}
	}

	if path != "" {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  files.ErrNotDirectory,
		}
	}

	var ret []os.FileInfo

	for _, clip := range clipboards {
		if info, err := clip.Stat(); err == nil {
			if fi, ok := info.(interface{ URL() *url.URL }); ok {
				u := fi.URL()
				u.Scheme = ""

				switch fi := fi.(type) {
				case interface{ SetNameFromURL(*url.URL) }:
					fi.SetNameFromURL(u)

				case interface{ SetName(string) }:
					fi.SetName(u.String())

				default:
					info = wrapper.NewInfo(u, int(info.Size()), info.ModTime())
				}
			}

			ret = append(ret, info)
		}
	}

	return ret, nil
}
