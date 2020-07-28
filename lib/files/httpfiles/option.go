package httpfiles

import (
	"net/url"

	"github.com/puellanivis/breton/lib/files"
)

// WithForm returns a files.Option that that will add to the underlying HTTP
// request the url.Values given as a POST request. (A GET request can always
// be composed through the URL string itself.
func WithForm(form url.Values) files.Option {
	body := []byte(form.Encode())
	return WithContent("POST", "application/x-www-form-urlencoded", body)
}

// WithContent returns a files.Option that will set the Method, Body and
// Content-Type of the underlying HTTP request to the given values.
func WithContent(method, contentType string, data []byte) files.Option {
	data = append([]byte(nil), data...)

	type contentSetter interface {
		SetMethod(string) (string, error)
		SetContentType(string) (string, error)
		SetBody([]byte) ([]byte, error)
	}

	return func(f files.File) (files.Option, error) {
		if f, ok := f.(contentSetter); ok {
			methodSave, err := f.SetMethod(method)
			if err != nil {
				return nil, err
			}

			ctypeSave, err := f.SetContentType(contentType)
			if err != nil {
				return nil, err
			}

			dataSave, err := f.SetBody(data)
			if err != nil {
				return nil, err
			}

			return WithContent(methodSave, ctypeSave, dataSave), nil
		}

		return nil, files.ErrNotSupported
	}
}

// WithMethod returns a files.Option that sets the Method of the
// underlying HTTP request to be the given value.
func WithMethod(method string) files.Option {
	type methodSetter interface {
		SetMethod(string) (string, error)
	}

	return func(f files.File) (files.Option, error) {
		if f, ok := f.(methodSetter); ok {
			save, err := f.SetMethod(method)
			if err != nil {
				return nil, err
			}

			return WithMethod(save), nil
		}

		return nil, files.ErrNotSupported
	}
}

// WithContentType returns a files.Option that sets the Content-Type of the
// underlying HTTP request to be the given value. (This is intended to allow
// setting a specific type during a files.Create() and not have it auto-detect
// during the eventual commit of the request at Sync() or Close().)
func WithContentType(contentType string) files.Option {
	type ctypeSetter interface {
		SetContentType(string) (string, error)
	}

	return func(f files.File) (files.Option, error) {
		if f, ok := f.(ctypeSetter); ok {
			save, err := f.SetContentType(contentType)
			if err != nil {
				return nil, err
			}

			return WithContentType(save), nil
		}

		return nil, files.ErrNotSupported
	}
}
