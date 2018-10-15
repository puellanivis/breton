package sftpfiles

import (
	"context"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"

	"github.com/pkg/sftp"
)

type writer struct {
	name string
	*Host

	loading <-chan struct{}
	f       *sftp.File
	err     error
}

func (r *writer) Name() string {
	return r.name
}

func (r *writer) Stat() (os.FileInfo, error) {
	for range r.loading {
	}

	if r.err != nil {
		return nil, r.err
	}

	return r.f.Stat()
}

func (r *writer) Write(b []byte) (n int, err error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.f.Write(b)
}

func (r *writer) Seek(offset int64, whence int) (int64, error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.f.Seek(offset, whence)
}

func (r *writer) Sync() error {
	for range r.loading {
	}

	return nil
}

func (r *writer) Close() error {
	for range r.loading {
	}

	if r.f == nil {
		return nil
	}

	return r.f.Close()
}

type noopSync struct {
	*sftp.File
}

func (f noopSync) Sync() error {
	return nil
}

func (fs *filesystem) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	h := fs.getHost(uri)

	if cl := h.GetClient(); cl != nil {
		f, err := cl.Create(uri.Path)
		if err != nil {
			return nil, &os.PathError{"create", uri.String(), err}
		}

		return noopSync{f}, nil
	}

	loading := make(chan struct{})

	r := &writer{
		name: uri.String(),
		Host: h,

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

		cl, err := h.ConnectClient()
		if err != nil {
			r.err = &os.PathError{"connect", uri.Host, err}
			return
		}

		f, err := cl.Create(uri.Path)
		if err != nil {
			r.err = &os.PathError{"create", uri.String(), err}
			return
		}

		r.f = f
	}()

	return r, nil
}
