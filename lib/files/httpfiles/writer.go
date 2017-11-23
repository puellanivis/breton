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

func (w *writer) Header() (http.Header, error) {
	return w.request.req.Header, nil
}

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	uri = elideDefaultPort(uri)

	cl, ok := getClient(ctx)
	if !ok {
		cl = http.DefaultClient
	}

	req := &http.Request{
		Method: "POST",
		URL:    uri,
		Header: make(http.Header),
	}
	req = req.WithContext(ctx)

	if ua, ok := getUserAgent(ctx); ok {
		req.Header.Set("User-Agent", ua)
	}

	r := &request{
		name: uri.String(),
		req:  req,
	}

	w := wrapper.NewWriter(ctx, uri, func(b []byte) error {
		if r.req.Header.Get("Content-Type") == "" {
			r.req.Header.Set("Content-Type", http.DetectContentType(b))
		}
		_ = r.SetBody(b)

		resp, err := cl.Do(r.req)
		if err != nil {
			return err
		}

		if err := files.Discard(resp.Body); err != nil {
			return err
		}

		return getErr(resp)
	})

	return &writer{
		request: r,
		Writer:  w,
	}, nil
}
