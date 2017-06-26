package http

import (
	"context"
	"encoding/json"
)

// WithJSON takes a value which is marshalled to JSON, and then attached to a Context, which is then used as the POST body in this library.
func WithJSON(ctx context.Context, v interface{}) (context.Context, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return WithContent(ctx, "application/json", b), nil
}

// WithJSONIndent takes a value which is marshalled to JSON with indentions, and then attached to a Context, which is then used as the POST body in this library. (Sometimes, some RPC servers do not like non-indented JSON.)
func WithJSONIndent(ctx context.Context, v interface{}) (context.Context, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}

	return WithContent(ctx, "application/json", b), nil
}
