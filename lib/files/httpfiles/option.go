package httpfiles

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/puellanivis/breton/lib/files"
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

// WithForm returns a files.Option that that will add to the underlying HTTP
// request the url.Values given as a POST request. (A GET request can always
// be composed through the URL string itself.
func WithForm(vals url.Values) files.Option {
	body := []byte(vals.Encode())
	return WithContent("POST", "application/x-www-form-urlencoded", body)
}

// WithContent returns a files.Option that will set the Method, Body and
// Content-Type of the underlying HTTP request to the given values.
func WithContent(method, contentType string, data []byte) files.Option {
	return func(f files.File) (files.Option, error) {
		r, ok := f.(*request)
		if !ok {
			return nil, files.ErrNotSupported
		}

		methodSave := r.req.Method
		ctypeSave := r.req.Header.Get("Content-Type")
		dataSave := r.body

		r.req.Method = "POST"
		r.req.Header.Set("Content-Type", contentType)
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

// WithContentType returns a files.Option that sets the Content-Type of the
// underlying HTTP request to be the given value. (This is intended to allow
// setting a specific type during a files.Create() and not have it auto-detect
// during the eventual commit of the request at Sync() or Close().)
func WithContentType(contentType string) files.Option {
	return func(f files.File) (files.Option, error) {
		r, ok := f.(*request)
		if !ok {
			return nil, files.ErrNotSupported
		}

		save := r.req.Header.Get("Content-Type")

		r.req.Header.Set("Content-Type", contentType)

		return WithContentType(save), nil
	}
}
