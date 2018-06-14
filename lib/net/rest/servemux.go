package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
)

type ServeMux struct {
	mu sync.Mutex

	mux map[string]Handler

	def Handler
}

func splitPath(uri *url.URL) []string {
	p := uri.Path
	if p == "" {
		p = uri.Opaque
	}

	p = path.Clean(p)

	return strings.Split(p, "/")
}

func (m *ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	path := splitPath(req.URL)

	r := &Request{
		Path:    path,
		Values:  make(map[string]interface{}),
		Request: req,
	}

	w.Header().Set("Content-Type", "application/json")

	type statuser interface {
		StatusCode() int
	}

	out, err := m.ServeREST(ctx, path, r)
	if err != nil {
		code := http.StatusInternalServerError

		if s, ok := err.(statuser); ok {
			code = s.StatusCode()
		}

		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), code)
		return
	}

	data, err := json.Marshal(out)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), 500)
		return
	}

	w.Write(data)
}

func (m *ServeMux) Handler(pathElem string) Handler {
	m.mu.Lock()
	defer m.mu.Unlock()

	if h, ok := m.mux[pathElem]; ok {
		return h
	}

	return m.def
}

func (m *ServeMux) ServeREST(ctx context.Context, path []string, req *Request) (interface{}, error) {
	if len(path) < 1 {
		if m.def == nil {
			return nil, ErrNotFound
		}

		return m.def.ServeREST(ctx, nil, req)
	}

	head, tail := path[0], path[1:]

	h := m.Handler(head)
	if h == nil {
		return nil, ErrNotFound
	}

	return h.ServeREST(ctx, tail, req)
}
