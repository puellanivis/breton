package files

import (
	"encoding/json"
	"io"

	"context"
)

func ReadJSONFrom(r io.ReadCloser, v interface{}) error {
	data, err := ReadAndClose(r)
	if err != nil {
		return err
	}

	if len(data) < 1 {
		v = nil
		return nil
	}

	return json.Unmarshal(data, v)
}

func ReadJSON(ctx context.Context, filename string, v interface{}) error {
	f, err := Open(ctx, filename)
	if err != nil {
		return err
	}

	return ReadJSONFrom(f, v)
}

func WriteJSONIndentTo(w io.WriteCloser, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return WriteAndClose(w, b)
}

func WriteJSONTo(w io.WriteCloser, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return WriteAndClose(w, b)
}

func WriteJSONIndent(ctx context.Context, filename string, v interface{}) error {
	f, err := Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteJSONIndentTo(f, v)
}

func WriteJSON(ctx context.Context, filename string, v interface{}) error {
	f, err := Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteJSONTo(f, v)
}
