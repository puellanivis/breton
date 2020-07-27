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

// FS is a caching structure that holds copies of the content of files.
type FS struct {
	sync.RWMutex

	cache map[string]*line
}

// New returns a new caching FS, which can be registered into lib/files
func New() *FS {
	return &FS{}
}

// Default is the default cache attached to the "cache" Scheme
var Default = New()

func init() {
	files.RegisterScheme(Default, "cache")
}

func (h *FS) expire(filename string) {
	h.Lock()
	defer h.Unlock()

	delete(h.cache, filename)
}

func resolveReference(uri *url.URL) (string, error) {
	if uri.Host != "" || uri.User != nil {
		return "", files.ErrURLCannotHaveAuthority
	}

	if uri.Path != "" {
		return uri.Path, nil
	}

	path, err := url.PathUnescape(uri.Opaque)
	if err != nil {
		return "", files.ErrURLInvalid
	}

	return path, nil
}

// Create implements files.CreateFS.
// At this time, it just returns the files.Create() from the wrapped url.
func (h *FS) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	filename, err := resolveReference(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	return files.Create(ctx, filename)
}

// Open implements files.FS.
// It returns a buffered copy of the files.Reader returned from reading the uri escaped by the "cache:" scheme.
// Any access within the next ExpireTime set by the context.Context (or 5 minutes by default) will return a new copy of a files.Reader, with the same content.
func (h *FS) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	filename, err := resolveReference(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	ctx, safe := isReentrySafe(ctx)
	if !safe {
		// We are in a rentrant caching scenario.
		// Continuing will deadlock, so we won’t even try to cache at all.
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

	// We have to test existence again.
	// Maybe another thread already did our work.

	if !ok {
		raw, err := files.Open(ctx, filename)
		if err != nil {
			return nil, err
		}

		info, err := raw.Stat()
		if err != nil {
			info = nil // safety guard.
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
	}

	return wrapper.NewReaderWithInfo(bytes.NewReader(f.data), f.info), nil
}

// ReadDir implements files.ReadDirFS.
// It does not cache anything and just returns the files.ReadDir() from the wrapped url.
func (h *FS) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	filename, err := resolveReference(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: uri.String(),
			Err:  err,
		}
	}

	return files.ReadDir(ctx, filename)
}
