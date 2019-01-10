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
	h := fs.getHost(uri)

	if cl := h.GetClient(); cl != nil {
		f, err := cl.Open(uri.Path)
		if err != nil {
			return nil, files.PathError("open", uri.String(), err)
		}

		return f, nil
	}

	loading := make(chan struct{})

	fixURL := *uri
	fixURL.Host = h.uri.Host
	fixURL.User = h.uri.User

	r := &reader{
		uri:  &fixURL,
		Host: h,

		loading: loading,
	}

	go func() {
		defer close(loading)

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			r.err = files.PathError("connect", h.Name(), ctx.Err())
			return
		}

		cl, err := h.Connect()
		if err != nil {
			r.err = files.PathError("connect", h.Name(), err)
			return
		}

		f, err := cl.Open(uri.Path)
		if err != nil {
			r.err = files.PathError("open", r.Name(), err)
			return
		}

		r.f = f
	}()

	return r, nil
}
