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

// ReadFrom will read the whole of io.Reader into memory.
// It will then close the reader, if it implements io.Closer.
// Finally it will unmarshal that data into v as per json.Unmarshal.
func ReadFrom(r io.Reader, v interface{}) error {
	data, err := files.ReadFrom(r)
	if err != nil {
		return err
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

// Marshal is a wrapper around encoding/json.Marshal,
// that will optionally apply Indent, EscapeHTML, or Compact options.
func Marshal(v interface{}, opts ...Option) ([]byte, error) {
	b := new(bytes.Buffer)

	c := &config{
		Encoder: json.NewEncoder(b),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.prefix != "" || c.indent != "" {
		c.SetIndent(c.prefix, c.indent)
	}

	if err := c.Encode(v); err != nil {
		return nil, err
	}

	if c.compact {
		buf := new(bytes.Buffer)
		if err := json.Compact(buf, b.Bytes()); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	return b.Bytes(), nil
}

// WriteTo marshals v as per json.Marshal,
// it then writes that data to the the given io.Writer.
// Finally, it will close it, if it implements io.Closer.
func WriteTo(w io.Writer, v interface{}, opts ...Option) error {
	b, err := Marshal(v, opts...)
	if err != nil {
		return err
	}

	return files.WriteTo(w, b)
}

// Write writes a marshaled JSON output to the given filename.
func Write(ctx context.Context, filename string, v interface{}, opts ...Option) error {
	f, err := files.Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteTo(f, v, opts...)
}
