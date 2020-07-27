package httpfiles

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type writer struct {
	*wrapper.Writer
	*request
}

func (w *writer) Name() string {
	return w.request.Name()
}

func (w *writer) Header() (http.Header, error) {
	return w.request.req.Header, nil
}

func (handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	uri = elideDefaultPort(uri)

	cl, ok := getClient(ctx)
	if !ok {
		cl = http.DefaultClient
	}

	req := httpNewRequestWithContext(ctx, http.MethodPost, uri)

	if ua, ok := getUserAgent(ctx); ok {
		req.Header.Set("User-Agent", ua)
	}

	r := &request{
		name: uri.String(),
		req:  req,
	}

	// The http.Writer does not actually perform the http.Request until wrapper.Sync is called,
	// So there is no need for complex synchronization like the httpfiles.Reader needs.
	w := wrapper.NewWriter(ctx, uri, func(b []byte) error {
		_, err := r.SetBody(b)
		if err != nil {
			return err
		}

		r.mu.Lock()
		defer r.mu.Unlock()

		// Unlike on the reader side, we never want to call r.markSent()
		// Because, we perform a brand new request, every Sync() or Close().
		// So, we can continuously update headers, and bodies, and methods.

		resp, err := cl.Do(r.req)
		if err != nil {
			return &os.PathError{
				Op:   "write",
				Path: r.name,
				Err:  err,
			}
		}

		_ = files.Discard(resp.Body)

		if err := getErr(resp); err != nil {
			return &os.PathError{
				Op:   "write",
				Path: r.name,
				Err:  err,
			}
		}

		return nil
	})

	return &writer{
		request: r,
		Writer:  w,
	}, nil
}
