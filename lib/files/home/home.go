// Package home implements a URL scheme "home:" which references files according to user home directories.
package home

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/os/user"
)

var userDir string

type handler struct{}

func init() {
	var err error

	// Short-circuit figuring out the whole User, in case we're on Windows.
	userDir, err = user.CurrentHomeDir()
	if err != nil {
		return
	}

	files.RegisterScheme(&handler{}, "home")
}

// Filename takes a given url, and returns a filename that is an absolute path
// for the specific default user if home:filename, or a specific user if home://user@/filename.
func Filename(uri *url.URL) (string, error) {
	if uri.Host != "" {
		return "", os.ErrInvalid
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	dir := userDir

	if uri.User != nil {
		u, err := user.Lookup(uri.User.Username())
		if err != nil {
			return "", err
		}

		if u.HomeDir != "" {
			dir = u.HomeDir
		}
	}

	return filepath.Join(dir, path), nil
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	filename, err := Filename(uri)
	if err != nil {
		return nil, err
	}

	return os.Open(filename)
}

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	filename, err := Filename(uri)
	if err != nil {
		return nil, err
	}

	return os.Create(filename)
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	filename, err := Filename(uri)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadDir(filename)
}
