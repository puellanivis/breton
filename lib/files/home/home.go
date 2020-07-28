// Package home implements a URL scheme "home:" which references files according to user home directories.
package home

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"
)

type handler struct{}

func init() {
	files.RegisterScheme(handler{}, "home")
}

func (handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	filename, err := Filename(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	return os.Open(filename)
}

func (handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	filename, err := Filename(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	return os.Create(filename)
}

func (handler) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	filename, err := Filename(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  err,
		}
	}

	return ioutil.ReadDir(filename)
}
