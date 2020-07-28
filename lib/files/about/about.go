// Package aboutfiles implements a simple "about:" scheme.
package aboutfiles

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
	"github.com/puellanivis/breton/lib/os/process"
	"github.com/puellanivis/breton/lib/sort"
)

type handler struct{}

func init() {
	files.RegisterScheme(handler{}, "about")
}

type reader interface {
	ReadAll() ([]byte, error)
}

type lister interface {
	ReadDir() ([]os.FileInfo, error)
}

type stringFunc func() string

func (f stringFunc) ReadAll() ([]byte, error) {
	return append([]byte(f()), '\n'), nil
}

var (
	blank   stringFunc = func() string { return "" }
	version stringFunc = func() string { return process.Version() }
	now     stringFunc = func() string { return time.Now().Truncate(0).String() }
)

type errorURL struct {
	error
}

func (e errorURL) ReadAll() ([]byte, error) {
	return nil, e.error
}

func (e errorURL) ReadDir() ([]os.FileInfo, error) {
	return nil, e.error
}

// ErrNoSuchHost defines an error, where a DNS host lookup failed to resolve.
var ErrNoSuchHost = errors.New("no such host")

var (
	notfound     = errorURL{os.ErrNotExist}
	unresolvable = errorURL{ErrNoSuchHost}
)

type aboutMap map[string]reader

var (
	about = aboutMap{
		"":              version,
		"blank":         blank,
		"cache":         blank,
		"invalid":       notfound,
		"html-kind":     unresolvable,
		"legacy-compat": unresolvable,
		"now":           now,
		"plugins":       plugins,
		"srcdoc":        unresolvable,
		"version":       version,
	}
)

func init() {
	// if aboutMap references about, then about references aboutMap
	// and go errors with "initialization loop"
	about["about"] = about
}

func (m aboutMap) keys() []string {
	var list []string

	for key := range m {
		if key == "" || strings.HasPrefix(key, ".") {
			continue
		}

		list = append(list, key)
	}

	sort.Strings(list)

	return list
}

func (m aboutMap) ReadAll() ([]byte, error) {
	keys := m.keys()

	b := new(bytes.Buffer)

	for _, key := range keys {
		uri := &url.URL{
			Scheme: "about",
			Opaque: url.PathEscape(key),
		}

		fmt.Fprintln(b, uri)
	}

	return b.Bytes(), nil
}

func (m aboutMap) ReadDir() ([]os.FileInfo, error) {
	keys := m.keys()

	var infos []os.FileInfo

	for _, key := range keys {
		f := m[key]

		data, err := f.ReadAll()
		if err != nil {
			// skip errorURL endpoints.
			continue
		}

		uri := &url.URL{
			Path: key,
		}

		info := wrapper.NewInfo(uri, len(data), time.Now())

		if _, ok := f.(lister); ok {
			info.Chmod(info.Mode() | os.ModeDir)
		}

		infos = append(infos, info)
	}

	return infos, nil
}

type schemeList struct{}

func (schemeList) ReadAll() ([]byte, error) {
	schemes := files.RegisteredSchemes()

	b := new(bytes.Buffer)

	for _, scheme := range schemes {
		uri := &url.URL{
			Scheme: scheme,
		}

		fmt.Fprintln(b, uri)
	}

	return b.Bytes(), nil
}

func (schemeList) ReadDir() ([]os.FileInfo, error) {
	schemes := files.RegisteredSchemes()

	var infos []os.FileInfo

	for _, scheme := range schemes {
		uri := &url.URL{
			Path: scheme,
		}

		infos = append(infos, wrapper.NewInfo(uri, 0, time.Now()))
	}

	return infos, nil
}

var (
	plugins schemeList
)

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
			Op:   "read",
			Path: uri.String(),
			Err:  err,
		}
	}

	return wrapper.NewReaderFromBytes(data, uri, time.Now()), nil
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
				Op:   "open",
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

	if f, ok := f.(lister); ok {
		infos, err := f.ReadDir()
		if err != nil {
			return nil, &os.PathError{
				Op:   "readdir",
				Path: uri.String(),
				Err:  err,
			}
		}

		return infos, nil
	}

	return nil, &os.PathError{
		Op:   "readdir",
		Path: uri.String(),
		Err:  files.ErrNotDirectory,
	}
}
