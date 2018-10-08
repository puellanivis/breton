package sftpfiles

import (
	"context"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"

	"github.com/pkg/sftp"
)

type reader struct {
	name string
	*host

	loading <-chan struct{}
	f       *sftp.File
	err     error
}

func (r *reader) Name() string {
	return r.name
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

	if r.f == nil {
		return nil
	}

	return r.f.Close()
}

func (fs *filesystem) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	h := fs.getHost(uri)
	loading := make(chan struct{})

	r := &reader{
		name: uri.String(),
		host: h,

		loading: loading,
	}

	go func() {
		defer close(loading)

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			r.err = &os.PathError{"connect", uri.Host, ctx.Err()}
			return
		}

		cl, err := h.GetClient()
		if err != nil {
			r.err = &os.PathError{"connect", uri.Host, err}
			return
		}

		f, err := cl.Open(uri.Path)
		if err != nil {
			r.err = &os.PathError{"open", uri.String(), err}
		}

		r.f = f
	}()

	return r, nil
}