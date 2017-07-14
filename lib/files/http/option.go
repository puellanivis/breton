package http

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"lib/files"
)

type request struct {
	// we tuck this in here, so that go says this is a files.File
	// so that it can be passed through files.Option functions, _but_
	// we never actually _define_ this value
	files.File

	body []byte
	// this is what we really care about
	req *http.Request
}

func WithForm(vals url.Values) files.Option {
	body := []byte(vals.Encode())
	return WithContent("POST", "application/x-www-form-urlencoded", body)
}

func WithContent(method, ctype string, data []byte) files.Option {
	return func(f files.File) (files.Option, error) {
		r, ok := f.(*request)
		if !ok {
			return nil, files.ErrNotSupported
		}

		methodSave := r.req.Method
		ctypeSave := r.req.Header.Get("Content-Type")
		dataSave := r.body

		r.req.Method = "POST"
		r.req.Header.Set("Content-Type", ctype)
		r.req.ContentLength = int64(len(data))

		r.body = data
		r.req.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader(data)), nil
		}

		// we know this http.Request.GetBody won’t throw an error
		r.req.Body, _ = r.req.GetBody()

		// option is not reversible
		return WithContent(methodSave, ctypeSave, dataSave), nil
	}
}

func WithContentType(ctype string) files.Option {
	return func(f files.File) (files.Option, error) {
		r, ok := f.(*request)
		if !ok {
			return nil, files.ErrNotSupported
		}

		save := r.req.Header.Get("Content-Type")

		r.req.Header.Set("Content-Type", ctype)

		return WithContentType(save), nil
	}
}
