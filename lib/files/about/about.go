// Package aboutfiles implements a simple "about:" scheme.
package aboutfiles

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
	"github.com/puellanivis/breton/lib/sort"
	"github.com/puellanivis/breton/lib/util"
)

type handler struct{}

func init() {
	files.RegisterScheme(&handler{}, "about")
}

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	return nil, &os.PathError{"create", uri.String(), os.ErrInvalid}
}

type fn func() ([]byte, error)

func blank() ([]byte, error) {
	return nil, nil
}

func notfound() ([]byte, error) {
	return nil, os.ErrNotExist
}

var errUnresolvable = errors.New("unresolvable address")

func unresolvable() ([]byte, error) {
	return nil, errUnresolvable
}

func version() ([]byte, error) {
	return append([]byte(util.Version()), '\n'), nil
}

var (
	aboutMap = map[string]fn{
		"":              version,
		"blank":         blank,
		"cache":         blank,
		"invalid":       notfound,
		"html-kind":     unresolvable,
		"legacy-compat": unresolvable,
		"plugins":       plugins,
		"srcdoc":        unresolvable,
		"version":       version,
	}
)

func init() {
	// if aboutMap references about, then about references aboutMap
	// and go errors with "initialization loop"
	aboutMap["about"] = about
}

func listOf(list []string) ([]byte, error) {
	sort.Strings(list)

	b := new(bytes.Buffer)

	for _, item := range list {
		fmt.Fprintln(b, item)
	}

	return b.Bytes(), nil
}

func plugins() ([]byte, error) {
	return listOf(files.RegisteredSchemes())
}

func about() ([]byte, error) {
	var list []string

	for name := range aboutMap {
		uri := &url.URL{
			Scheme: "about",
			Opaque: name,
		}

		list = append(list, uri.String())
	}

	return listOf(list)
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, &os.PathError{"open", uri.String(), os.ErrInvalid}
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	f, ok := aboutMap[path]
	if !ok {
		return nil, &os.PathError{"open", uri.String(), os.ErrNotExist}
	}

	data, err := f()
	if err != nil {
		return nil, &os.PathError{"open", uri.String(), err}
	}

	return wrapper.NewReaderFromBytes(data, uri, time.Now()), nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	if uri.Host != "" || uri.User != nil {
		return nil, &os.PathError{"readdir", uri.String(), os.ErrInvalid}
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	if f, ok := aboutMap[path]; !ok {
		return nil, &os.PathError{"readdir", uri.String(), os.ErrNotExist}

	} else if f != nil {
		if _, err := f(); err != nil {
			return nil, &os.PathError{"readdir", uri.String(), err}
		}
	}

	if path != "about" && path != "" {
		return nil, &os.PathError{"readdir", uri.String(), syscall.ENOTDIR}
	}

	var list []string
	for name := range aboutMap {
		list = append(list, name)
	}
	sort.Strings(list)

	var ret []os.FileInfo

	for _, name := range list {
		f := aboutMap[name]

		uri := &url.URL{
			Scheme: "about",
			Opaque: name,
		}

		if f == nil {
			continue
		}

		data, err := f()
		if err != nil {
			continue
		}

		ret = append(ret, wrapper.NewInfo(uri, len(data), time.Now()))
	}

	return ret, nil
}
