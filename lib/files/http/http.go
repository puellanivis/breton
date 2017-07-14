package http

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
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

type Reader struct {
	sync.Mutex

	filling chan bool
	cancel func()
	err error

	req *http.Request
	r *wrapper.Reader
}

func (r *Reader) Name() string {
	return r.req.URL.String()
}

func (r *Reader) Stat() (os.FileInfo, error) {
	for <-r.filling {
	}

	if r.err != nil {
		return nil, r.err
	}

	return r.r.Stat()
}

func (r *Reader) Close() error {
	r.cancel()

	for <-r.filling {
	}

	if r.err != nil {
		return r.err
	}

	return r.r.Close()
}

func (r *Reader) Read(b []byte) (n int, err error) {
	for <-r.filling {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.r.Read(b)
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	for <-r.filling {
	}

	if r.err != nil {
		return 0, r.err
	}

	return r.r.Seek(offset, whence)
}

func WithQuery(r files.Reader, vals url.Values) error {
	body := []byte(vals.Encode())
	return WithContent(r, "application/x-www-form-urlencoded", body)
}

func WithContent(r files.Reader, ctype string, data []byte) error {
	rd, ok := r.(*Reader)
	if !ok {
		return errors.New("files.Reader is not an http.Reader")
	}

	rd.Lock()
	defer rd.Unlock()

	rd.req.Method = "POST"
	if ctype != "" {
		rd.req.Header.Add("Content-Type", ctype)
	}
	rd.req.ContentLength = int64(len(data))
	rd.req.Body = ioutil.NopCloser(bytes.NewReader(data))
	return nil
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	ctx, cancel := context.WithCancel(ctx)

	uri = elideDefaultPort(uri)

	req := &http.Request{
		URL:           uri,
		Header:        make(http.Header),
	}

	req = req.WithContext(ctx)

	r := &Reader{
		filling: make(chan bool),
		cancel: cancel,
		req: req,
	}

	go func() {
		defer close(r.filling)

		select {
		case r.filling <- true:
		case <-ctx.Done():
			r.err = ctx.Err()
			return
		}

		r.Lock()
		defer r.Unlock()

		cl, ok := getClient(ctx)
		if !ok {
			cl = http.DefaultClient
		}

		resp, err := cl.Do(req)
		if err != nil {
			r.err = err
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

		var t time.Time
		if lastmod := resp.Header.Get("Last-Modified"); lastmod != "" {
			if t1, err := http.ParseTime(lastmod); err == nil {
				t = t1
			}
		} else {
			t = time.Now()
		}

		r.r = wrapper.NewReader(uri, b, t)
	}()

	return r, nil
}

type Writer struct {
	sync.Mutex

	ctype string
	*wrapper.Writer
}

func WithContentType(w files.Writer, ctype string) error {
	wr, ok := w.(*Writer)
	if !ok {
		return errors.New("files.Writer is not an http.Writer")
	}

	wr.Lock()
	defer wr.Unlock()

	wr.ctype = ctype
	return nil
}

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	uri = elideDefaultPort(uri)

	addr := uri.String()

	cl, ok := getClient(ctx)
	if !ok {
		cl = http.DefaultClient
	}

	w := new(Writer)

	w.Writer = wrapper.NewWriter(ctx, uri, func(b []byte) error {
		w.Lock()
		defer w.Unlock()

		if w.ctype == "" {
			w.ctype = http.DetectContentType(b)
		}

		resp, err := cl.Post(addr, w.ctype, bytes.NewReader(b))
		if err != nil {
			return err
		}

		if err := files.Discard(resp.Body); err != nil {
			return err
		}

		return getErr(resp)
	})

	return w, nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
