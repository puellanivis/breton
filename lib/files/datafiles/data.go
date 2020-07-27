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
	files.RegisterScheme(handler{}, "data")
}

var b64enc = base64.StdEncoding

type withHeader struct {
	files.Reader
	header http.Header
}

func (w *withHeader) Header() http.Header {
	return w.header
}

func (handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  os.ErrInvalid,
		}
	}

	path := uri.Path
	if path == "" {
		var err error
		path, err = url.PathUnescape(uri.Opaque)
		if err != nil {
			return nil, &os.PathError{
				Op:   "open",
				Path: uri.String(),
				Err:  err,
			}
		}
	}

	i := strings.IndexByte(path, ',')
	if i < 0 {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  os.ErrInvalid,
		}
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

	header := http.Header{
		"Content-Type": []string{contentType},
	}

	if isBase64 {
		b := make([]byte, b64enc.DecodedLen(len(data)))

		n, err := b64enc.Decode(b, data)
		if err != nil {
			return nil, &os.PathError{
				Op:   "decode_base64",
				Path: uri.String(),
				Err:  err,
			}
		}

		data = b[:n]
	}

	return &withHeader{
		Reader: wrapper.NewReaderFromBytes(data, uri, time.Now()),
		header: header,
	}, nil
}
