package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"lib/files"
	"lib/files/wrapper"
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

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	uri = elideDefaultPort(uri)

	var method, ctype string
	var body io.ReadCloser
	var size int64

	if b, ok := getContent(ctx); ok {
		method = "POST"
		size = int64(len(b))
		body = ioutil.NopCloser(bytes.NewReader(b))

		if ctype, ok = getContentType(ctx); !ok {
			ctype = http.DetectContentType(b)
		}
	}

	req := &http.Request{
		Method:        method,
		URL:           uri,
		Header:        make(http.Header),
		Body:          body,
		ContentLength: size,
	}

	req = req.WithContext(ctx)

	if ctype != "" {
		req.Header.Add("Content-Type", ctype)
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

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	uri = elideDefaultPort(uri)

	addr := uri.String()

	cl, ok := getClient(ctx)
	if !ok {
		cl = http.DefaultClient
	}

	ctype, ok := getContentType(ctx)
	if !ok {
		ctype = "application/octet-stream"
	}

	return wrapper.NewWriter(ctx, uri, func(b []byte) error {
		resp, err := cl.Post(addr, ctype, bytes.NewReader(b))
		if err != nil {
			return err
		}

		if err := files.Discard(resp.Body); err != nil {
			return err
		}

		return getErr(resp)
	}), nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
