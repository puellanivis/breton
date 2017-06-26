package http

import (
	"context"
	"encoding/json"
)

func WithJSON(ctx context.Context, v interface{}) (context.Context, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return WithContent(ctx, "application/json", b), nil
}

func WithJSONIndent(ctx context.Context, v interface{}) (context.Context, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}

	return WithContent(ctx, "application/json", b), nil
}
