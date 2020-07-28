package sftpfiles

import (
	"context"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"

	"github.com/pkg/sftp"
)

type reader struct {
	uri *url.URL
	*Host

	loading <-chan struct{}
	f       *sftp.File
	err     error
}

func (r *reader) Name() string {
	return r.uri.String()
}

func (r *reader) Stat() (os.FileInfo, error) {
	for range r.loading {
	}

	if r.err != nil {
		return nil, r.err
	}

	return r.f.Stat()
}

func (r *reader) Read(b []byte) (n int, err error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.f.Read(b)
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.f.Seek(offset, whence)
}

func (r *reader) Close() error {
	for range r.loading {
	}

	if r.err != nil {
		// This error is a connection error, and request-scoped.
		// So, in the context of Close, the error is irrelevant, so we ignore it.
		return nil
	}

	return r.f.Close()
}

func (fs *filesystem) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	h, u := fs.getHost(uri)

	if cl := h.GetClient(); cl != nil {
		f, err := cl.Open(u.Path)
		if err != nil {
			return nil, &os.PathError{
				Op:   "open",
				Path: u.String(),
				Err:  err,
			}
		}

		return f, nil
	}

	loading := make(chan struct{})

	r := &reader{
		uri:  u,
		Host: h,

		loading: loading,
	}

	go func() {
		defer close(loading)

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			r.err = &os.PathError{
				Op:   "connect",
				Path: h.Name(),
				Err:  ctx.Err(),
			}
			return
		}

		cl, err := h.Connect()
		if err != nil {
			r.err = &os.PathError{
				Op:   "connect",
				Path: h.Name(),
				Err:  err,
			}
			return
		}

		f, err := cl.Open(u.Path)
		if err != nil {
			r.err = &os.PathError{
				Op:   "open",
				Path: u.String(),
				Err:  err,
			}
			return
		}

		r.f = f
	}()

	return r, nil
}
