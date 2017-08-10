// Package about implements a simple "about:" scheme.
package aboutfiles

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
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
	return nil, os.ErrInvalid
}

type fn func() ([]byte, error)

func blank() ([]byte, error) {
	return nil, nil
}

func notfound() ([]byte, error) {
	return nil, os.ErrNotExist
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
		"legacy-compat": notfound,
		"plugins":       plugins,
		"srcdoc":        notfound,
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

	var b bytes.Buffer

	for _, item := range list {
		b.WriteString(fmt.Sprintln(item))
	}

	return b.Bytes(), nil
}

func plugins() ([]byte, error) {
	return listOf(files.RegisteredSchemes())
}

func about() ([]byte, error) {
	var list []string

	for name := range aboutMap {
		if name == "" {
			continue
		}

		list = append(list, name)
	}

	return listOf(list)
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	if uri.Opaque == "" {
		filename := uri.String()
		if len(uri.Scheme)+3 < len(filename) {
			uri.Opaque = filename[len(uri.Scheme)+3:]
		}
	}

	f, ok := aboutMap[uri.Opaque]
	if !ok {
		return nil, os.ErrInvalid
	}

	data, err := f()
	if err != nil {
		return nil, err
	}

	return wrapper.NewReader(uri, data, time.Now()), nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	var list []string

	for name := range aboutMap {
		if name == "" || !strings.HasPrefix(name, uri.Opaque) {
			continue
		}

		list = append(list, name)
	}

	sort.Strings(list)

	var ret []os.FileInfo

	for _, name := range list {
		f, _ := aboutMap[name]

		data, err := f()
		if err != nil {
			return nil, err
		}

		ret = append(ret, wrapper.NewInfo(uri, len(data), time.Now()))
	}

	return ret, nil
}
