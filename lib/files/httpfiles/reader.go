package httpfiles

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type reader struct {
	r io.Reader
	info *wrapper.Info

	*request
	header http.Header

	err     error
	loading <-chan struct{}
}

func (r *reader) Header() (http.Header, error) {
	for range r.loading {
	}

	if r.err != nil {
		return nil, r.err
	}

	return r.header, nil
}

func (r *reader) Stat() (os.FileInfo, error) {
	for range r.loading {
	}

	if r.err != nil {
		return nil, r.err
	}

	return r.info.Stat()
}

func (r *reader) Read(b []byte) (n int, err error) {
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.r.Read(b)
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	// TODO: if we’ve not loaded yet, it’s possible to try and use http Range header? (header has poor support, so… blech for now.
	for range r.loading {
	}

	if r.err != nil {
		return 0, r.err
	}

	seeker, ok := r.r.(io.Seeker)
	if !ok {
		return 0, os.ErrInvalid
	}

	return seeker.Seek(offset, whence)
}

func (r *reader) Close() error {
	for range r.loading {
	}

	if r.err != nil {
		return r.err
	}

	if c, ok := r.r.(io.Closer); ok {
		return c.Close()
	}

	return nil
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	uri = elideDefaultPort(uri)

	cl, ok := getClient(ctx)
	if !ok {
		cl = http.DefaultClient
	}

	req := &http.Request{
		URL:    uri,
		Header: make(http.Header),
	}
	req = req.WithContext(ctx)

	if ua, ok := getUserAgent(ctx); ok {
		req.Header.Set("User-Agent", ua)
	}

	loading := make(chan struct{})
	r := &reader{
		loading: loading,

		request: &request{
			name: uri.String(),
			req:  req,
		},
	}

	go func() {
		// So, all of the file operations block on a range over the loading channel.
		// They will not end this blocking until loading is closed.
		// But they will also swallow any sends, though sends will block until someone is receiving.
		//
		// So, we will block on the first send until someone receives from the loading channel,
		// or the context expires.
		//
		// But none of the receivers will actually unblock until the loading channel is _closed_.
		// And once the channel is closed, each range over loading won’t even stop to block.

		defer close(loading)

		select {
		case loading <- struct{}{}:
		case <-ctx.Done():
			r.err = ctx.Err()
			return
		}

		// So, we will not arrive here until someone is ranging over the loading channel.
		//
		// This ensures the actual http request HAPPENS AFTER the first file operation is called,
		// but that all file operation behavior HAPPENS AFTER the actual http request is made.
		//
		// This lets us apply files.Option functions after files.Open,
		// and change the http.Request before actually doing it.

		resp, err := cl.Do(req)
		if err != nil {
			r.err = err
			return
		}

		r.header = resp.Header
		uri := resp.Request.URL

		t := time.Now()
		if lastmod := r.header.Get("Last-Modified"); lastmod != "" {
			if t1, err := http.ParseTime(lastmod); err == nil {
				t = t1
			}
		}

		r.info = wrapper.NewInfo(uri, int(resp.ContentLength), t)

		if resp.ContentLength < 0 {
			r.r = resp.Body
			return
		}

		b, err := files.ReadFrom(resp.Body)
		if err != nil {
			r.err = err
			return
		}

		if err := getErr(resp); err != nil {
			r.err = err
			return
		}

		r.r = bytes.NewReader(b)
	}()

	return r, nil
}
