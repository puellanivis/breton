package httpfiles

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/url"
	"sync"
)

// ErrRequestAlreadySent is returned, if you attempt to modify a Request field
// after the request has already been sent.
var ErrRequestAlreadySent = errors.New("request already sent")

func httpNewRequestWithContext(ctx context.Context, method string, uri *url.URL) *http.Request {
	r := &http.Request{
		Method:     method,
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}

	return r.WithContext(ctx)
}

type request struct {
	mu sync.RWMutex

	sent bool // true if request has already been sent.

	setContentType bool // true if Content-Type has been set to a specific value.

	name string

	// this is what we really care about
	body []byte
	req  *http.Request
}

func (r *request) markSent() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sent = true
}

func (r *request) Name() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.name
}

func (r *request) SetName(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.name = name
}

func (r *request) SetMethod(method string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	save := r.req.Method

	if r.sent {
		return save, ErrRequestAlreadySent
	}

	r.req.Method = method

	return save, nil
}

func (r *request) SetHeader(key string, values ...string) ([]string, error) {
	key = textproto.CanonicalMIMEHeaderKey(key)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.req.Header == nil {
		// Safety check.
		r.req.Header = make(http.Header)
	}

	save := r.req.Header[key]

	if r.sent {
		return append([]string(nil), save...), ErrRequestAlreadySent
	}

	if key == "Content-Type" {
		r.setContentType = len(values) > 0
	}

	r.req.Header[key] = append([]string(nil), values...)

	return save, nil
}

func (r *request) AddHeader(key string, values ...string) ([]string, error) {
	key = textproto.CanonicalMIMEHeaderKey(key)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.req.Header == nil {
		// Safety check.
		r.req.Header = make(http.Header)
	}

	save := r.req.Header[key]

	switch {
	case r.sent:
		return append([]string(nil), save...), ErrRequestAlreadySent

	case len(values) < 1:
		return append([]string(nil), save...), nil
	}

	cur := save[:len(save):len(save)] // truncate capacity, so the append below clones save.
	r.req.Header[key] = append(cur, values...)

	// Go ahead and return `save` with full allocated capacity.
	return save, nil
}

func (r *request) SetContentType(contentType string) (string, error) {
	prev, err := r.SetHeader("Content-Type", contentType)
	if len(prev) > 0 {
		return prev[0], err
	}
	return "", err
}

func (r *request) SetBody(body []byte) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	save := r.body

	if r.sent {
		// Since we are not changing `r.body` at all,
		// multiple calls would all return the same backing store.
		// So, we clone the body we’re returning just to be safe.
		return append([]byte(nil), save...), ErrRequestAlreadySent
	}

	// To ensure we have an exclusive copy of the backing store,
	// we clone the input body as a safety measure.
	// Otherwise, a caller could mutate the backing store behind our back.
	body = append([]byte(nil), body...)
	r.body = body // `save` is now the exclusive reference to the previous backing store.

	r.req.ContentLength = int64(len(r.body))

	if !r.setContentType {
		r.req.Header.Set("Content-Type", http.DetectContentType(body))
	}

	switch {
	case len(body) < 1:
		r.req.GetBody = func() (io.ReadCloser, error) {
			return http.NoBody, nil
		}

	default:
		r.req.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader(body)), nil
		}
	}

	// we know this http.Request.GetBody can’t throw an error
	r.req.Body, _ = r.req.GetBody()

	// Since save is the exclusive reference to the previous backing store,
	// there is no need for a copy here.
	return save, nil
}
