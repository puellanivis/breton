package http

import (
	"encoding/json"

	"lib/files"
)

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
