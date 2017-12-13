// Package json is intended to replace uses of encoding/json while integrating with the files package.
package json

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/puellanivis/breton/lib/files"
)

// Unmarshal is encoding/json.Unmarshal
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

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

// Marshal is a wrapper around encoding/json.Marshal that will optionally apply
// Indent or Compact options.
func Marshal(v interface{}, opts ...Option) ([]byte, error) {
	c := &config{
		escapeHTML: true,
	}

	for _, opt := range opts {
		_ = opt(c)
	}

	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)

	if c.prefix != "" || c.indent != "" {
		enc.SetIndent(c.prefix, c.indent)
	}

	if !c.escapeHTML {
		enc.SetEscapeHTML(c.escapeHTML)
	}

	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	if c.compact {
		buf := new(bytes.Buffer)
		if err := json.Compact(buf, b.Bytes()); err != nil {
			return nil, err
		}
		b = buf
	}

	return b.Bytes(), nil
}

// WriteTo writes a value marshalled as JSON to the the given io.WriteCloser.
func WriteTo(w io.WriteCloser, v interface{}, opts ...Option) error {
	b, err := Marshal(v, opts...)
	if err != nil {
		return err
	}

	return files.WriteTo(w, b)
}

// Write writes a marshaled JSON to a filename with the given Context.
func Write(ctx context.Context, filename string, v interface{}, opts ...Option) error {
	f, err := files.Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteTo(f, v, opts...)
}
