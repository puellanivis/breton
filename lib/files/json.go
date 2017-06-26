package files

import (
	"encoding/json"
	"io"

	"context"
)

// ReadJSONFrom will ReadAndClose the given io.ReadCloser and unmarshal that data into v as per json.Unmarshal.
func ReadJSONFrom(r io.ReadCloser, v interface{}) error {
	data, err := ReadFrom(r)
	if err != nil {
		return err
	}

	if len(data) < 1 {
		v = nil
		return nil
	}

	return json.Unmarshal(data, v)
}

// ReadJSON will open a filename with the given context, and unmarshal that data into v as per json.Unmarshal.
func ReadJSON(ctx context.Context, filename string, v interface{}) error {
	f, err := Open(ctx, filename)
	if err != nil {
		return err
	}

	return ReadJSONFrom(f, v)
}

// WriteJSONIndentTo writes a value marshalled with indentions as JSON to the given io.WriteCloser.
func WriteJSONIndentTo(w io.WriteCloser, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return WriteTo(w, b)
}

// WriteJSONTo writes a value marshalled as JSON to the the given io.WriteCloser.
func WriteJSONTo(w io.WriteCloser, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return WriteTo(w, b)
}

// WriteJSONIndent writes a marshaled JSON with indention to a filename with the given Context.
func WriteJSONIndent(ctx context.Context, filename string, v interface{}) error {
	f, err := Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteJSONIndentTo(f, v)
}

// WriteJSON writes a marshaled JSON to a filename with the given Context.
func WriteJSON(ctx context.Context, filename string, v interface{}) error {
	f, err := Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteJSONTo(f, v)
}
