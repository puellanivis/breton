// Package datafiles implements the "data:" URL scheme.
package datafiles

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/url"
	"os"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "data")
}

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	return nil, os.ErrInvalid
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, os.ErrInvalid
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
		if p, err := url.PathUnescape(path); err == nil {
			path = p
		}
	}

	data := []byte(path)

	var isBase64 bool

	fields := bytes.SplitN(data, []byte(","), 2)
	if len(fields) < 2 {	
		return nil, os.ErrInvalid
	}

	data = fields[1]
	for _, meta := range bytes.Split(fields[0], []byte(";")) {
		switch string(meta) {
		case "base64":
			isBase64 = true
		}
	}

	if isBase64 {
		b := make([]byte, len(data))

		n, err := base64.StdEncoding.Decode(b, data)
		if err != nil {
			return nil, err
		}

		data = b[:n]
	}

	return wrapper.NewReaderFromBytes(data, uri, time.Now()), nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}
