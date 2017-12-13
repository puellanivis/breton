package httpfiles

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

type request struct {
	name string

	// this is what we really care about
	body []byte
	req  *http.Request
}

func (r *request) Name() string {
	return r.name
}

func (r *request) SetMethod(method string) string {
	save := r.req.Method
	r.req.Method = method

	return save
}

func (r *request) SetContentType(contentType string) string {
	if r.req.Header == nil {
		r.req.Header = make(http.Header)
	}

	save := r.req.Header.Get("Content-Type")
	r.req.Header.Set("Content-Type", contentType)

	return save
}

func (r *request) SetBody(body []byte) []byte {
	save := r.body

	r.req.ContentLength = int64(len(body))
	r.body = body

	r.req.GetBody = func() (io.ReadCloser, error) {
		if len(r.body) < 1 {
			return nil, nil
		}

		return ioutil.NopCloser(bytes.NewReader(r.body)), nil
	}

	// we know this http.Request.GetBody won’t throw an error
	r.req.Body, _ = r.req.GetBody()

	return save
}
