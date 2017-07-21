package json

import (
	"context"
	"bytes"
	"encoding/json"
	"io"

	"github.com/puellanivis/breton/lib/files"
)

// ReadFrom will ReadAndClose the given io.ReadCloser and unmarshal that data into v as per json.Unmarshal.
func ReadFrom(r io.ReadCloser, v interface{}) error {
	data, err := files.ReadFrom(r)
	if err != nil {
		return err
	}

	if len(data) < 1 {
		v = nil
		return nil
	}

	return json.Unmarshal(data, v)
}

// Read will open a filename with the given context, and Unmarshal that data into v as per json.Unmarshal.
func Read(ctx context.Context, filename string, v interface{}) error {
	f, err := files.Open(ctx, filename)
	if err != nil {
		return err
	}

	return ReadFrom(f, v)
}

// WriteTo writes a value marshalled as JSON to the the given io.WriteCloser.
func WriteTo(w io.WriteCloser, v interface{}, opts ...Option) error {
	c := new(config)

	for _, opt := range opts {
		_ = opt(c)
	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if c.prefix != "" || c.indent != "" {
		var buf bytes.Buffer
		if err := json.Indent(&buf, b, c.prefix, c.indent); err != nil {
			return err
		}
		b = buf.Bytes()
	}

	if c.compact {
		var buf bytes.Buffer
		if err := json.Compact(&buf, b); err != nil {
			return err
		}
		b = buf.Bytes()
	}

	return files.WriteTo(w, b)
}

// Write writes a marshaled JSON to a filename with the given Context.
func Write(ctx context.Context, filename string, v interface{}, opts ...Option) error {
	f, err := files.Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteTo(f, v)
}
