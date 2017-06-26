package home

import (
	"context"
	"net/url"
	"os"
	"strings"

	"lib/files"
	"lib/os/user"
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

	if !strings.HasSuffix(userDir, string(os.PathSeparator)) {
		userDir += string(os.PathSeparator)
	}

	files.RegisterScheme(&handler{}, "home")
}

func HomePath(uri *url.URL) (string, error) {
	if uri.User != nil {
		u, err := user.Lookup(uri.User.Username())
		if err != nil {
			return "", err
		}

		if dir := u.HomeDir; dir != "" {
			return dir + uri.Path, nil
		}
	}

        if uri.Opaque == "" {
                filename := uri.String()
                if len(uri.Scheme) + 3 < len(filename) {
                        uri.Opaque = filename[len(uri.Scheme)+3:]
                }
        }

	return userDir + uri.Opaque, nil
}

func (_ *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	filename, err := HomePath(uri)
	if err != nil {
		return nil, err
	}

	return os.Open(filename)
}

func (_ *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) { 
	filename, err := HomePath(uri)
	if err != nil {
		return nil, err
	}

	return os.Create(filename)
}

func (_ *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	filename, err := HomePath(uri)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}
