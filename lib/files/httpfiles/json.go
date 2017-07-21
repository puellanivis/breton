package httpfiles

import (
	"context"
	"encoding/json"

	"github.com/puellanivis/breton/lib/files"
)

// REST is a convenience function, that marshals the send parameter into JSON, uses it as an attached content to read the uri, and then unmarshals the JSON received into the recv parameter.
func REST(ctx context.Context, uri string, send, recv interface{}) error {
	r, err := files.Open(ctx, uri, WithJSON(send))
	if err != nil {
		return err
	}

	return files.ReadJSONFrom(r, recv)
}

// WithJSON takes a value which is marshalled to JSON, and then attached to a Context, which is then used as the POST body in this library.
func WithJSON(v interface{}) files.Option {
	return func(f files.File) (files.Option, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		return WithContent("POST", "application/json", b)(f)
	}
}

// WithJSONIndent takes a value which is marshalled to JSON with indentions, and then attached to a Context, which is then used as the POST body in this library. (Sometimes, some RPC servers do not like non-indented JSON.)
func WithJSONIndent(v interface{}) files.Option {
	return func(f files.File) (files.Option, error) {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return nil, err
		}

		return WithContent("POST", "application/json", b)(f)
	}
}
