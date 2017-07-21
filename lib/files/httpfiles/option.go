package httpfiles

import (
	"net/url"

	"github.com/puellanivis/breton/lib/files"
)

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
	type methodSetter interface {
		SetMethod(string) string
	}

	type ctypeSetter interface {
		SetContentType(string) string
	}

	type bodySetter interface {
		SetBody([]byte) []byte
	}

	return func(f files.File) (files.Option, error) {
		var methodSave, ctypeSave string
		var dataSave []byte

		if r, ok := f.(methodSetter); ok {
			methodSave = r.SetMethod(method)
		}

		if r, ok := f.(ctypeSetter); ok {
			ctypeSave = r.SetContentType(contentType)
		}

		if r, ok := f.(bodySetter); ok {
			dataSave = r.SetBody(data)
		}

		// option is not reversible
		return WithContent(methodSave, ctypeSave, dataSave), nil
	}
}

// WithMethod returns a files.Option that sets the Method of the
// underlying HTTP request to be the given value.
func WithMethod(method string) files.Option {
	type methodSetter interface {
		SetMethod(string) string
	}

	return func(f files.File) (files.Option, error) {
		var save string

		if r, ok := f.(methodSetter); ok {
			save = r.SetMethod(method)
		}

		return WithMethod(save), nil
	}
}

// WithContentType returns a files.Option that sets the Content-Type of the
// underlying HTTP request to be the given value. (This is intended to allow
// setting a specific type during a files.Create() and not have it auto-detect
// during the eventual commit of the request at Sync() or Close().)
func WithContentType(contentType string) files.Option {
	type ctypeSetter interface {
		SetContentType(string) string
	}

	return func(f files.File) (files.Option, error) {
		var save string

		if r, ok := f.(ctypeSetter); ok {
			save = r.SetContentType(contentType)
		}

		return WithContentType(save), nil
	}
}
