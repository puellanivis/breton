package http

import (
	"encoding/json"

	"lib/files"
)

// WithJSON takes a value which is marshalled to JSON, and then attached to a Context, which is then used as the POST body in this library.
func WithJSON(r files.Reader, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return WithContent(r, "application/json", b)
}

// WithJSONIndent takes a value which is marshalled to JSON with indentions, and then attached to a Context, which is then used as the POST body in this library. (Sometimes, some RPC servers do not like non-indented JSON.)
func WithJSONIndent(r files.Reader, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return WithContent(r, "application/json", b)
}
