package httpfiles

import (
	"context"
	"net/http"
	"net/url"

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

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	uri = elideDefaultPort(uri)

	cl, ok := getClient(ctx)
	if !ok {
		cl = http.DefaultClient
	}

	req := newHTTPRequest(http.MethodPost, uri)
	req = req.WithContext(ctx)

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
		if r.req.Header.Get("Content-Type") == "" {
			r.req.Header.Set("Content-Type", http.DetectContentType(b))
		}
		_ = r.SetBody(b)

		resp, err := cl.Do(r.req)
		if err != nil {
			return files.PathError("write", r.name, err)
		}

		if err := files.Discard(resp.Body); err != nil {
			return err
		}

		if err := getErr(resp); err != nil {
			return files.PathError("write", r.name, err)
		}

		return nil
	})

	return &writer{
		request: r,
		Writer:  w,
	}, nil
}
