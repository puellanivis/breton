package httpfiles

import (
	"context"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/json"
)

// REST is a convenience function, that marshals the send parameter into JSON, uses it as an attached content to read the uri, and then unmarshals the JSON received into the recv parameter.
func REST(ctx context.Context, uri string, send, recv interface{}) error {
	r, err := files.Open(ctx, uri, WithJSON(send))
	if err != nil {
		return err
	}

	return json.ReadFrom(r, recv)
}

// WithJSON takes a value which is marshalled to JSON, and then attached to a Context, which is then used as the POST body in this library.
func WithJSON(v interface{}, opts ...json.Option) files.Option {
	return func(f files.File) (files.Option, error) {
		b, err := json.Marshal(v, opts...)
		if err != nil {
			return nil, err
		}

		return WithContent("POST", "application/json", b)(f)
	}
}
