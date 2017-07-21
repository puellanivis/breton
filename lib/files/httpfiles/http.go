package httpfiles

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type handler struct{}

var schemes = map[string]string{
	"http":  "80",
	"https": "443",
}

func init() {
	var schemeList []string

	for scheme := range schemes {
		schemeList = append(schemeList, scheme)
	}

	files.RegisterScheme(&handler{}, schemeList...)
}

func elideDefaultPort(uri *url.URL) *url.URL {
	port := uri.Port()

	/* elide default ports  */
	if defport, ok := schemes[uri.Scheme]; ok && defport == port {
		newuri := *uri
		newuri.Host = uri.Hostname()
		return &newuri
	}

	return uri
}

func getErr(resp *http.Response) error {
	switch resp.StatusCode {
	case 200, 204:
		return nil
	case 401, 403:
		return os.ErrPermission
	case 404:
		return os.ErrNotExist
	}

	return errors.New(resp.Status)
}

func (h *handler) Open(ctx context.Context, uri *url.URL, options ...files.Option) (files.Reader, error) {
	uri = elideDefaultPort(uri)

	req := &http.Request{
		URL:    uri,
		Header: make(http.Header),
	}

	req = req.WithContext(ctx)

	r := &request{
		req: req,
	}

	for _, opt := range options {
		if _, err := opt(r); err != nil {
			return nil, err
		}
	}

	if ua, ok := getUserAgent(ctx); ok {
		r.req.Header.Set("User-Agent", ua)
	}

	cl, ok := getClient(ctx)
	if !ok {
		cl = http.DefaultClient
	}

	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := files.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := getErr(resp); err != nil {
		return nil, err
	}

	var t time.Time
	if lastmod := resp.Header.Get("Last-Modified"); lastmod != "" {
		if t1, err := http.ParseTime(lastmod); err == nil {
			t = t1
		}
	} else {
		t = time.Now()
	}

	return wrapper.NewReader(uri, b, t), nil
}

func (h *handler) Create(ctx context.Context, uri *url.URL, options ...files.Option) (files.Writer, error) {
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

	r := &request{
		req: req,
	}

	for _, opt := range options {
		if _, err := opt(r); err != nil {
			return nil, err
		}
	}

	return wrapper.NewWriter(ctx, uri, func(b []byte) error {
		if ua, ok := getUserAgent(ctx); ok {
			r.req.Header.Set("User-Agent", ua)
		}

		if r.req.Header.Get("Content-Type") == "" {
			r.req.Header.Set("Content-Type", http.DetectContentType(b))
		}

		r.req.Body = ioutil.NopCloser(bytes.NewReader(b))

		resp, err := cl.Do(r.req)
		if err != nil {
			return err
		}

		if err := files.Discard(resp.Body); err != nil {
			return err
		}

		return getErr(resp)
	}), nil
}

func (h *handler) List(ctx context.Context, uri *url.URL, options ...files.Option) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
