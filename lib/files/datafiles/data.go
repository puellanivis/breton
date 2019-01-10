// Package datafiles implements the "data:" URL scheme.
package datafiles

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "data")
}

var b64enc = base64.StdEncoding

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	return nil, files.PathError("create", uri.String(), os.ErrInvalid)
}

type withHeaders struct {
	files.Reader
	header http.Header
}

func (w *withHeaders) Header() http.Header {
	return w.header
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, files.PathError("open", uri.String(), os.ErrInvalid)
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
		if p, err := url.PathUnescape(path); err == nil {
			path = p
		}
	}

	i := strings.IndexByte(path, ',')
	if i < 0 {
		return nil, files.PathError("open", uri.String(), os.ErrInvalid)
	}

	contentType, data := path[:i], []byte(path[i+1:])
	var isBase64 bool

	if strings.HasSuffix(contentType, ";base64") {
		contentType = strings.TrimSuffix(contentType, ";base64")
		isBase64 = true
	}

	if contentType == "" {
		contentType = "text/plain;charset=US-ASCII"
	}

	header := make(http.Header)
	header.Set("Content-Type", contentType)

	if isBase64 {
		b := make([]byte, b64enc.DecodedLen(len(data)))

		n, err := b64enc.Decode(b, data)
		if err != nil {
			return nil, files.PathError("decode", uri.String(), err)
		}

		data = b[:n]
	}

	return &withHeaders{
		Reader: wrapper.NewReaderFromBytes(data, uri, time.Now()),
		header: header,
	}, nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, files.PathError("readdir", uri.String(), os.ErrInvalid)
}
