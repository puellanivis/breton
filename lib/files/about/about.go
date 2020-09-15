// Package aboutfiles implements a simple "about:" scheme.
package aboutfiles

import (
	"context"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type reader interface {
	ReadAll() ([]byte, error)
}

type lister interface {
	ReadDir() ([]os.FileInfo, error)
}

type aboutMap map[string]reader

func (m aboutMap) keys() []string {
	var keys []string

	for key := range m {
		if key == "" || strings.HasPrefix(key, ".") {
			continue
		}

		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (m aboutMap) ReadAll() ([]byte, error) {
	var lines []string

	for _, key := range m.keys() {
		uri := &url.URL{
			Scheme: "about",
			Opaque: url.PathEscape(key),
		}

		lines = append(lines, uri.String())
	}

	return []byte(strings.Join(append(lines, ""), "\n")), nil
}

func (m aboutMap) ReadDir() ([]os.FileInfo, error) {
	var infos []os.FileInfo

	for _, key := range m.keys() {
		uri := &url.URL{
			Scheme: "about",
			Opaque: url.PathEscape(key),
		}

		infos = append(infos, wrapper.NewInfo(uri, 0, time.Now()))
	}

	return infos, nil
}

var (
	about = aboutMap{
		"":              version,
		"blank":         blank,
		"cache":         blank,
		"invalid":       errorURL{os.ErrNotExist},
		"html-kind":     errorURL{ErrNoSuchHost},
		"legacy-compat": errorURL{ErrNoSuchHost},
		"now":           now,
		"plugins":       schemeList{},
		"srcdoc":        errorURL{ErrNoSuchHost},
		"version":       version,
	}
)

func init() {
	// During initialization about is not allowed to reference about,
	// else Go errors with "initialization loop".
	about["about"] = about
}

func init() {
	files.RegisterScheme(handler{}, "about")
}

type handler struct{}

func (h handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	return nil, &os.PathError{
		Op:   "create",
		Path: uri.String(),
		Err:  files.ErrNotSupported,
	}
}

func (h handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  files.ErrURLCannotHaveAuthority,
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
				Err:  files.ErrURLInvalid,
			}
		}
	}

	f, ok := about[path]
	if !ok {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  os.ErrNotExist,
		}
	}

	data, err := f.ReadAll()
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	return wrapper.NewReaderFromBytes(data, uri, time.Now()), nil
}

func (h handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return h.ReadDir(ctx, uri)
}

func (h handler) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  files.ErrURLCannotHaveAuthority,
		}
	}

	path := uri.Path
	if path == "" {
		var err error
		path, err = url.PathUnescape(uri.Opaque)
		if err != nil {
			return nil, &os.PathError{
				Op:   "readdir",
				Path: uri.String(),
				Err:  files.ErrURLInvalid,
			}
		}
	}

	if path == "" {
		path = "about"
	}

	f, ok := about[path]
	if !ok {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  os.ErrNotExist,
		}
	}

	l, ok := f.(lister)
	if !ok {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  files.ErrNotDirectory,
		}
	}

	infos, err := l.ReadDir()
	if err != nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  err,
		}
	}

	return infos, nil
}
