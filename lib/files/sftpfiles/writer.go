package sftpfiles

import (
	"context"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"

	"github.com/pkg/sftp"
)

type writer struct {
	uri *url.URL
	*Host

	loading <-chan struct{}
	f       *sftp.File
	err     error
}

func (w *writer) Name() string {
	return w.uri.String()
}

func (w *writer) Stat() (os.FileInfo, error) {
	for range w.loading {
	}

	if w.err != nil {
		return nil, w.err
	}

	return w.f.Stat()
}

func (w *writer) Write(b []byte) (n int, err error) {
	for range w.loading {
	}

	if w.err != nil {
		return 0, w.err
	}

	return w.f.Write(b)
}

func (w *writer) Seek(offset int64, whence int) (int64, error) {
	for range w.loading {
	}

	if w.err != nil {
		return 0, w.err
	}

	return w.f.Seek(offset, whence)
}

func (w *writer) Close() error {
	for range w.loading {
	}

	if w.err != nil {
		// This error is a connection error, and request-scoped.
		// So, in the context of Close, the error is irrelevant, so we ignore it.
		return nil
	}

	return w.f.Close()
}

func (fs *filesystem) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	h, u := fs.getHost(uri)

	if cl := h.GetClient(); cl != nil {
		f, err := cl.Create(uri.Path)
		if err != nil {
			return nil, &os.PathError{
				Op:   "create",
				Path: u.String(),
				Err:  err,
			}
		}

		return f, nil
	}

	loading := make(chan struct{})

	w := &writer{
		uri:  u,
		Host: h,

		loading: loading,
	}

	go func() {
		defer close(loading)

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			w.err = &os.PathError{
				Op:   "connect",
				Path: h.Name(),
				Err:  ctx.Err(),
			}
			return
		}

		cl, err := h.Connect()
		if err != nil {
			w.err = &os.PathError{
				Op:   "connect",
				Path: h.Name(),
				Err:  ctx.Err(),
			}
			return
		}

		f, err := cl.Create(uri.Path)
		if err != nil {
			w.err = &os.PathError{
				Op:   "create",
				Path: u.String(),
				Err:  err,
			}
			return
		}

		w.f = f
	}()

	return w, nil
}
