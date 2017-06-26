package datafiles

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/url"
	"os"
	"time"

	"lib/files"
	"lib/files/wrapper"
)

type Handler struct{}

func init() {
	files.RegisterScheme(&Handler{}, "data")
}

func (h *Handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	return nil, os.ErrInvalid
}

func splitByte(b []byte, sep byte) (fields [][]byte) {
	for {
		i := bytes.IndexByte(b, sep)
		if i < 0 {
			fields = append(fields,b)
			return
		}

		fields = append(fields, b[:i])
		b = b[i+1:]
	}
}

func (h *Handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	q := uri.Opaque
	if q == "" {
		q = uri.String()
		q = q[len(uri.Scheme)+1:]
	}

	data := []byte(q)

	var isBase64 bool

	for _, field := range splitByte(data, ',') {
		data = field

		for _, meta := range splitByte(field, ';') {
			s := string(meta)

			switch s {
			case "base64":
				isBase64 = true
			}
		}
	}

	if isBase64 {
		var err error

		for i, b := range data {
			if b == ' ' {
				data[i] = '+'
			}
		}

		b := make([]byte, len(data))

		n, err := base64.StdEncoding.Decode(b, data)
		if err != nil {
			return nil, err
		}

		data = b[:n]
	}

	return wrapper.NewReader(uri, data, time.Now()), nil
}

func (h *Handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
