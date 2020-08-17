// Package cachefiles implements a caching filestore accessable through "cache:opaqueURL".
package cachefiles

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type line struct {
	info os.FileInfo
	data []byte
}

// FileStore is a caching structure that holds copies of the content of files.
type FileStore struct {
	sync.RWMutex

	cache map[string]*line
}

// New returns a new caching FileStore, which can be registered into lib/files
func New() *FileStore {
	return &FileStore{}
}

// Default is the default cache attached to the "cache" Scheme
var Default = New()

func init() {
	files.RegisterScheme(Default, "cache")
}

func (h *FileStore) expire(filename string) {
	h.Lock()
	defer h.Unlock()

	delete(h.cache, filename)
}

func trimScheme(uri *url.URL) string {
	u := *uri
	u.Scheme = ""

	return u.String()
}

// Create implements the files.FileStore Create. At this time, it just returns the files.Create() from the wrapped url.
func (h *FileStore) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	return files.Create(ctx, trimScheme(uri))
}

// Open implements the files.FileStore Open. It returns a buffered copy of the files.Reader returned from reading the uri escaped by the "cache:" scheme. Any access within the next ExpireTime set by the context.Context (5 minutes by default) will return a new copy of an bytes.Reader of the same buffer.
func (h *FileStore) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	h.Lock()
	defer h.Unlock()

	filename := trimScheme(uri)

	ctx, reentrant := isReentrant(ctx)
	if reentrant {
		// We are in a reentrant caching scenario.
		// Continuing will deadlock, so we won‘t even try to cache at all.
		return files.Open(ctx, filename)
	}

	h.RLock()
	f, ok := h.cache[filename]
	h.RUnlock()

	if ok {
		return wrapper.NewReaderWithInfo(bytes.NewReader(f.data), f.info), nil
	}

	// default 5 minute expiration
	expiration := 5 * time.Minute
	if d, ok := GetExpire(ctx); ok {
		expiration = d
	}

	h.Lock()
	defer h.Unlock()

	f, ok = h.cache[filename]
	if ok {
		// We have to test existance again.
		// Maybe another goroutine already did our work.
		return wrapper.NewReaderWithInfo(bytes.NewReader(f.data), f.info), nil
	}

	raw, err := files.Open(ctx, filename)
	if err != nil {
		return nil, err
	}

	info, err := raw.Stat()
	if err != nil {
		info = nil // safety guard
	}

	data, err := files.ReadFrom(raw)
	if err != nil {
		return nil, err
	}

	if info == nil {
		info = wrapper.NewInfo(uri, len(data), time.Now())
	}

	f = &line{
		data: data,
		info: info,
	}

	if h.cache == nil {
		h.cache = make(map[string]*line)
	}

	h.cache[filename] = f

	timer := time.NewTimer(expiration)

	go func() {
		<-timer.C
		h.expire(filename)
	}()

	return wrapper.NewReaderWithInfo(bytes.NewReader(data), info), nil
}

// List implements the files.FileStore List. It does not cache anything and just returns the files.List() from the wrapped url.
func (h *FileStore) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return files.List(ctx, trimScheme(uri))
}
